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
	PublicKeyX   string
	PublicKeyY   string
	SharedSecret [32]byte
}

type KeyResponse struct {
	x string
	y string
}

var nodeKey, _ = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
var SharedSecret [32]byte

func main() {
	PORT := "8088"
	dummyArr := new([32]byte)
	// Alert router that this node is active
	jsonDetails, err := json.Marshal(NodeDetails{"TBD", PORT, nodeKey.X.Text(10), nodeKey.Y.Text(10), *dummyArr})
	check(err)
	request, err := http.NewRequest("POST", "http://127.0.0.1:8080/connect", bytes.NewBuffer(jsonDetails))
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")
	client := http.Client{}
	response, err := client.Do(request)
	var keyResponse KeyResponse
	decoder := json.NewDecoder(response.Body)
	err = decoder.Decode(&keyResponse)
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

func getSharedSecret(response KeyResponse) [32]byte {
	x, _ := new(big.Int).SetString(response.x, 16)
	y, _ := new(big.Int).SetString(response.y, 16)
	a, _ := nodeKey.PublicKey.Curve.ScalarMult(x, y, nodeKey.D.Bytes())
	sharedSecret := sha256.Sum256(a.Bytes())

	return sharedSecret
}

func check(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}
