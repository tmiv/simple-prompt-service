package main

import (
	"bytes"
	"testing"
)

func TestEncryptDecrypt(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		wantErr bool
	}{
		{
			name:    "basic string",
			input:   []byte("hello world"),
			wantErr: false,
		},
		{
			name:    "empty string",
			input:   []byte(""),
			wantErr: false,
		},
		{
			name:    "binary data",
			input:   []byte{0x00, 0x01, 0x02, 0x03, 0xFF},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test encryption
			encrypted, err := Encrypt(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Encrypt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			// Verify encrypted data is different from input
			if bytes.Equal(encrypted, tt.input) {
				t.Error("Encrypted data matches input data")
			}

			// Test decryption
			decrypted, err := Decrypt(encrypted)
			if (err != nil) != tt.wantErr {
				t.Errorf("Decrypt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Verify decrypted data matches original input
			if !bytes.Equal(decrypted, tt.input) {
				t.Errorf("Decrypt() = %v, want %v", decrypted, tt.input)
			}
		})
	}
}

func TestDecryptInvalidData(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		wantErr bool
	}{
		{
			name:    "too short",
			input:   []byte("too short"),
			wantErr: true,
		},
		{
			name:    "invalid nonce",
			input:   make([]byte, 28), // 12 (nonce) + 16 (minimum AES block)
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Decrypt(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Decrypt() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
