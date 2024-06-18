package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"

	"github.com/albski/go-distributed-file-system/p2p"
)

type Message struct {
	Payload any
}

type MessageStoreFile struct {
	ID   string
	Key  string
	Size int64
}

type MessageGetFile struct {
	ID  string
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

	n, err := fs.storage.Write(fs.id, m.Key, io.LimitReader(peer, m.Size))
	if err != nil {
		return err
	}

	log.Printf("%s written %d bytes to disk\n", fs.transport.Addr(), n)

	peer.CloseStream()

	return nil
}

func (fs *FileServer) handleMessageGetFile(from string, m MessageGetFile) error {
	if !fs.storage.Has(fs.id, m.Key) {
		return fmt.Errorf("%s need to serve %s but it doesnt exist on disk", fs.transport.Addr(), m.Key)
	}

	fmt.Printf("%s serving file %s\n", fs.transport.Addr(), m.Key)
	fileSize, r, err := fs.storage.Read(fs.id, m.Key)
	if err != nil {
		return err
	}

	rc, ok := r.(io.ReadCloser)
	if ok {
		fmt.Println("closing ReadCloser")
		defer rc.Close()
	}

	peer, ok := fs.peers[from]
	if !ok {
		return fmt.Errorf("peer %s not in map", from)
	}

	peer.Send([]byte{p2p.StreamRPC})
	binary.Write(peer, binary.LittleEndian, fileSize)

	n, err := io.Copy(peer, r)
	if err != nil {
		return err
	}

	fmt.Printf("%s written %d bytes over the network to %s\n", fs.transport.Addr(), n, from)

	return nil
}
