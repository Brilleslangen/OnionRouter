package ecdh

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/base64"
	. "github.com/Brilleslangen/OnionRouter/orstructs"
	"math/big"
)

type KeyResponse struct {
	X string `json:"x"`
	Y string `json:"y"`
}

func EstablishSharedSecret(node Node, routerKey *ecdsa.PrivateKey) [32]byte {
	x, _ := new(big.Int).SetString(node.PublicKeyX, 10)
	y, _ := new(big.Int).SetString(node.PublicKeyY, 10)
	a, _ := routerKey.PublicKey.Curve.ScalarMult(x, y, routerKey.D.Bytes())
	sharedSecret := sha256.Sum256(a.Bytes())

	return sharedSecret
}

func Encode(in []byte) string {
	return base64.StdEncoding.EncodeToString(in)
}

func Encrypt(key []byte, in []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	check(err)
	cfb := cipher.NewCFBEncrypter(block, in)
	cipherText := make([]byte, len(in))
	cfb.XORKeyStream(cipherText, in)
	return []byte(Encode(cipherText)), nil
}

func Decode(in []byte) []byte {
	decoded, err := base64.StdEncoding.DecodeString(string(in))
	check(err)
	return decoded
}

func Decrypt(key []byte, in []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	check(err)
	cipherText := Decode(in)
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
