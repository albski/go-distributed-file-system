package main

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCopyEncryptDecrypt(t *testing.T) {
	payload := "foo bar"

	key := newEncryptionKey()
	src := bytes.NewReader([]byte(payload))
	dst := new(bytes.Buffer)
	_, err := copyEncrypt(key, src, dst)
	if err != nil {
		t.Error(err)
	}

	out := new(bytes.Buffer)
	_, err = copyDecrypt(key, dst, out)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, payload, out.String())
}
