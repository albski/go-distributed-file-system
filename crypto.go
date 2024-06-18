package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
)

func generateId() string {
	buf := make([]byte, 32)
	io.ReadFull(rand.Reader, buf)
	return hex.EncodeToString(buf)
}

func hashKey(key string) string {
	hash := md5.Sum([]byte(key))
	return hex.EncodeToString(hash[:])
}

func newEncryptionKey() []byte {
	containsZero := func(buf []byte) bool {
		for _, b := range buf {
			if b == 0x0 {
				return true
			}
		}
		return false
	}

	for {
		buf := make([]byte, 32)
		io.ReadFull(rand.Reader, buf)
		if !containsZero(buf) {
			return buf
		}
	}
}

func copyDecrypt(key []byte, src io.Reader, dst io.Writer) (int, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return 0, err
	}

	iv := make([]byte, block.BlockSize()) // 16 bytes
	if _, err := src.Read(iv); err != nil {
		return 0, err
	}

	buf := make([]byte, 32*1024) // value from io.copyBuffer()
	stream := cipher.NewCTR(block, iv)
	nw := block.BlockSize()

	for {
		n, err := src.Read(buf)
		if n > 0 {
			stream.XORKeyStream(buf, buf[:n])
			nn, err := dst.Write(buf[:n])
			if err != nil {
				return 0, err
			}
			nw += nn
		}

		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return 0, err
		}
	}

	return nw, nil
}

func copyEncrypt(key []byte, src io.Reader, dst io.Writer) (int, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return 0, err
	}

	iv := make([]byte, block.BlockSize()) // 16 bytes
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return 0, err
	}

	if _, err := dst.Write(iv); err != nil {
		return 0, err
	}

	buf := make([]byte, 32*1024) // value from io.copyBuffer()
	stream := cipher.NewCTR(block, iv)
	nw := block.BlockSize()
	for {
		n, err := src.Read(buf)
		if n > 0 {
			stream.XORKeyStream(buf, buf[:n])
			nn, err := dst.Write(buf[:n])
			if err != nil {
				return 0, err
			}
			nw += nn
		}

		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return 0, err
		}
	}

	return nw, nil
}
