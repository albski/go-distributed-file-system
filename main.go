package main

import (
	"log"

	"github.com/albski/go-distributed-file-system/p2p"
)

func makeServer(listenAddr string, nodes ...string) *FileServer {
	tcpTransportOpts := p2p.TCPTransportOpts{
		ListenAddr:    ":3000",
		HandshakeFunc: p2p.NOPHandshakeFunc,
		Decoder:       p2p.DefaultDecoder{},
	}

	tcpTransport := p2p.NewTCPTransport(tcpTransportOpts)

	fileServerOpts := FileServerOpts{
		storageRoot:       listenAddr + "_network",
		transformPathFunc: transformPathCrypt,
		transport:         tcpTransport,
		bootstrapNodes:    nodes,
	}

	return NewFileServer(fileServerOpts)
}

func main() {
	fs1 := makeServer(":3000", "")
	fs2 := makeServer(":4000", ":3000")

	go func() {
		err := fs1.Start()
		if err != nil {
			log.Fatal(err)
		}
	}()

	go func() {
		err := fs2.Start()
		if err != nil {
			log.Fatal(err)
		}
	}()
}
