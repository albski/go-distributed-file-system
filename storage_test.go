package main

import (
	"bytes"
	"fmt"
	"testing"
)

func TestCryptPathTransformFunc(t *testing.T) {
	key := "randomKey"
	path := CryptPathTransformFunc(key)

	expectedPath := `1d7db/dcda1/992ee/24e72/32d2f/cbe8d/49f28/ca22c`
	if path != expectedPath {
		t.Errorf("have %s\nwant %s", path, expectedPath)
	}
	fmt.Println(path)
}

func TestStorage(t *testing.T) {
	opts := StorageOpts{
		PathTransformFunc: CryptPathTransformFunc,
	}

	s := NewStorage(opts)

	data := bytes.NewReader([]byte("some jpg data"))
	if err := s.writeStream("img", data); err != nil {
		t.Error(err)
	}
}
