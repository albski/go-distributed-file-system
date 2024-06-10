package p2p

// remote node
type Peer interface {
	Close() error
}

// anything that handles communication
// between nodes in the network
// e.g. TCP or UDP or websockets
type Transport interface {
	ListenAndAccept() error
	Consume() <-chan RPC
	Close() error
}
