package main

import (
	"crypto/sha1"
	"encoding/hex"
	"io"
	"log"
	"os"
	"path/filepath"
)

type PathTransformFunc func(string) string

type StorageOpts struct {
	PathTransformFunc PathTransformFunc
}

var DefualtPathTransformFunc = func(key string) string {
	return key
}

func CryptPathTransformFunc(key string) string {
	hash := sha1.Sum([]byte(key))
	hashStr := hex.EncodeToString(hash[:])

	blockSize := 5
	sliceLen := len(hashStr) / blockSize
	paths := make([]string, sliceLen)

	for i := 0; i < sliceLen; i++ {
		from, to := i*blockSize, (i*blockSize)+blockSize
		paths[i] = hashStr[from:to]
	}

	return filepath.Join(paths...)
}

type Storage struct {
	StorageOpts
}

func NewStorage(opts StorageOpts) *Storage {
	return &Storage{StorageOpts: opts}
}

func (s *Storage) writeStream(key string, r io.Reader) error {
	pathName := s.PathTransformFunc(key)

	// figure out sth better than ModePerm
	if err := os.MkdirAll(pathName, os.ModePerm); err != nil {
		return nil
	}

	fileName := "placeholder"

	absPath := filepath.Join(pathName, fileName)
	f, err := os.Create(absPath)
	if err != nil {
		return err
	}

	n, err := io.Copy(f, r)
	if err != nil {
		return err
	}

	log.Printf("written %d bytes to disk: %s", n, absPath)

	return nil
}
