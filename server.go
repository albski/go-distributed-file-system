package main

import (
	"fmt"

	"github.com/albski/go-distributed-file-system/p2p"
)

type FileServerOpts struct {
	storageRoot       string
	transformPathFunc transformPathFunc
	transport         p2p.Transport
}

type FileServer struct {
	FileServerOpts

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
	}
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

func (fs *FileServer) Start() error {
	if err := fs.transport.ListenAndAccept(); err != nil {
		return err
	}

	fs.loop()

	return nil
}

func (fs *FileServer) Stop() {
	close(fs.quitCh)
}
