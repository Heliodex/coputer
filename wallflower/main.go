package main

import (
	"fmt"
	"os"

	gnet "net"

	"github.com/Heliodex/coputer/wallflower/keys"
	"github.com/Heliodex/coputer/wallflower/net"
)

// Execution System communicates on port 2505
// Communication System communicates on port 2506 with peers
// Gateway communicates on port 2507
// client applications (future) communicate on port 2508

func gatewayServer() {}

func clientServer() {}

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

func main() {
	fmt.Println()

	// read secret key from environment variable
	kp := getKeypair()

	fmt.Println("Public key", kp.Pk.Encode())

	// find current IP address

	ips, err := getPublicIPs()
	if err != nil {
		fmt.Println("Failed to get public IP addresses:", err)
		os.Exit(1)
	}

	fmt.Println(len(ips), "public IP addresses found")

	addrs := make([]keys.Address, len(ips))
	for i, ip := range ips {
		addrs[i] = keys.Address([]byte(ip)) // that or [keys.AddressLen]byte(ip)
		fmt.Println("-", ip)
	}

	// generate local IP address
	lip, err := gnet.ResolveIPAddr("ip6", "::1")
	if err != nil {
		fmt.Println("Failed to resolve local IP address:", err)
		os.Exit(1)
	}

	net := net.NewTestNet()
	n := net.NewNode(kp, addrs...)

	fmt.Println("Find string", n.FindString())

	// start udp server
	server, err := gnet.ListenUDP("udp6", &gnet.UDPAddr{
		IP:   lip.IP,
		Port: 2506,
	})
	if err != nil {
		panic(fmt.Sprintf("failed to start UDP server: %v", err))
	}

	fmt.Println("UDP server listening on", server.LocalAddr())

	for {
		buf := make([]byte, 1024)
		n, addr, err := server.ReadFromUDP(buf)
		if err != nil {
			fmt.Println("Error reading from UDP:", err)
			continue
		}
		fmt.Printf("Received %d bytes from %s: %s\n", n, addr, buf[:n])
	}

	// go gatewayServer()
	// go clientServer()

	// fmt.Println(server, n)
}
