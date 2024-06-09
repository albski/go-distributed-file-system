package main

import (
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"io"
	"log"
	"os"
	"path/filepath"
)

type StorageOpts struct {
	transformPathFunc transformPathFunc
}

type Storage struct {
	StorageOpts
}

func NewStorage(opts StorageOpts) *Storage {
	return &Storage{StorageOpts: opts}
}

type KeyPath struct {
	Key  string
	Path string // Path is based on Key
}

type transformPathFunc func(string) KeyPath

func transformPathCrypt(key string) KeyPath {
	hash := sha1.Sum([]byte(key))
	hashStr := hex.EncodeToString(hash[:])

	blockSize := 5
	sliceLen := len(hashStr) / blockSize
	paths := make([]string, sliceLen)

	for i := 0; i < sliceLen; i++ {
		from, to := i*blockSize, (i*blockSize)+blockSize
		paths[i] = hashStr[from:to]
	}

	return KeyPath{
		Key:  hashStr,
		Path: filepath.Join(paths...),
	}
}

func (s *Storage) writeStream(key string, r io.Reader) error {
	keyPath := s.transformPathFunc(key)

	if err := os.MkdirAll(keyPath.Path, os.ModePerm); err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	io.Copy(buf, r)
	fileNameBytes := md5.Sum(buf.Bytes())
	fileName := hex.EncodeToString(fileNameBytes[:])

	absPath := filepath.Join(keyPath.Path, fileName)
	f, err := os.Create(absPath)
	if err != nil {
		return err
	}

	n, err := io.Copy(f, buf)
	if err != nil {
		return err
	}

	log.Printf("written %d bytes to disk: %s", n, absPath)

	return nil
}
