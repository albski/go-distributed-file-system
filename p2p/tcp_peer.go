package p2p

import (
	"net"
	"sync"
)

type TCPPeer struct {
	net.Conn

	// dial and retrieve => outbound = true
	// accept and retrieve => outbound = false
	outbound bool

	wg *sync.WaitGroup
}

func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{
		Conn:     conn,
		outbound: outbound,
		wg:       &sync.WaitGroup{},
	}
}

func (p *TCPPeer) Send(b []byte) error {
	_, err := p.Conn.Write(b)
	return err
}

func (p *TCPPeer) CloseStream() {
	p.wg.Done()
}
