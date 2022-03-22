package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"math/big"
	"net/http"
)

type Payload struct {
	NextNode string
	Payload  []byte
}

type NodeDetails struct {
	IP           string
	Port         string
	PublicKeyX   big.Int
	PublicKeyY   big.Int
	SharedSecret [32]byte
}

type KeyResponse struct {
	x big.Int
	y big.Int
}

func main() {
	PORT := "8088"
	pubKey, privKey := generateKey()
	dummyArr := new([32]byte)
	fmt.Println(pubKey)
	// Alert router that this node is active
	jsonDetails, err := json.Marshal(NodeDetails{"TBD", PORT, *pubKey.X, *pubKey.Y, *dummyArr})
	check(err)
	request, err := http.NewRequest("POST", "http://127.0.0.1:8080/connect", bytes.NewBuffer(jsonDetails))
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")
	client := http.Client{}
	response, err := client.Do(request)
	var keyResponse KeyResponse
	decoder := json.NewDecoder(request.Response.Body)
	err = decoder.Decode(&keyResponse)
	sharedSecret := getSharedSecret(keyResponse, privKey)
	fmt.Println(sharedSecret)
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

		// Allocate payload struct, then decode and write response body into it.
		var payload Payload
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&payload)
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

		// Forward response from origin-url to client.
		_, err = io.Copy(w, resp.Body)
		check(err)
	} else {
		t, _ := template.ParseFiles("../html/blank.html")
		err := t.Execute(w, nil)
		check(err)
		_, _ = fmt.Fprintf(w, "%s", "This is port only accepts POST-methods")
	}
}

func generateKey() (ecdsa.PublicKey, big.Int) {
	key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	return key.PublicKey, *key.D
}

func getSharedSecret(response KeyResponse, privKey big.Int) [32]byte {
	generator, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	a, _ := generator.Curve.ScalarMult(&response.y, &response.y, privKey.Bytes())
	sharedSecret := sha256.Sum256(a.Bytes())

	return sharedSecret
}

func check(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}
