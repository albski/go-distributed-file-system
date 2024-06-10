package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const defaultRootDir = "network"

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

type KeyPath struct {
	Key  string
	Path string // Path is based on Key
}

func (k KeyPath) dirPath(basePath string) string {
	if basePath == "" {
		return k.Path
	}
	basePath = filepath.Clean(basePath)
	return filepath.Join(basePath, k.Path)
}

func (k KeyPath) fullPath(basePath string) string {
	if basePath == "" {
		return filepath.Join(k.Path, k.Key)
	}

	basePath = filepath.Clean(basePath)
	return filepath.Join(basePath, k.Path, k.Key)
}

func (k KeyPath) rootPath(basePath string) string {
	if basePath == "" {
		return filepath.Join(k.Path, k.Key)
	}

	basePath = filepath.Clean(basePath)
	rootPath := strings.Split(filepath.ToSlash(k.fullPath("")), "/")[0]
	return filepath.Join(basePath, rootPath)
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
