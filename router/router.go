package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	random "crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	. "github.com/Brilleslangen/OnionRouter/ecdh"
	. "github.com/Brilleslangen/OnionRouter/orstructs"
	"html/template"
	"io"
	"log"
	"math/rand"
	"net"
	"net/http"
	"reflect"
)

var nodes []Node
var routerKey *ecdsa.PrivateKey

func main() {

	// Set handlers
	http.HandleFunc("/connect", connectNode)
	http.HandleFunc("/", handler)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		t, _ := template.ParseFiles("../html/index.html")
		err := t.Execute(w, nil)
		check(err)
	} else {
		err := r.ParseForm()
		check(err)

		linkAddress := r.Form["code"][0]

		resp := sendThroughNodes("http://" + linkAddress)

		// Print to client
		defer func() {
			err = resp.Body.Close()
			check(err)
		}()
		_, err = io.Copy(w, resp.Body)
		check(err)
	}
}

func connectNode(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		routerKey, _ = ecdsa.GenerateKey(elliptic.P256(), random.Reader)
		// Extract IP-address and key
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		var node Node
		decoder := json.NewDecoder(r.Body)
		err = decoder.Decode(&node)
		check(err)
		node.IP = ip

		node.SharedSecret = EstablishSharedSecret(node, routerKey)
		fmt.Printf("\n IP: %x \n Port: %x \n PublicKey: (%x,%x) \n Shared Secret: %x", ip, node.Port, node.PublicKeyX, node.PublicKeyY, node.SharedSecret)
		x := routerKey.X.Text(10)
		y := routerKey.Y.Text(10)
		response := KeyResponse{X: x, Y: y}
		jsonDetails, err := json.Marshal(response)
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		_, err = w.Write(jsonDetails)
		check(err)

		// Add to available nodes
		nodes = append(nodes, node)
	}
}

func sendThroughNodes(url string) *http.Response {
	// Select random nodes and pack payload in encrypted layers
	selectedNodes, payload, err := selectAndPack(url)

	// Create request
	request, err := http.NewRequest("POST", "http://"+string(payload.NextNode), bytes.NewBuffer(payload.Payload))
	check(err)
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")

	// Send request and collect response
	client := http.Client{}
	resp, err := client.Do(request)
	check(err)

	// Unpack decryption layers and replace response body
	resp.Body = unpack(resp.Body, selectedNodes)

	return resp
}

func selectAndPack(url string) ([3]Node, Payload, error) {
	var selectedNodes [3]Node

	// Ensure there are at least three nodes to traverse
	if len(nodes) < 3 {
		return selectedNodes, Payload{}, errors.New("there has to be at least three nodes connected")
	}

	// Randomly select three unique nodes
	for i := 0; i < 3; i++ {
		currentNode := nodes[rand.Intn(len(nodes))]
		for _, prevNode := range selectedNodes {
			if reflect.DeepEqual(currentNode, prevNode) {
				i--
				continue
			}
		}
		selectedNodes[i] = currentNode
	}

	// Recursively pack payload
	currentPayload := Payload{Payload: []byte(url)}
	for i := 0; i < 2; i++ {
		// Convert previous payload to a JSON string and pack into new payload
		jsonPayload, err := json.Marshal(currentPayload)
		check(err)
		encryptedPayload, _ := Encrypt(selectedNodes[i].SharedSecret[:], string(jsonPayload))
		currentPayload = Payload{NextNode: selectedNodes[i].Address(), Payload: []byte(encryptedPayload)}
	}

	// Pack final payload to be sent from this entity
	jsonFinal, err := json.Marshal(currentPayload)
	check(err)
	encryptedPayload, _ := Encrypt(selectedNodes[2].SharedSecret[:], string(jsonFinal))

	return selectedNodes, Payload{NextNode: selectedNodes[2].Address(), Payload: []byte(encryptedPayload)}, nil
}

func unpack(respBody io.ReadCloser, selectedNodes [3]Node) io.ReadCloser {
	for i := 2; i < 0; i-- {
		// Convert from io.ReadCloser to encrypted bytes
		encryptedBody, err := io.ReadAll(respBody)
		check(err)

		// Decrypt bytes
		decryptedBody, err := Decrypt(selectedNodes[i].SharedSecret[:], string(encryptedBody))

		// Convert decrypted bytes to JSON
		err = json.Unmarshal([]byte(decryptedBody), &respBody)
	}
	return respBody
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
