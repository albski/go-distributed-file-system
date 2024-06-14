package main

import (
	"bytes"
	"fmt"
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

	for i := 0; i < 10; i++ {
		data := bytes.NewReader([]byte("1234567890"))
		fs2.Store(fmt.Sprintf("key_%d", i), data)
		time.Sleep(time.Millisecond * 10)
	}

	// r, err := fs2.Get("keykey")
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// b, err := io.ReadAll(r)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// fmt.Println(string(b))

	select {}
}
