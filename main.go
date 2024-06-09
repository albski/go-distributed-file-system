package main

import (
	"fmt"
	"log"

	"github.com/albski/go-distributed-file-system/p2p"
)

func main() {
	opts := p2p.TCPTransportOpts{
		ListenAddr:    ":3000",
		HandshakeFunc: p2p.NOPHandshakeFunc,
		Decoder:       p2p.DefaultDecoder{},
	}

	tr := p2p.NewTCPTransport(opts)

	go func() {
		for {
			msg := <-tr.Consume()
			fmt.Printf("consumed: %v\n", msg)
		}
	}()

	if err := tr.ListenAndAccept(); err != nil {
		log.Fatal(err)
	}

	select {}
}
