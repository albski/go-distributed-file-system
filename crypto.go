package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
)

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
	for {
		n, err := src.Read(buf)
		if n > 0 {
			stream.XORKeyStream(buf, buf[:n])
			if _, err := dst.Write(buf[:n]); err != nil {
				return 0, err
			}
		}

		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return 0, err
		}
	}

	return 0, nil
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
	for {
		n, err := src.Read(buf)
		if n > 0 {
			stream.XORKeyStream(buf, buf[:n])
			if _, err := dst.Write(buf[:n]); err != nil {
				return 0, err
			}
		}

		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return 0, err
		}
	}

	return 0, nil
}
