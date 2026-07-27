package security

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"regexp"

	"golang.org/x/crypto/chacha20poly1305"
)

var databaseKeyPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{32}$`)

// ValidateDatabaseEncryptionKey enforces the deployment configuration contract.
func ValidateDatabaseEncryptionKey(key string) error {
	if !databaseKeyPattern.MatchString(key) {
		return fmt.Errorf("database.encryption_key must contain exactly 32 ASCII characters from A-Z, a-z, 0-9, _ and -")
	}
	return nil
}

// Encrypt returns a self-contained, unpadded Base64URL XChaCha20-Poly1305 payload.
func Encrypt(plaintext []byte, databaseEncryptionKey string) (string, error) {
	if len(plaintext) == 0 {
		return "", fmt.Errorf("plaintext must not be empty")
	}
	if err := ValidateDatabaseEncryptionKey(databaseEncryptionKey); err != nil {
		return "", err
	}
	aead, err := chacha20poly1305.NewX([]byte(databaseEncryptionKey))
	if err != nil {
		return "", fmt.Errorf("initialize secretbox: %w", err)
	}
	nonce := make([]byte, chacha20poly1305.NonceSizeX)
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("generate secretbox nonce: %w", err)
	}
	payload := append(nonce, aead.Seal(nil, nonce, plaintext, nil)...)
	return base64.RawURLEncoding.EncodeToString(payload), nil
}

// Decrypt accepts only the canonical unpadded Base64URL representation.
func Decrypt(encryptedValue string, databaseEncryptionKey string) ([]byte, error) {
	if err := ValidateDatabaseEncryptionKey(databaseEncryptionKey); err != nil {
		return nil, err
	}
	if encryptedValue == "" || len(encryptedValue) < 2 || encryptedValue[len(encryptedValue)-1] == '=' {
		return nil, fmt.Errorf("encrypted value is not canonical Base64URL")
	}
	payload, err := base64.RawURLEncoding.Strict().DecodeString(encryptedValue)
	if err != nil || base64.RawURLEncoding.EncodeToString(payload) != encryptedValue {
		return nil, fmt.Errorf("encrypted value is not canonical Base64URL")
	}
	minimum := chacha20poly1305.NonceSizeX + chacha20poly1305.Overhead
	if len(payload) <= minimum {
		return nil, fmt.Errorf("encrypted value is truncated or empty")
	}
	aead, err := chacha20poly1305.NewX([]byte(databaseEncryptionKey))
	if err != nil {
		return nil, fmt.Errorf("initialize secretbox: %w", err)
	}
	plaintext, err := aead.Open(nil, payload[:chacha20poly1305.NonceSizeX], payload[chacha20poly1305.NonceSizeX:], nil)
	if err != nil {
		return nil, fmt.Errorf("decrypt encrypted value: authentication failed")
	}
	if len(plaintext) == 0 {
		return nil, fmt.Errorf("decrypted value is empty")
	}
	return plaintext, nil
}
