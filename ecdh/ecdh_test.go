package ecdh

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/Brilleslangen/OnionRouter/orstructs"
	"reflect"
	"testing"
)

func TestEncryptionAndDecryption(t *testing.T) {
	// Create arbitrary test instance
	testInstance := orstructs.Payload{Content: []byte("vg.no"), NextNode: "123.123:123123"}
	fmt.Println("INSTANCE PASSED IN: ", testInstance)

	// Marshal it into a JSON
	jsonDetails, _ := json.Marshal(testInstance)
	fmt.Println("MARSHALED: ", string(jsonDetails))

	// Generate keys and calculate shared secret
	key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	a, _ := key.Curve.ScalarMult(key.X, key.Y, key.D.Bytes())
	sharedSecret := sha256.Sum256(a.Bytes())

	// Encrypt JSON
	sEncrypted, _ := Encrypt(jsonDetails, sharedSecret[:])
	fmt.Println("ENCRYPTED: " + string(sEncrypted))

	// Decrypt JSON
	sDecrypted, _ := Decrypt(sEncrypted, sharedSecret[:])
	fmt.Println("DECRYPTED", string(sDecrypted))

	// Unmarshal the JSON into a payload instance
	decryptedInstance := new(orstructs.Payload)
	_ = json.Unmarshal(sDecrypted, decryptedInstance)
	fmt.Println("UNMARSHALED: ", decryptedInstance)

	// Assert that the two are equal
	if !reflect.DeepEqual(&testInstance, decryptedInstance) {
		t.Errorf("Decrypted Instance is: %x, expected: %x", decryptedInstance, testInstance)
	}
}
