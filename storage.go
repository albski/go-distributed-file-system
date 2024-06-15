package main

import (
	"errors"
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

func (s *Storage) Read(key string) (int64, io.Reader, error) {
	return s.readStream(key)
}

func (s *Storage) Write(key string, r io.Reader) (size int64, err error) {
	return s.writeStream(key, r)
}

func (s *Storage) readStream(key string) (int64, io.ReadCloser, error) {
	keyPath := s.transformPathFunc(key)

	f, err := os.Open(keyPath.fullPath(s.rootDir))
	if err != nil {
		return 0, nil, err
	}

	fi, err := f.Stat()
	if err != nil {
		return 0, nil, err
	}

	return fi.Size(), f, nil
}

func (s *Storage) writeStream(key string, r io.Reader) (int64, error) {
	keyPath := s.transformPathFunc(key)

	if err := os.MkdirAll(keyPath.dirPath(s.rootDir), os.ModePerm); err != nil {
		return 0, err
	}

	fullPath := keyPath.fullPath(s.rootDir)

	f, err := os.Create(fullPath)
	if err != nil {
		return 0, err
	}

	n, err := io.Copy(f, r)
	if err != nil {
		return 0, err
	}

	return n, nil
}

func (s *Storage) Has(key string) bool {
	keyPath := s.transformPathFunc(key)

	_, err := os.Stat(keyPath.fullPath(s.rootDir))

	return !errors.Is(err, os.ErrNotExist)
}

func (s *Storage) Delete(key string) error {
	keyPath := s.transformPathFunc(key)

	defer func() {
		log.Printf("deleted %s", keyPath.Key)
	}()

	return os.RemoveAll(keyPath.rootPath(s.rootDir))
}

func (s *Storage) Clear() error {
	return os.RemoveAll(s.rootDir)
}
