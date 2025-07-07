package main

import (
	"context"
	"crypto/tls"
	"encoding/binary"
	"fmt"
	gnet "net"
	"os"

	"github.com/Heliodex/coputer/wallflower/keys"
	"github.com/Heliodex/coputer/wallflower/net"
	"github.com/quic-go/quic-go"
)

func addrToUdp(addr keys.Address) (udpAddr *gnet.UDPAddr) {
	// Convert keys.Address to net.UDPAddr
	udpAddr = &gnet.UDPAddr{
		IP:   gnet.IP(addr[:]),
		Port: PortCommunication,
	}
	return
}

func addrToReadable(addr keys.Address) (readable string) {
	return gnet.IP(addr[:]).String()
}

type QuicNet struct {
	tr       *quic.Transport
	tlsConf  *tls.Config
	quicConf *quic.Config
	listener *quic.Listener
	streams  map[keys.Address]*quic.SendStream
}

func NewQuicNet(tlsConf *tls.Config, quicConf *quic.Config) (n net.Net, err error) {
	server, err := gnet.ListenUDP("udp6", &gnet.UDPAddr{
		Port: PortCommunication,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start UDP server: %w", err)
	}

	// set environment variable
	if err = os.Setenv("QUIC_GO_DISABLE_RECEIVE_BUFFER_WARNING", "true"); err != nil {
		fmt.Println("Failed to set OS environment variable.")
	}
	if err = os.Setenv("QUIC_GO_LOG_LEVEL", "debug"); err != nil {
		fmt.Println("Failed to set OS environment variable.")
	}

	tr := &quic.Transport{Conn: server}
	ln, err := tr.Listen(tlsConf, quicConf)
	if err != nil {
		return nil, fmt.Errorf("failed to start QUIC server: %w", err)
	}

	return &QuicNet{
		tr:       tr,
		tlsConf:  tlsConf,
		quicConf: quicConf,
		listener: ln,
		streams:  make(map[keys.Address]*quic.SendStream),
	}, nil
}

func sendMsg(stream *quic.SendStream, msg []byte) (ok bool) {
	if len(msg) == 0 {
		return true
	}

	msgl := make([]byte, 4+len(msg))
	binary.BigEndian.PutUint32(msgl[:4], uint32(len(msg)))
	copy(msgl[4:], msg)

	_, err := stream.Write(msgl)
	return err == nil
}

func (n *QuicNet) sendTo(addr keys.Address, msg net.EncryptedMsg) (ok bool) {
	stream, ok := n.streams[addr]
	if !ok {
		fmt.Println("No stream for   ", addrToReadable(addr))
		return // no stream for this address
	}

	fmt.Println("Sending message ", addrToReadable(addr), "(existing)  length", len(msg))

	// send message on existing stream
	if !sendMsg(stream, msg) {
		delete(n.streams, addr) // remove broken stream
		return
	}
	return true
}

func (n *QuicNet) dialStream(addr keys.Address) (err error) {
	fmt.Println("Dialing         ", addrToReadable(addr))
	qc, err := n.tr.DialEarly(context.TODO(), addrToUdp(addr), n.tlsConf, n.quicConf)
	if err != nil {
		return fmt.Errorf("failed to dial QUIC connection: %w", err)
	}

	stream, err := qc.OpenUniStream()
	if err != nil {
		return fmt.Errorf("failed to open QUIC stream: %w", err)
	}

	n.streams[addr] = stream
	return
}

func (n *QuicNet) transportFromSender(s net.Sender) {
mainloop:
	for msg := range s {
		fmt.Println("Received message to send:", len(msg.EncryptedMsg))
		addrs := append([]keys.Address{msg.MainAddr}, msg.AltAddrs...)

		for _, addr := range addrs {
			if n.sendTo(addr, msg.EncryptedMsg) {
				continue mainloop // found and sent message on existing stream
			}
		}

		for _, addr := range addrs {
			if err := n.dialStream(addr); err != nil {
				fmt.Println("Dialing failed  ", addrToReadable(addr), ":", err)
				continue // failed to dial stream
			}
			n.sendTo(addr, msg.EncryptedMsg) // send on newly created stream
		}
	}
}

func readChunks(stream *quic.ReceiveStream, chunkChan chan<- []byte) {
	for {
		b := make([]byte, msgChunk)

		n, err := stream.Read(b)
		if err != nil || n == 0 {
			continue
		}

		chunkChan <- b[:n]
	}
}

func readMsgs(chunkChan <-chan []byte, msgChan chan<- net.EncryptedMsg) {
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
		b = b[il:]        // remaining bytes are for the next message
	}
}

func parseStream(stream *quic.ReceiveStream, r net.Receiver) {
	chunkChan := make(chan []byte)
	go readChunks(stream, chunkChan)
	go readMsgs(chunkChan, r)
}

func (n *QuicNet) serveToReceiver(ln *quic.Listener, r net.Receiver) {
	for {
		qc, err := ln.Accept(context.TODO())
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}

		fmt.Println("Accepted connection from", qc.RemoteAddr())

		stream, err := qc.AcceptUniStream(context.TODO())
		if err != nil {
			fmt.Println("Error accepting stream:", err)
			continue
		}

		go parseStream(stream, r)
	}
}

func (n *QuicNet) AddNode(node *net.Node) {
	go n.transportFromSender(node.SendRaw)
	go n.serveToReceiver(n.listener, node.ReceiveRaw)
}
