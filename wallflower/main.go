package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"time"

	gnet "net"

	"github.com/Heliodex/coputer/wallflower/keys"
	"github.com/Heliodex/coputer/wallflower/net"
)
// Maximum UDP payload size
// The fact that this is almost the encrypted chunk size is purely coincidental
const maxPayloadSize = 1<<16 - 8 - 1

// Execution System communicates on port 2505
// Communication System communicates on port 2506 with peers
// Gateway communicates on port 2507
// client/management applications (future) communicate on port 2508

func gatewayServer() {}

func managementServer() {}

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
		panic(fmt.Sprintf("failed to start UDP server: %v", err))
	}

	fmt.Println("Public key", kp.Pk.Encode())
	fmt.Println(len(ips), "public IP addresses found")
	fmt.Println("Find string", n.FindString())
	fmt.Println("UDP server listening on", server.LocalAddr())

	go gatewayServer()
	go managementServer()

	go func() {
		for {
			time.Sleep(time.Second)
			// send messages to the server
			msg := make([]byte, 0)
			conn, err := gnet.DialUDP("udp6", nil, ua)

			if err != nil {
				fmt.Println("Error dialing UDP:", err)
				continue
			}

			_, err = conn.Write(msg)
			if err != nil {
				fmt.Println("Error writing to UDP:", err)
				continue
			}

			fmt.Println("Sent message to", conn.RemoteAddr())
		}
	}()

	for {
		// read messages from the server
		// var bb []byte

		for {
			b := make([]byte, maxPayloadSize)
			nb, addr, err := server.ReadFromUDP(b)
			if err != nil {
				fmt.Println("Error reading from UDP:", err)
				continue
			}

			fmt.Println("Received", nb, "bytes from", addr)
		}
	}

	// fmt.Println(server, n)
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
