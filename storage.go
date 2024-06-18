package main

import (
	"errors"
	"io"
	"log"
	"os"
	"path/filepath"
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

func (s *Storage) Read(id, key string) (int64, io.Reader, error) {
	return s.readStream(id, key)
}

func (s *Storage) Has(id, key string) bool {
	keyPath := s.transformPathFunc(key)
	basePath := filepath.Join(s.rootDir, id)

	_, err := os.Stat(keyPath.fullPath(basePath))

	return !errors.Is(err, os.ErrNotExist)
}

func (s *Storage) Delete(id, key string) error {
	keyPath := s.transformPathFunc(key)
	basePath := filepath.Join(s.rootDir, id)

	defer func() {
		log.Printf("deleted %s", keyPath.Key)
	}()

	return os.RemoveAll(keyPath.rootPath(basePath))
}

func (s *Storage) Clear() error {
	return os.RemoveAll(s.rootDir)
}

func (s *Storage) Write(id, key string, r io.Reader) (size int64, err error) {
	return s.writeStream(id, key, r)
}

func (s *Storage) WriteDecrypt(encryptionKey []byte, id, key string, r io.Reader) (int64, error) {
	f, err := s.fileWrite(id, key)
	if err != nil {
		return 0, err
	}

	n, err := copyDecrypt(encryptionKey, r, f)
	return int64(n), err
}

func (s *Storage) readStream(id, key string) (int64, io.ReadCloser, error) {
	keyPath := s.transformPathFunc(key)
	basePath := filepath.Join(s.rootDir, id)

	f, err := os.Open(keyPath.fullPath(basePath))
	if err != nil {
		return 0, nil, err
	}

	fi, err := f.Stat()
	if err != nil {
		return 0, nil, err
	}

	return fi.Size(), f, nil
}

func (s *Storage) fileWrite(id, key string) (*os.File, error) {
	keyPath := s.transformPathFunc(key)
	basePath := filepath.Join(s.rootDir, id)

	if err := os.MkdirAll(keyPath.dirPath(basePath), os.ModePerm); err != nil {
		return nil, err
	}

	return os.Create(keyPath.fullPath(basePath))
}

func (s *Storage) writeStream(id, key string, r io.Reader) (int64, error) {
	f, err := s.fileWrite(id, key)
	if err != nil {
		return 0, err
	}

	return io.Copy(f, r)
}
