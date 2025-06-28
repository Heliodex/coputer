package main

import (
	"fmt"

	gnet "net"
)

// Execution System communicates on port 2505
// Communication System communicates on port 2506 with peers
// Gateway communicates on port 2507
// client applications (future) communicate on port 2508

func gatewayServer() {}

func clientServer() {}

func main() {
	// const sk = "cosec:0aqouiilz3-ynmmxunwx1-7u6e5xppqa-hmz7q8yd3f-5l92e17yos"
	// skBytes, err := keys.DecodeSK(sk)
	// if err != nil {
	// 	panic("invalid key")
	// }

	// kp, err := keys.KeypairSK(skBytes)
	// if err != nil {
	// 	panic("invalid keypair")
	// }

	// find current IP address
	addrs, err := gnet.InterfaceAddrs()
	if err != nil {
		panic(fmt.Sprintf("failed to get interface addresses: %v", err))
	}

	fmt.Println("Current IP addresses:")
	for _, addr := range addrs {
		ipnet, ok := addr.(*gnet.IPNet); 
		if !ok || ipnet.IP.To4() != nil {
			continue
		}

		ip := ipnet.IP.To16()
		if ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsMulticast() {
			continue
		}

		fmt.Printf("  %s\n", ip)
	}

	// generate local IP address
	lip, err := gnet.ResolveIPAddr("ip6", "::1")
	if err != nil {
		panic(fmt.Sprintf("failed to resolve local IP address: %v", err))
	}

	fmt.Println(lip)
	fmt.Println(lip.IP.To4())

	// start udp server
	server, err := gnet.ListenUDP("udp6", &gnet.UDPAddr{
		IP:   lip.IP.To16(),
		Port: 2506,
	})
	if err != nil {
		panic(fmt.Sprintf("failed to start UDP server: %v", err))
	}

	fmt.Println("UDP server listening on", server.LocalAddr())
	fmt.Println("UDP server listening on", server.RemoteAddr())

	for {
		buf := make([]byte, 1024)
		n, addr, err := server.ReadFromUDP(buf)
		if err != nil {
			fmt.Println("Error reading from UDP:", err)
			continue
		}
		fmt.Printf("Received %d bytes from %s: %s\n", n, addr, buf[:n])
	}

	// net := net.NewTestNet()
	// n := net.NewNode(kp, keys.Address{})

	// go gatewayServer()
	// go clientServer()

	// fmt.Println(server, n)
}
