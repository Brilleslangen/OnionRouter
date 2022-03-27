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
	"os"
	"reflect"
)

var nodes []Node
var routerKey *ecdsa.PrivateKey

func main() {
	routerKey, _ = ecdsa.GenerateKey(elliptic.P256(), random.Reader)

	// Set handlers
	http.HandleFunc("/connect", connectNode)
	http.HandleFunc("/", handler)
	fmt.Println("Router initiated")

	// http.ListenAndServe assigns a new thread to each connection
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

// handler handles requests to the "/" endpoint.
func handler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		wd, _ := os.Getwd()
		t, err := template.ParseFiles(wd + "/html/index.html")
		check(err)
		err = t.Execute(w, nil)
		check(err)
	} else {
		err := r.ParseForm()
		check(err)

		linkAddress := r.Form["code"][0]

		var respBody []byte
		if linkAddress[0:7] == "http://" {
			respBody = sendThroughNodes(linkAddress)
		} else {
			respBody = sendThroughNodes("http://" + linkAddress)
		}

		// Print to client
		_, err = w.Write(respBody)
		check(err)
	}
}

// connectNode connects a node making a request to the "/connect" endpoint.
func connectNode(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		// Extract IP-address and key
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		var node Node
		decoder := json.NewDecoder(r.Body)
		err = decoder.Decode(&node)
		check(err)
		node.IP = ip

		// Generate shared secret
		node.SharedSecret = ShareSecret(routerKey, *node.PubX, *node.PubY)
		fmt.Println("\nA new node has connected:")
		fmt.Printf(" IP: %s \n Port: %s \n Shared Secret Symmetric Key: %x\n", ip, node.Port, node.SharedSecret)

		// Create response with public key coordinates
		response := KeyResponse{X: routerKey.X, Y: routerKey.Y}
		jsonDetails, err := json.Marshal(response)
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		_, err = w.Write(jsonDetails)
		check(err)

		// Add to available nodes
		nodes = append(nodes, node)
	}
}

// sendThroughNodes makes an HTTP-request to the desired URL.
func sendThroughNodes(url string) []byte {
	// Select random nodes and pack payload in encrypted layers
	selectedNodes, payload, err := selectAndPack(url)

	// Create request
	request, err := http.NewRequest("POST", "http://"+string(payload.NextNode), bytes.NewBuffer(payload.Content))
	check(err)
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")

	// Send request and collect response
	client := http.Client{}
	resp, err := client.Do(request)
	check(err)

	// Unpack decryption layers and replace response body
	respBody := unpack(resp.Body, selectedNodes)

	return respBody
}

// selectAndPack randomly selects three nodes and recursively packs the payload and address of the next node
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
	currentPayload := Payload{Content: []byte(url)}
	for i := 0; i < 2; i++ {
		// Convert previous payload to a JSON string and pack into new payload
		jsonPayload, err := json.Marshal(currentPayload)
		check(err)
		encryptedPayload, err := Encrypt(jsonPayload, selectedNodes[i].SharedSecret)
		currentPayload = Payload{NextNode: selectedNodes[i].Address(), Content: encryptedPayload}
	}

	// Pack final payload to be sent from this entity
	jsonFinal, err := json.Marshal(currentPayload)
	check(err)
	encryptedPayload, err := Encrypt(jsonFinal, selectedNodes[2].SharedSecret)

	return selectedNodes, Payload{NextNode: selectedNodes[2].Address(), Content: encryptedPayload}, nil
}

// unpack decrypts three layers of encryption on a response body
func unpack(respBody io.ReadCloser, selectedNodes [3]Node) []byte {
	body, err := io.ReadAll(respBody)

	for i := 2; i >= 0; i-- {
		// Decrypt bytes
		body, err = Decrypt(body, selectedNodes[i].SharedSecret)
		check(err)
	}

	return body
}

func check(err error) {
	if err != nil {
		fmt.Println(err)
	}
}
