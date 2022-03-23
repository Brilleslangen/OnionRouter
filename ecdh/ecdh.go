// Package ecdh provides Elliptic Curve Diffie Hellman encryption-functionality
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

var curve = elliptic.P256()

func ShareSecret(internalKey *ecdsa.PrivateKey, extKeyX big.Int, extKeyY big.Int) []byte {
	a, _ := curve.ScalarMult(&extKeyX, &extKeyY, internalKey.D.Bytes())
	sharedSecret := sha256.Sum256(a.Bytes())
	return sharedSecret[:]
}

func Encrypt(payload []byte, key []byte) ([]byte, error) {
	// Generate a new AES cipher using our 32 byte long key
	c, err := aes.NewCipher(key)
	if err != nil {
		return []byte{}, err
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return []byte{}, err
	}

	nonce := make([]byte, gcm.NonceSize())
	_, err = io.ReadFull(rand.Reader, nonce)
	if err != nil {
		return []byte{}, err
	}

	return gcm.Seal(nonce, nonce, payload, nil), nil
}

func Decrypt(encryptedPayload []byte, key []byte) ([]byte, error) {
	c, err := aes.NewCipher(key)
	if err != nil {
		return []byte{}, err
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return []byte{}, err
	}

	nonceSize := gcm.NonceSize()
	if len(encryptedPayload) < nonceSize {
		return []byte{}, err
	}

	nonce, encryptedPayload := encryptedPayload[:nonceSize], encryptedPayload[nonceSize:]
	payload, err := gcm.Open(nil, nonce, encryptedPayload, nil)
	if err != nil {
		return []byte{}, err
	}

	return payload, nil
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
