// Package ecdh provides Elliptic Curve Diffie Hellman (ECDH) key exchange, and
// Advanced Encryption Standard (AES) encryption and decryption with Galois/Counter Mode
package ecdh

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"io"
	"math/big"
)

// Curve used for key exchange
var curve = elliptic.P256()

// ShareSecret calculates and returns a hashed shared secret
func ShareSecret(internalKey *ecdsa.PrivateKey, extKeyX big.Int, extKeyY big.Int) []byte {
	a, _ := curve.ScalarMult(&extKeyX, &extKeyY, internalKey.D.Bytes())
	sharedSecret := sha256.Sum256(a.Bytes())
	return sharedSecret[:]
}

// Encrypt encrypts a payload using AES and GCM with the provided key
func Encrypt(payload []byte, key []byte) ([]byte, error) {
	// Generate a new AES cipher (block) using our 32 byte long key
	c, err := aes.NewCipher(key)
	if err != nil {
		return []byte{}, err
	}

	// Wrap block in Galois/Counter Mode
	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return []byte{}, err
	}

	// Nonce (numbers used once) is the initialization vector of the GCM, and is filled with random numbers
	nonce := make([]byte, gcm.NonceSize())
	_, err = io.ReadFull(rand.Reader, nonce)
	if err != nil {
		return []byte{}, err
	}

	// gcm.Seal returns the encrypted payload
	return gcm.Seal(nonce, nonce, payload, nil), nil
}

// Decrypt decrypts a payload
func Decrypt(encryptedPayload []byte, key []byte) ([]byte, error) {
	c, err := aes.NewCipher(key)
	if err != nil {
		return []byte{}, err
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return []byte{}, err
	}

	// Verify length of the payload
	nonceSize := gcm.NonceSize()
	if len(encryptedPayload) < nonceSize {
		return []byte{}, err
	}

	// Trim nonce from the encrypted payload
	nonce, encryptedPayload := encryptedPayload[:nonceSize], encryptedPayload[nonceSize:]

	// gcm.Open decrypts the payload
	payload, err := gcm.Open(nil, nonce, encryptedPayload, nil)
	if err != nil {
		return []byte{}, err
	}

	return payload, nil
}
