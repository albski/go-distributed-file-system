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

func (fs *FileServer) Get(key string) (io.Reader, error) {
	if fs.storage.Has(key) {
		return fs.storage.Read(key)
	}
	fmt.Printf("dont have file %s, fetching from the network\n", key)

	m := Message{
		Payload: MessageGetFile{
			Key: key,
		},
	}

	if err := fs.broadcast(&m); err != nil {
		return nil, err
	}

	for _, peer := range fs.peers {
		fmt.Println("receiveing stream from peer", peer.RemoteAddr())
		fileBuffer := new(bytes.Buffer)
		n, err := io.CopyN(fileBuffer, peer, 10)
		if err != nil {
			return nil, err
		}
		fmt.Println("received bytes over the network: ", n)
		fmt.Println(fileBuffer.String())
	}

	select {}

	return nil, nil
}

func (fs *FileServer) Store(key string, r io.Reader) error {
	bufFile := new(bytes.Buffer)
	tee := io.TeeReader(r, bufFile)

	size, err := fs.storage.Write(key, tee)
	if err != nil {
		return err
	}

	m := Message{
		Payload: MessageStoreFile{
			Key:  key,
			Size: size,
		},
	}

	if err := fs.broadcast(&m); err != nil {
		return err
	}

	time.Sleep(time.Millisecond * 10)

	for _, peer := range fs.peers {
		peer.Send([]byte{p2p.StreamRPC})
		n, err := io.Copy(peer, bufFile)
		if err != nil {
			return err
		}

		fmt.Println("received and written bytes to disk:", n)
	}

	return nil
}

func (fs *FileServer) stream(m *Message) error {
	peers := []io.Writer{}

	for _, peer := range fs.peers {
		peers = append(peers, peer)
	}

	mw := io.MultiWriter(peers...)

	return gob.NewEncoder(mw).Encode(m)
}

func (fs *FileServer) broadcast(m *Message) error {
	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(m); err != nil {
		return err
	}

	for _, peer := range fs.peers {
		peer.Send([]byte{p2p.MessageRPC})

		if err := peer.Send(buf.Bytes()); err != nil {
			return err
		}
	}

	return nil
}

func (fs *FileServer) loop() {
	defer func() {
		log.Println("file server stopped due to error or quit action")
		fs.transport.Close()
	}()

	for {
		select {
		case rpc := <-fs.transport.Consume():
			var m Message

			if err := gob.NewDecoder(bytes.NewReader(rpc.Payload)).Decode(&m); err != nil {
				log.Println("decoding error:", err)
			}

			if err := fs.handleMessage(rpc.From, &m); err != nil {
				fmt.Println("handling message: ", m)
				log.Println("handle message error", err)
			}

		case _, ok := <-fs.quitCh:
			if !ok {
				return
			}
		}
	}
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

func init() {
	gob.Register(MessageStoreFile{})
	gob.Register(MessageGetFile{})
}
