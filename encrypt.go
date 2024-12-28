package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
	"os"
)

var (
	encryptKey []byte
)

func init() {
	key := os.Getenv("CONTEXT_KEY")
	if len(key) != 32 {
		panic("CONTEXT_KEY environment variable must be exactly 32 characters")
	}
	encryptKey = []byte(key)
}

func Encrypt(data []byte) ([]byte, error) {
	block, err := aes.NewCipher(encryptKey)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, 12) // GCM nonce size
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	ciphertext := aesgcm.Seal(nonce, nonce, data, nil)
	return ciphertext, nil
}

func Decrypt(ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(encryptKey)
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := aesgcm.NonceSize()
	if nonceSize > len(ciphertext) {
		return nil, fmt.Errorf("data too short")
	}
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}
