package main

import (
	"fmt"
	"io"
	"log"

	"github.com/albski/go-distributed-file-system/p2p"
)

type Message struct {
	Payload any
}

type MessageStoreFile struct {
	Key  string
	Size int64
}

type MessageGetFile struct {
	Key string
}

func (fs *FileServer) handleMessage(from string, m *Message) error {
	switch v := m.Payload.(type) {
	case MessageStoreFile:
		return fs.handleMessageStoreFile(from, v)
	case MessageGetFile:
		return fs.handleMessageGetFile(from, v)
	}

	return nil
}

func (fs *FileServer) handleMessageStoreFile(from string, m MessageStoreFile) error {
	peer, ok := fs.peers[from]
	if !ok {
		return fmt.Errorf("peer (%s) could not be founc in the peer list", from)
	}

	n, err := fs.storage.Write(m.Key, io.LimitReader(peer, m.Size))
	if err != nil {
		return err
	}

	log.Printf("written %d bytes to disk\n", n)

	peer.(*p2p.TCPPeer).Wg.Done()

	return nil
}

func (fs *FileServer) handleMessageGetFile(from string, m MessageGetFile) error {
	if !fs.storage.Has(m.Key) {
		return fmt.Errorf("%s doesnt exist on disk", m.Key)
	}

	fmt.Println("serving file: ", m.Key)
	r, err := fs.storage.Read(m.Key)
	if err != nil {
		return err
	}

	peer, ok := fs.peers[from]
	if !ok {
		return fmt.Errorf("peer %s not in map", from)
	}

	n, err := io.Copy(peer, r)
	if err != nil {
		return err
	}

	fmt.Printf("written %d bytes over the network to %s\n", n, from)

	return nil
}
