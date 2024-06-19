package main

import (
	"crypto/sha1"
	"encoding/hex"
	"path/filepath"
	"strings"
)

const defaultRootDir = "network"

type KeyPath struct {
	Key  string
	Path string
}

func (k KeyPath) dirPath(basePath string) string {
	return k.joinPaths(basePath, k.Path)
}

func (k KeyPath) fullPath(basePath string) string {
	return k.joinPaths(basePath, k.Path, k.Key)
}

func (k KeyPath) rootPath(basePath string) string {
	fullPath := k.joinPaths("", k.Path, k.Key)
	rootPath := strings.Split(filepath.ToSlash(fullPath), "/")[0]
	return k.joinPaths(basePath, rootPath)
}

func (k KeyPath) joinPaths(basePath string, paths ...string) string {
	if basePath == "" {
		return filepath.Join(paths...)
	}
	basePath = filepath.Clean(basePath)
	return filepath.Join(append([]string{basePath}, paths...)...)
}

type transformPathFunc func(string) KeyPath

func transformPathCrypto(key string) KeyPath {
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
