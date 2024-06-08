package p2p

import (
	"fmt"
	"net"
)

type TCPPeer struct {
	conn net.Conn

	// dial and retrieve => outbound = true
	// accept and retrieve => outbound = false
	outbound bool
}

func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{
		conn:     conn,
		outbound: outbound,
	}
}

type TCPTransportOpts struct {
	ListenAddr    string
	HandshakeFunc HandshakeFunc
	Decoder       Decoder
}

type TCPTransport struct {
	TCPTransportOpts
	listener net.Listener

	// mu    sync.RWMutex
	// peers map[net.Addr]Peer
}

func NewTCPTransport(opts TCPTransportOpts) *TCPTransport {
	return &TCPTransport{TCPTransportOpts: opts}
}

func (t *TCPTransport) ListenAndAccept() error {
	l, err := net.Listen("tcp", t.ListenAddr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", t.ListenAddr, err)
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

	if err := t.HandshakeFunc(peer); err != nil {
		conn.Close()

		fmt.Printf("TCP handshake error: %s\n", err)
		return
	}

	msg := struct{}{} // placeholder
	for {
		if err := t.Decoder.Decode(conn, msg); err != nil {
			fmt.Printf("TCP error: %s\n", err)
			continue
		}
	}
}
