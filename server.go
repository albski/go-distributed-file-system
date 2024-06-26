package main

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"github.com/albski/go-distributed-file-system/p2p"
)

type FileServerOpts struct {
	id                string
	encryptionKey     []byte
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

	if opts.id == "" {
		opts.id = generateId()
	}

	return &FileServer{
		FileServerOpts: opts,
		storage:        NewStorage(storageOpts),
		quitCh:         make(chan struct{}),
		peers:          make(map[string]p2p.Peer),
	}
}

func (fs *FileServer) Start() error {
	fmt.Printf("%s starting file server\n", fs.transport.Addr())

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

func (fs *FileServer) Delete(key string) error {
	if !fs.storage.Has(fs.id, key) {
		return fmt.Errorf("%s key not in storage: %v", key, fs.transport.Addr())
	}

	m := Message{
		Payload: MessageDeleteFile{
			ID:  fs.id,
			Key: hashKey(key),
		},
	}

	if err := fs.broadcast(&m); err != nil {
		return err
	}

	if err := fs.storage.Delete(fs.id, key); err != nil {
		return err
	}

	return nil
}

func (fs *FileServer) Get(key string) (io.Reader, error) {
	if fs.storage.Has(fs.id, key) {
		fmt.Printf("%s serving file %s from local disk\n", fs.transport.Addr(), key)

		_, r, err := fs.storage.Read(fs.id, key)
		return r, err
	}

	fmt.Printf("%s dont have file %s, fetching from the network\n", fs.transport.Addr(), key)

	m := Message{
		Payload: MessageGetFile{
			ID:  fs.id,
			Key: hashKey(key),
		},
	}

	if err := fs.broadcast(&m); err != nil {
		return nil, err
	}

	for _, peer := range fs.peers {
		var fileSize int64
		binary.Read(peer, binary.LittleEndian, &fileSize)

		n, err := fs.storage.WriteDecrypt(fs.encryptionKey, fs.id, key, io.LimitReader(peer, fileSize))
		if err != nil {
			return nil, err
		}

		fmt.Printf("%s received %d bytes over the network from %s\n", fs.transport.Addr(), n, peer.RemoteAddr())

		peer.CloseStream()
	}

	_, r, err := fs.storage.Read(fs.id, key)
	return r, err
}

func (fs *FileServer) Store(key string, r io.Reader) error {
	fileBuf := new(bytes.Buffer)
	tee := io.TeeReader(r, fileBuf)

	size, err := fs.storage.Write(fs.id, key, tee)
	if err != nil {
		return err
	}

	m := Message{
		Payload: MessageStoreFile{
			ID:   fs.id,
			Key:  hashKey(key),
			Size: size + 16, // size + BlockSize()
		},
	}

	if err := fs.broadcast(&m); err != nil {
		return err
	}

	time.Sleep(time.Millisecond) // store in peers doesnt work without this line

	peers := []io.Writer{}
	for _, peer := range fs.peers {
		peers = append(peers, peer)
	}
	mw := io.MultiWriter(peers...)
	mw.Write([]byte{p2p.StreamRPC})
	n, err := copyEncrypt(fs.encryptionKey, fileBuf, mw)
	if err != nil {
		return err
	}

	fmt.Println("received and written bytes to disk:", n)

	return nil
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
				log.Println("handle message error", err)
			}

		case <-fs.quitCh:
			return
		}
	}
}

func (fs *FileServer) bootstrapNetwork() error {
	for _, addr := range fs.bootstrapNodes {
		if addr == "" {
			continue
		}

		go func(addr string) {
			fmt.Printf("%s attempting to connect to %s\n", fs.transport.Addr(), addr)
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
	gob.Register(MessageDeleteFile{})
}
