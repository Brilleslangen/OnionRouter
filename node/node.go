package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	. "github.com/Brilleslangen/OnionRouter/ecdh"
	. "github.com/Brilleslangen/OnionRouter/orstructs"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"os"
)

var nodeKey *ecdsa.PrivateKey
var SharedSecret [32]byte

func main() {

	// Initialize keypair in the node
	nodeKey, _ = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	PORT := os.Args[1]

	// Alert router that this node is active
	jsonDetails, err := json.Marshal(Node{IP: "TBD", Port: PORT, PublicKeyX: nodeKey.X.Text(10), PublicKeyY: nodeKey.Y.Text(10), SharedSecret: *new([32]byte)})
	check(err)
	request, err := http.NewRequest("POST", "http://127.0.0.1:8080/connect", bytes.NewBuffer(jsonDetails))
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")
	client := http.Client{}
	response, err := client.Do(request)

	// Response from the router contains the routers public key
	var keyResponse KeyResponse
	decoder := json.NewDecoder(response.Body)
	err = decoder.Decode(&keyResponse)
	check(err)
	SharedSecret = getSharedSecret(keyResponse)
	fmt.Println(SharedSecret)
	check(err)

	if response.Status == "200 OK" {
		fmt.Println("Connected to router")
	}

	// Start listening for requests
	http.HandleFunc("/", handler)
	err = http.ListenAndServe(":"+PORT, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		fmt.Println("New POST-request from " + r.RemoteAddr)

		//Decrypt payload
		body, err := io.ReadAll(r.Body)
		decryptedBody, _ := decrypt(SharedSecret[:], body)
		check(err)

		// Allocate payload struct, then decode and write response body into it.
		var payload Payload
		err = json.Unmarshal(decryptedBody, &payload)
		check(err)

		// Execute request if last node or send to next node
		var resp *http.Response
		if payload.NextNode == "" {
			resp, err = http.Get(string(payload.Payload))
			check(err)
		} else {
			// Create request
			request, err :=
				http.NewRequest("POST", "http://"+payload.NextNode, bytes.NewBuffer(payload.Payload))
			check(err)
			request.Header.Set("Content-Type", "application/json; charset=UTF-8")

			// Send request and collect response
			client := http.Client{}
			resp, err = client.Do(request)
			check(err)
		}

		// Convert response to JSON
		jsonData, err := json.Marshal(resp.Body)
		check(err)
		fmt.Println(jsonData)

		// Encrypt JSON to bytes
		encryptedResponse, err := encrypt(SharedSecret[:], jsonData)
		check(err)

		// Convert bytes to io.readCloser, the type of resp.Body
		encryptedBody := ioutil.NopCloser(bytes.NewReader(encryptedResponse))
		check(err)

		// Forward response from origin-url to client with encrypted body.
		_, err = io.Copy(w, encryptedBody)
		check(err)
	} else {
		t, _ := template.ParseFiles("../html/blank.html")
		err := t.Execute(w, nil)
		check(err)
		_, _ = fmt.Fprintf(w, "%s", "This is port only accepts POST-methods")
	}
}

func getSharedSecret(response KeyResponse) [32]byte {
	x, _ := new(big.Int).SetString(response.X, 10)
	y, _ := new(big.Int).SetString(response.Y, 10)
	a, _ := nodeKey.PublicKey.Curve.ScalarMult(x, y, nodeKey.D.Bytes())
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
