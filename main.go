package main

import (
	"bytes"
	"log"
	"time"

	"github.com/albski/go-distributed-file-system/p2p"
)

func makeServer(listenAddr string, nodes ...string) *FileServer {
	tcpTransportOpts := p2p.TCPTransportOpts{
		ListenAddr:    listenAddr,
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

	fs := NewFileServer(fileServerOpts)

	tcpTransport.OnPeer = fs.OnPeer

	return fs
}

func main() {
	fs1 := makeServer(":3000", "")
	fs2 := makeServer(":4000", ":3000")

	time.Sleep(time.Second)

	go func() {
		err := fs1.Start()
		if err != nil {
			log.Fatal(err)
		}
	}()

	time.Sleep(time.Second)

	go func() {
		err := fs2.Start()
		if err != nil {
			log.Fatal(err)
		}
	}()

	time.Sleep(time.Second)

	data := bytes.NewReader([]byte("big data"))

	fs2.StoreData("key", data)

	select {}
}
