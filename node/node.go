package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/json"
	"fmt"
	. "github.com/Brilleslangen/OnionRouter/ecdh"
	. "github.com/Brilleslangen/OnionRouter/orstructs"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
)

var nodeKey *ecdsa.PrivateKey
var SharedSecret []byte
var PORT string
var RouterIP = "172.18.0.1" // If run in dockered container

func main() {
	if len(os.Args) < 2 {
		fmt.Println("You have to assign a port for the node listen on. \n " +
			"You can do this by adding port number as the first command line argument.")
	}
	// If Router IP is specified
	if len(os.Args) > 2 {
		RouterIP = os.Args[2]
	}

	// Initialize keypair in the node
	nodeKey, _ = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	PORT = os.Args[1]

	// Alert router that this node is active
	jsonDetails, err := json.Marshal(Node{IP: "TBD", Port: PORT, PubX: nodeKey.X, PubY: nodeKey.Y, SharedSecret: []byte{}})
	check(err)
	request, err := http.NewRequest("POST", "http://"+RouterIP+":8080/connect", bytes.NewBuffer(jsonDetails))
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")
	client := http.Client{}
	response, err := client.Do(request)
	check(err)

	// Response from the router contains the routers public key contained in two strings
	var routerKey KeyResponse
	decoder := json.NewDecoder(response.Body)
	err = decoder.Decode(&routerKey)
	check(err)

	SharedSecret = ShareSecret(nodeKey, *routerKey.X, *routerKey.Y)
	check(err)

	if response.Status == "200 OK" {
		fmt.Println("Connected to router")
		fmt.Println(" Available on port:", PORT)
		fmt.Printf(" Shared Secret Symmetric Key: %x\n", SharedSecret)
	} else {
		fmt.Println("Status OK not received. \nStatus code received:", response.Status)
	}

	// Start listening for requests
	http.HandleFunc("/", handler)

	// http.ListenAndServe assigns a new thread to each connection
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
		decryptedBody, err := Decrypt(body, SharedSecret)
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

		// Read to bytes
		bodyToBytes, err := io.ReadAll(resp.Body)
		check(err)

		// Encrypt JSON to bytes
		encryptedResponse, err := Encrypt(bodyToBytes, SharedSecret)
		check(err)

		// Write to response to requesting entity

		_, err = w.Write(encryptedResponse)
		check(err)
	} else {
		t, _ := template.ParseFiles("../html/blank.html")
		err := t.Execute(w, nil)
		check(err)
		_, _ = fmt.Fprintf(w, "%s", "This is port only accepts POST-methods")
	}
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
