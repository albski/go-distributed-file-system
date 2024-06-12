package p2p

import "net"

// remote node
type Peer interface {
	net.Conn

	Send([]byte) error
}

// anything that handles communication
// between nodes in the network
// e.g. TCP or UDP or websockets
type Transport interface {
	Dial(string) error
	ListenAndAccept() error
	Consume() <-chan RPC
	Close() error
}
