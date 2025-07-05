package main

import (
	"context"
	"crypto/tls"
	"encoding/binary"
	"flag"
	"fmt"
	"os"
	"runtime"
	"time"

	gnet "net"

	"github.com/Heliodex/coputer/wallflower/keys"
	"github.com/Heliodex/coputer/wallflower/net"
	"github.com/quic-go/quic-go"
)

// Execution System communicates on port 2505
// Communication System communicates on port 2506 with peers
// Gateway communicates on port 2507
// client/management applications (future) communicate on port 2508

func gatewayServer() {}

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

func dialPeer(tr *quic.Transport, addr *gnet.UDPAddr, tlsConf *tls.Config, quicConf *quic.Config) {
	qc, err := tr.Dial(context.TODO(), addr, tlsConf, quicConf)
	if err != nil {
		fmt.Println("Error dialing QUIC:", err)
		return
	}
	defer qc.CloseWithError(0, "done")

	stream, err := qc.OpenStream()
	if err != nil {
		fmt.Println("Error opening stream:", err)
		return
	}
	defer stream.Close()

	for {
		time.Sleep(time.Second)
		// send messages to the server

		msg := make([]byte, 1<<20)
		// fill with random data
		for i := range msg {
			msg[i] = byte(i % 256)
		}

		if err = sendMsg(stream, msg); err != nil {
			fmt.Println("Error sending message:", err)
		}
	}
}

func parseStream(stream *quic.Stream) {
	chunkChan, recvChan := make(chan []byte), make(chan []byte)
	go readChunks(stream, chunkChan)
	go readMsgs(chunkChan, recvChan)

	for msg := range recvChan {
		receiveMsg(msg)
	}
}

func sendMsg(stream *quic.Stream, msg []byte) (err error) {
	if len(msg) == 0 {
		return
	}

	msgl := make([]byte, 4+len(msg))
	binary.BigEndian.PutUint32(msgl[:4], uint32(len(msg)))
	copy(msgl[4:], msg)

	_, err = stream.Write(msgl)
	return
}

func receiveMsg(msg []byte) {
	fmt.Println("Received message:", len(msg))
}

func readChunks(stream *quic.Stream, chunkChan chan<- []byte) {
	for {
		b := make([]byte, msgChunk)

		n, err := stream.Read(b)
		if err != nil || n == 0 {
			continue
		}

		chunkChan <- b[:n]
	}
}

func readMsgs(chunkChan <-chan []byte, msgChan chan<- []byte) {
	const minChunkSize = 4 // well not really, as a message can't be just a length and 0 bytes, but whatever

	b := make([]byte, 0, minChunkSize)

	for {
		// get enough chunk to read the size
		for len(b) < minChunkSize {
			b = append(b, <-chunkChan...)
		}

		l := binary.BigEndian.Uint32(b[:4])
		b = b[4:] // remove the length bytes
		if l == 0 {
			continue
		}

		// we are not, I repeat, NOT, preallocating the memory for the message here, in case some yobo decides to send loads of messages with 4GiB length
		il := int(l)
		for len(b) < il { // now time for the real message
			b = append(b, <-chunkChan...)
		}

		msgChan <- b[:il] // send it off!!!!!!!
		b = b[il:]      // remaining bytes are for the next message
	}
}

func communicationServer(ln *quic.Listener) {
	for {
		qc, err := ln.Accept(context.TODO())
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}

		fmt.Println("Accepted connection from", qc.RemoteAddr())

		stream, err := qc.AcceptStream(context.TODO())
		if err != nil {
			fmt.Println("Error accepting stream:", err)
			continue
		}

		go parseStream(stream)
	}
}

func start() {
	// read secret key from environment variable
	kp := getKeypair()

	// find current IP address

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

	addrs := make([]keys.Address, len(ips))
	for i, ip := range ips {
		addrs[i] = keys.Address([]byte(ip)) // that or [keys.AddressLen]byte(ip)
	}

	// generate local IP address
	lip, err := gnet.ResolveIPAddr("ip6", "::1")
	if err != nil {
		fmt.Println("Failed to resolve local IP address:", err)
		os.Exit(1)
	}

	net := net.NewTestNet()
	n := net.NewNode(kp, addrs[0], addrs[1:]...)

	// start udp server
	ua := &gnet.UDPAddr{
		IP:   lip.IP,
		Port: 2506,
	}
	server, err := gnet.ListenUDP("udp6", ua)
	if err != nil {
		fmt.Println("Failed to start UDP server:", err)
		os.Exit(1)
	}

	tlsCert, err := kp.Sk.TLS()
	if err != nil {
		fmt.Println("Failed to create TLS certificate:", err)
		os.Exit(1)
	}

	// make a self-signed TLS certificate
	tlsConf := &tls.Config{
		Certificates:       []tls.Certificate{tlsCert},
		NextProtos:         []string{"quic"},
		MinVersion:         tls.VersionTLS13,
		InsecureSkipVerify: true, // "tls: failed to verify certificate: x509: certificate signed by unknown authority"
		// i've been in TLS hell for long enough today
	}

	quicConf := &quic.Config{
		Versions:  []quic.Version{quic.Version2},
		Allow0RTT: true,
	}

	// set environment variable
	if err = os.Setenv("QUIC_GO_DISABLE_RECEIVE_BUFFER_WARNING", "true"); err != nil {
		fmt.Println("Failed to set OS environment variable.")
	}

	tr := &quic.Transport{Conn: server}
	ln, err := tr.Listen(tlsConf, quicConf)
	if err != nil {
		panic(fmt.Sprintf("failed to start QUIC server: %v", err))
	}

	fmt.Println("Public key", kp.Pk.Encode())
	fmt.Println(len(ips), "public IP addresses found")
	fmt.Println("Find string", n.FindString())
	fmt.Println("Communication system listening on", server.LocalAddr())

	go gatewayServer()
	go managementServer()
	go communicationServer(ln)

	dialPeer(tr, ua, tlsConf, quicConf)
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

		fmt.Println(multiple, threads)

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
		} else {
			fmt.Println("Generating keypair...")
			start := time.Now()
			found := keys.GenerateKeys(threads)

			kp := <-found
			fmt.Println("Keypair generated in", time.Since(start))
			fmt.Println("Public key:", kp.Pk.Encode())
			fmt.Println("Secret key:", kp.Sk.Encode())

			fmt.Println("Share your public key or find string with others to connect to your node.")
			fmt.Println("DO NOT SHARE YOUR SECRET KEY WITH ANYONE!")
		}

	case "start":
		fmt.Println("Starting Wallflower...")
		start()
	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}
