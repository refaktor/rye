package util

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"io"
	"os"

	"golang.org/x/crypto/pbkdf2"
)

func SaveSecure(textToEncrypt string, filename string, password string) {
	// Example usage
	// password := "your-password"
	// textToEncrypt := "Hello, World!"

	// Encrypt
	encryptedData, err := encrypt([]byte(textToEncrypt), password)
	if err != nil {
		panic(err)
	}

	// Write encrypted data to a file
	err = os.WriteFile(filename, encryptedData, 0600)
	if err != nil {
		panic(err)
	}
}

func ReadSecure(filename string, password string) string {
	// Read encrypted data from the file
	data, err := os.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	// Decrypt
	decryptedData, err := decrypt(data, password)
	if err != nil {
		panic(err)
	}

	return string(decryptedData)
}

func encrypt(data []byte, passphrase string) ([]byte, error) {
	key := createPBKDF2Key(passphrase)
	block, _ := aes.NewCipher(key)
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return ciphertext, nil
}

func decrypt(data []byte, passphrase string) ([]byte, error) {
	key := createPBKDF2Key(passphrase)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonceSize := gcm.NonceSize()
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}
	return plaintext, nil
}

func createPBKDF2Key(passphrase string) []byte {
	salt := []byte("your-unique-salt") // Use a unique and random salt.
	return pbkdf2.Key([]byte(passphrase), salt, 4096, 32, sha256.New)
}
