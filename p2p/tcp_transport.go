package p2p

import (
	"fmt"
	"net"
	"sync"
)

// remote node of a TCP established conn
type TCPPeer struct {
	// conn is the underlying connection of the peer
	conn net.Conn

	// if we dial and retrieve a conn => outbound = true
	// if we accept and retrieve a conn => outbound = false
	outbound bool
}

func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{
		conn:     conn,
		outbound: outbound,
	}
}

type TCPTransport struct {
	listenAddr    string
	listener      net.Listener
	handshakeFunc HandshakeFunc
	decoder       Decoder

	mu    sync.RWMutex
	peers map[net.Addr]Peer
}

func NewTCPTransport(listenAddr string) *TCPTransport {
	return &TCPTransport{
		handshakeFunc: NOPHandshakeFunc, // temporary
		listenAddr:    listenAddr,
	}
}

func (t *TCPTransport) ListenAndAccept() error {
	l, err := net.Listen("tcp", t.listenAddr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", t.listenAddr, err)
	}

	t.listener = l

	go t.startAcceptLoop()

	return nil
}

func (t *TCPTransport) startAcceptLoop() {
	for {
		conn, err := t.listener.Accept()
		if err != nil {
			fmt.Printf("TCP accept error %s\n", err)
		}

		fmt.Printf("incoming conn: %+v\n", conn.RemoteAddr())
		go t.handleConn(conn)
	}
}

func (t *TCPTransport) handleConn(conn net.Conn) {
	peer := NewTCPPeer(conn, false)

	if err := t.handshakeFunc(peer); err != nil {
		conn.Close()

		fmt.Printf("TCP handshake error: %s\n", err)
		return
	}

	msg := struct{}{} // placeholder
	for {
		if err := t.decoder.Decode(conn, msg); err != nil {
			fmt.Printf("TCP error: %s\n", err)
			continue
		}
	}
}
