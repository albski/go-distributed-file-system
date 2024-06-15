package main

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStorage(t *testing.T) {
	s := newStorage()
	defer teardown(t, s)

	key := "random key"
	data := []byte("some jpg data")

	if _, err := s.writeStream(key, bytes.NewReader(data)); err != nil {
		t.Error(err)
	}

	assert.True(t, s.Has(key))

	_, r, err := s.Read(key)
	if err != nil {
		t.Error(err)
	}

	b, _ := io.ReadAll(r)
	fmt.Println(string(b))
	fmt.Println(string(data))

	assert.Equal(t, string(b), string(data))

	s.Delete(key)

	assert.False(t, s.Has(key))
}

func TestCryptPathTransformFunc(t *testing.T) {
	key := "randomKey"
	keyPath := transformPathCrypt(key)

	expectedKey := `1d7dbdcda1992ee24e7232d2fcbe8d49f28ca22c`
	expectedPath := `1d7db/dcda1/992ee/24e72/32d2f/cbe8d/49f28/ca22c`

	assert.Equal(t, keyPath.Key, expectedKey)
	assert.Equal(t, keyPath.Path, expectedPath)
}

func newStorage() *Storage {
	opts := StorageOpts{
		transformPathFunc: transformPathCrypt,
	}

	return NewStorage(opts)
}

func teardown(t *testing.T, s *Storage) {
	if err := s.Clear(); err != nil {
		t.Error(err)
	}
}
