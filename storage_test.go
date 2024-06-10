package main

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCryptPathTransformFunc(t *testing.T) {
	key := "randomKey"
	keyPath := transformPathCrypt(key)

	expectedKey := `1d7dbdcda1992ee24e7232d2fcbe8d49f28ca22c`
	expectedPath := `1d7db/dcda1/992ee/24e72/32d2f/cbe8d/49f28/ca22c`

	assert.Equal(t, keyPath.Key, expectedKey)
	assert.Equal(t, keyPath.Path, expectedPath)
}

func TestStorageDeleteKey(t *testing.T) {
	opts := StorageOpts{
		transformPathFunc: transformPathCrypt,
	}

	s := NewStorage(opts)

	key := "random key"
	data := []byte("some jpg data")

	if err := s.writeStream(key, bytes.NewReader(data)); err != nil {
		t.Error(err)
	}

	if err := s.Delete(key); err != nil {
		t.Error(err)
	}
}

func TestStorage(t *testing.T) {
	opts := StorageOpts{
		transformPathFunc: transformPathCrypt,
	}

	s := NewStorage(opts)

	key := "random key"
	data := []byte("some jpg data")

	if err := s.writeStream(key, bytes.NewReader(data)); err != nil {
		t.Error(err)
	}

	r, err := s.Read(key)
	if err != nil {
		t.Error(err)
	}

	b, _ := io.ReadAll(r)
	fmt.Println(string(b))
	fmt.Println(string(data))

	assert.Equal(t, string(b), string(data))
}
