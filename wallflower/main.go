package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	gnet "net"

	. "github.com/Heliodex/coputer/litecode/types"
	"github.com/Heliodex/coputer/wallflower/keys"
	"github.com/Heliodex/coputer/wallflower/net"
	"github.com/quic-go/quic-go"
)

// Execution System communicates on port 2505
// Communication System communicates on port 2506 with peers
// Gateway communicates on port 2507, and hosts on port 2517
// client/management applications (future) communicate on port 2508
const (
	PortExecution = iota + 2505
	PortCommunication
	PortGateway
	PortManagement
)

func gatewayServer(n *net.Node) {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /web/{pk}/{name}", func(w http.ResponseWriter, r *http.Request) {
		pks, name := r.PathValue("pk"), r.PathValue("name")
		pk, err := keys.DecodePKNoPrefix(pks)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to decode public key: %v", err), http.StatusBadRequest)
			return
		}

		bodybytes, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to read request body: %v", err), http.StatusBadRequest)
			return
		}

		args, err := DecodeArgs[WebArgs](bodybytes)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to decode request body: %v", err), http.StatusBadRequest)
			return
		}

		rets, err := n.RunWebProgram(pk, name, args, true)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to run web program: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(rets.Encode())
	})

	fmt.Println("Listening for gateway on port", PortGateway)
	http.ListenAndServe(fmt.Sprintf(":%d", PortGateway), mux)
}

func managementServer() {}

const msgChunk = 2 << 19

// IPv6 supremacy
func getPublicIPs() (ips []gnet.IP, err error) {
	addrs, err := gnet.InterfaceAddrs()
	if err != nil {
		return nil, fmt.Errorf("failed to get interface addresses: %v", err)
	}

	for _, addr := range addrs {
		ipnet, ok := addr.(*gnet.IPNet)
		if !ok {
			continue
		}

		ip := ipnet.IP
		if !ip.IsGlobalUnicast() || ip.IsPrivate() {
			continue
		}

		ips = append(ips, ip.To16())
	}

	return
}

func getKeypair() (kp keys.Keypair) {
	const skEnv = "WALLFLOWER_SK"

	if sk, ok := os.LookupEnv(skEnv); !ok {
		fmt.Printf("Environment variable %s not set.\n", skEnv)
		fmt.Println("If you don't have a secret key, you can generate one with the `genkeys` command.")
		os.Exit(1)
	} else if skBytes, err := keys.DecodeSK(sk); err != nil {
		fmt.Println("Invalid secret key provided.")
		os.Exit(1)
	} else if kp, err = keys.KeypairSK(skBytes); err != nil {
		fmt.Println("Failed to create keypair from secret key:", err)
		os.Exit(1)
	}
	return
}

func getAddrs() (addrs []keys.Address) {
	ips, err := getPublicIPs()
	if err != nil {
		fmt.Println("Failed to get public IP addresses:", err)
		os.Exit(1)
	}

	if len(ips) == 0 {
		fmt.Println("No public IP addresses found.")
		fmt.Println("Make sure you are connected to an IPv6 network.")
		os.Exit(1)
	}

	addrs = make([]keys.Address, len(ips))
	for i, ip := range ips {
		addrs[i] = keys.Address([]byte(ip)) // that or [keys.AddressLen]byte(ip)
	}
	return
}

func getPeers() (peers []*keys.Peer) {
	// open the peers file
	const peersFile = "peers"
	file, err := os.Open(peersFile)
	if err != nil && !os.IsNotExist(err) {
		fmt.Printf("Failed to open peers file %s: %v\n", peersFile, err)
		os.Exit(1)
	} else if os.IsNotExist(err) {
		fmt.Printf("Peers file %s does not exist. No peers will be loaded.\n", peersFile)
		return
	}
	defer file.Close()

	b, err := io.ReadAll(file)
	if err != nil {
		fmt.Printf("Failed to read peers file %s: %v\n", peersFile, err)
		os.Exit(1)
	}

	if len(b) == 0 {
		fmt.Printf("Peers file %s is empty. No peers will be loaded.\n", peersFile)
		return
	}

	for line := range strings.SplitSeq(strings.TrimSpace(string(b)), "\n") {
		peer, err := net.PeerFromFindString(line)
		if err != nil {
			fmt.Printf(`Failed to parse peer from line "%s": %v\n`, line, err)
			continue
		}

		fmt.Println("Found peer", peer.Pk.Encode())
		peers = append(peers, peer)
	}
	return
}

func start() {
	kp := getKeypair()
	addrs := getAddrs()
	peers := getPeers()

	// generate local IP address
	// lip, err := gnet.ResolveIPAddr("ip6", "::1")
	// if err != nil {
	// 	fmt.Println("Failed to resolve local IP address:", err)
	// 	os.Exit(1)
	// }

	// make a self-signed TLS certificate
	tlsCert, err := kp.Sk.TLS()
	if err != nil {
		fmt.Println("Failed to create TLS certificate:", err)
		os.Exit(1)
	}

	tlsConf := &tls.Config{
		Certificates:       []tls.Certificate{tlsCert},
		NextProtos:         []string{"quic"},
		MinVersion:         tls.VersionTLS13,
		InsecureSkipVerify: true, // "tls: failed to verify certificate: x509: certificate signed by unknown authority"
		// i've been in TLS hell for long enough today
	}

	quicConf := &quic.Config{
		Versions:             []quic.Version{quic.Version2},
		Allow0RTT:            true,
		KeepAlivePeriod:      15 * time.Second,
		HandshakeIdleTimeout: time.Second, // just 4 speed
	}

	qnet, err := NewQuicNet(tlsConf, quicConf)
	if err != nil {
		fmt.Println("Failed to create QUIC network:", err)
		os.Exit(1)
	}

	n := net.NewNode(kp, addrs[0], addrs[1:]...)

	for _, peer := range peers {
		n.AddPeer(peer)
	}

	fmt.Println("Public key", kp.Pk.Encode())
	fmt.Println(len(addrs), "public network addresses found")
	for _, addr := range addrs {
		fmt.Println("    ", gnet.IP(addr[:]).String())
	}
	fmt.Println("Find string", n.FindString())
	fmt.Println("Communication system listening on port", PortCommunication)

	qnet.AddNode(n)
	n.Start()
	go gatewayServer(n)
	go managementServer()

	select {}
}

func dev(path string) {
	kp := getKeypair()

	n := net.NewNode(kp, keys.Address{})

	fmt.Println("Public key", kp.Pk.Encode())
	fmt.Println("Find string", n.FindString())
	fmt.Println("Communication system listening on port", PortCommunication)

	n.Start()
	go gatewayServer(n)
	go watchPath(n, path)

	select {}
}

func main() {
	if len(os.Args) <= 1 {
		fmt.Println("Usage: <command>")
		fmt.Println("Available commands: genkeys, start")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "genkeys":
		fmultiple := flag.Bool("m", false, "Generate multiple keypairs")
		fthreads := flag.Int("t", runtime.NumCPU(), "Number of threads to use for key generation")

		flag.CommandLine.Parse(os.Args[2:])
		multiple, threads := *fmultiple, *fthreads

		// get cpu cores
		fmt.Printf("Using %d-threaded key generation.\n", threads)

		if multiple {
			fmt.Println("Generating keypairs...")
			start := time.Now()
			found := keys.GenerateKeys(threads)

			for k := range found {
				fmt.Println("Keypair generated in", time.Since(start))
				start = time.Now()

				fmt.Println("Public key:", k.Pk.Encode())
				fmt.Println("Secret key:", k.Sk.Encode())
			}
			return
		}

		fmt.Println("Generating keypair...")
		start := time.Now()
		found := keys.GenerateKeys(threads)

		kp := <-found
		fmt.Println("Keypair generated in", time.Since(start))
		fmt.Println("Public key:", kp.Pk.Encode())
		fmt.Println("Secret key:", kp.Sk.Encode())

		fmt.Println("Share your public key or find string with others to connect to your node.")
		fmt.Println("DO NOT SHARE YOUR SECRET KEY WITH ANYONE!")

	case "start":
		fmt.Println("Starting Wallflower...")
		start()
	case "dev":
		if len(os.Args) < 3 {
			fmt.Println("Usage: dev <filepath>")
			os.Exit(1)
		}

		path := os.Args[2]

		fmt.Println("Starting Wallflower in development mode...")
		dev(path)
	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}
