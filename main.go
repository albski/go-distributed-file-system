package main

import (
	"bytes"
	"fmt"
	"io"
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
	fs2 := makeServer(":3001", "")
	fs3 := makeServer(":3002", ":3000", ":3001")

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

	time.Sleep(time.Second * 2)

	go func() {
		err := fs3.Start()
		if err != nil {
			log.Fatal(err)
		}
	}()

	time.Sleep(time.Second * 2)

	for i := 0; i < 20; i++ {
		key := fmt.Sprintf("cool_%d.txt", i)
		data := bytes.NewReader([]byte("1234567890"))
		fs3.Store(key, data)

		if err := fs3.storage.Delete(fs3.id, key); err != nil {
			log.Fatal(err)
		}

		r, err := fs3.Get(key)
		if err != nil {
			log.Fatal(err)
		}

		b, err := io.ReadAll(r)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(string(b))
	}
}
