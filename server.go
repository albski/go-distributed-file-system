package main

import (
	"fmt"
	"log"
	"sync"

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

func (fs *FileServer) OnPeer(p p2p.Peer) error {
	return nil // todo implementation
}

func (fs *FileServer) loop() {
	for {
		select {
		case msg := <-fs.transport.Consume():
			fmt.Println(msg)
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

func (fs *FileServer) Start() error {
	defer func() {
		fs.transport.Close()
	}()

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
