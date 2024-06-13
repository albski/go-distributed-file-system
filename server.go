package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"github.com/albski/go-distributed-file-system/p2p"
)

type FileServerOpts struct {
	storageRoot       string
	transformPathFunc transformPathFunc
	transport         p2p.Transport

	bootstrapNodes []string
}

type FileServer struct {
	FileServerOpts

	peerLock sync.RWMutex
	peers    map[string]p2p.Peer

	storage *Storage
	quitCh  chan struct{}
}

func NewFileServer(opts FileServerOpts) *FileServer {
	storageOpts := StorageOpts{
		rootDir:           opts.storageRoot,
		transformPathFunc: opts.transformPathFunc,
	}

	return &FileServer{
		FileServerOpts: opts,
		storage:        NewStorage(storageOpts),
		quitCh:         make(chan struct{}),
		peers:          make(map[string]p2p.Peer),
	}
}

func (fs *FileServer) Start() error {
	if err := fs.transport.ListenAndAccept(); err != nil {
		return err
	}

	fs.bootstrapNetwork()

	fs.loop()

	return nil
}

func (fs *FileServer) Stop() {
	close(fs.quitCh)
}

func (fs *FileServer) OnPeer(p p2p.Peer) error {
	fs.peerLock.Lock()
	defer fs.peerLock.Unlock()

	fs.peers[p.RemoteAddr().String()] = p

	log.Printf("connected with remote %s", p.RemoteAddr())
	return nil
}

func (fs *FileServer) bootstrapNetwork() error {
	for _, addr := range fs.bootstrapNodes {
		if addr == "" {
			continue
		}

		go func(addr string) {
			if err := fs.transport.Dial(addr); err != nil {
				log.Println("dial error: ", err)
			}
		}(addr)
	}

	return nil
}

func (fs *FileServer) StoreData(key string, r io.Reader) error {
	buf := new(bytes.Buffer)
	m := Message{
		Payload: MessageStoreFile{
			Key: key,
		},
	}

	if err := gob.NewEncoder(buf).Encode(m); err != nil {
		return err
	}

	for _, peer := range fs.peers {
		fmt.Println(peer, m)
		if err := peer.Send(buf.Bytes()); err != nil {
			return err
		}
	}

	time.Sleep(time.Second * 3)
	payload := []byte("large file")
	for _, peer := range fs.peers {
		if err := peer.Send(payload); err != nil {
			return err
		}
	}

	return nil
}

type Message struct {
	Payload any
}

type MessageStoreFile struct {
	Key string
}

func (fs *FileServer) broadcast(m *Message) error {
	peers := []io.Writer{}

	for _, peer := range fs.peers {
		peers = append(peers, peer)
	}

	mw := io.MultiWriter(peers...)

	return gob.NewEncoder(mw).Encode(m)
}

func (fs *FileServer) loop() {
	defer fs.transport.Close()

	for {
		select {
		case rpc := <-fs.transport.Consume():
			var m Message

			if err := gob.NewDecoder(bytes.NewReader(rpc.Payload)).Decode(&m); err != nil {
				log.Println(err)
				return
			}

			fmt.Println("payload: ", m.Payload)

			peer, ok := fs.peers[rpc.From]
			if !ok {
				panic("peer not found")
			}

			b := make([]byte, 1000)
			if _, err := peer.Read(b); err != nil {
				panic(err)
			}

			fmt.Printf("received: %s\n", string(b))

			peer.(*p2p.TCPPeer).Wg.Done() // temporary

		case _, ok := <-fs.quitCh:
			if !ok {
				return
			}
		}
	}
}

func init() {
	gob.Register(MessageStoreFile{})
	gob.Register(Message{})
}
