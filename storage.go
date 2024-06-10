package main

import (
	"bytes"
	"io"
	"log"
	"os"
)

type StorageOpts struct {
	rootDir           string
	transformPathFunc transformPathFunc
}

type Storage struct {
	StorageOpts
}

func NewStorage(opts StorageOpts) *Storage {
	if opts.rootDir == "" {
		opts.rootDir = defaultRootDir
	}

	return &Storage{StorageOpts: opts}
}

func (s *Storage) Delete(key string) error {
	keyPath := s.transformPathFunc(key)

	defer func() {
		log.Printf("deleted %s", keyPath.Key)
	}()

	return os.RemoveAll(keyPath.rootPath(s.rootDir))
}

func (s *Storage) Read(key string) (io.Reader, error) {
	f, err := s.readStream(key)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, f)

	return buf, err
}

func (s *Storage) readStream(key string) (io.ReadCloser, error) {
	keyPath := s.transformPathFunc(key)
	return os.Open(keyPath.fullPath(s.rootDir))
}

func (s *Storage) writeStream(key string, r io.Reader) error {
	keyPath := s.transformPathFunc(key)

	if err := os.MkdirAll(keyPath.dirPath(s.rootDir), os.ModePerm); err != nil {
		return err
	}

	fullPath := keyPath.fullPath(s.rootDir)

	f, err := os.Create(fullPath)
	if err != nil {
		return err
	}

	n, err := io.Copy(f, r)
	if err != nil {
		return err
	}

	log.Printf("written %d bytes to disk: %s", n, fullPath)

	return nil
}
