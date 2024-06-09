package main

import (
	"bytes"
	"testing"
)

func TestCryptPathTransformFunc(t *testing.T) {
	key := "randomKey"
	keyPath := CryptPathTransformFunc(key)

	expectedKey := `1d7dbdcda1992ee24e7232d2fcbe8d49f28ca22c`
	expectedPath := `1d7db/dcda1/992ee/24e72/32d2f/cbe8d/49f28/ca22c`
	if keyPath.Key != expectedKey {
		t.Errorf("have %s\nwant %s", keyPath.Key, expectedKey)
	}
	if keyPath.Path != expectedPath {
		t.Errorf("have %s\nwant %s", keyPath.Path, expectedPath)
	}
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
