package main

import (
	"log"

	"github.com/albski/go-distributed-file-system/p2p"
)

func main() {
	opts := p2p.TCPTransportOpts{
		ListenAddr:    ":3000",
		HandshakeFunc: p2p.NOPHandshakeFunc,
		Decoder:       p2p.GOBDecoder{},
	}

	tr := p2p.NewTCPTransport(opts)

	if err := tr.ListenAndAccept(); err != nil {
		log.Fatal(err)
	}

	select {}
}
