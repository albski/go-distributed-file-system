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
		encryptionKey:     newEncryptionKey(),
		storageRoot:       listenAddr + "_network",
		transformPathFunc: transformPathCrypto,
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

	data := bytes.NewReader([]byte("1234567890"))
	fs2.Store("cool.txt", data)

	// r, err := fs2.Get("cool.txt")
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// b, err := io.ReadAll(r)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// fmt.Println(string(b))
}
