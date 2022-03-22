package ecdh

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/base64"
	"math/big"
)

type KeyResponse struct {
	X string `json:"x"`
	Y string `json:"y"`
}

func establishSharedSecret(node Node) [32]byte {
	x, _ := new(big.Int).SetString(node.PublicKeyX, 10)
	y, _ := new(big.Int).SetString(node.PublicKeyY, 10)
	a, _ := routerKey.PublicKey.Curve.ScalarMult(x, y, routerKey.D.Bytes())
	sharedSecret := sha256.Sum256(a.Bytes())

	return sharedSecret
}

func encode(in []byte) string {
	return base64.StdEncoding.EncodeToString(in)
}

func encrypt(key []byte, in []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	check(err)
	cfb := cipher.NewCFBEncrypter(block, in)
	cipherText := make([]byte, len(in))
	cfb.XORKeyStream(cipherText, in)
	return []byte(encode(cipherText)), nil
}

func decode(in []byte) []byte {
	decoded, err := base64.StdEncoding.DecodeString(string(in))
	check(err)
	return decoded
}

func decrypt(key []byte, in []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	check(err)
	cipherText := decode(in)
	cfb := cipher.NewCFBEncrypter(block, in)
	text := make([]byte, len(cipherText))
	cfb.XORKeyStream(text, cipherText)
	return text, nil
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
