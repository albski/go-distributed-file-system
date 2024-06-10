package main

import (
	"log"
	"time"

	"github.com/albski/go-distributed-file-system/p2p"
)

func main() {
	tcpTransportOpts := p2p.TCPTransportOpts{
		ListenAddr:    ":3000",
		HandshakeFunc: p2p.NOPHandshakeFunc,
		Decoder:       p2p.DefaultDecoder{},
	}

	tcpTransport := p2p.NewTCPTransport(tcpTransportOpts)

	fileServerOpts := FileServerOpts{
		storageRoot:       "3000_network",
		transformPathFunc: transformPathCrypt,
		transport:         tcpTransport,
	}

	fs := NewFileServer(fileServerOpts)

	go func() {
		time.Sleep(time.Second * 3)
		fs.Stop()
	}()

	if err := fs.Start(); err != nil {
		log.Fatal(err)
	}

}
