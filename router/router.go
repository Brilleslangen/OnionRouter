package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log"
	"math/rand"
	"net"
	"net/http"
)

type Payload struct {
	NextNode string
	Payload  []byte
}

type Node struct {
	IP        string
	Port      string
	PublicKey string
}

func (node *Node) address() string {
	return node.IP + ":" + node.Port
}

var nodes []Node

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
		// Extract IP-address and key
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		var node Node
		decoder := json.NewDecoder(r.Body)
		err = decoder.Decode(&node)
		check(err)
		node.IP = ip
		fmt.Println("IP: ", node.IP, " Port: ", node.Port, " PublicKey: "+node.PublicKey)

		// Add to available nodes
		nodes = append(nodes, node)
	}
}

func sendThroughNodes(url string) *http.Response {
	// Select random nodes and pack payload in encrypted layers
	payload, err := selectAndPack(url)

	// Create request
	request, err := http.NewRequest("POST", "http://"+payload.NextNode, bytes.NewBuffer(payload.Payload))
	check(err)
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")

	// Send request and collect response
	client := http.Client{}
	resp, err := client.Do(request)
	check(err)

	return resp
}

func selectAndPack(url string) (Payload, error) {
	var selectedNodes [3]Node

	// Ensure there are at least three nodes to traverse
	if len(nodes) < 3 {
		return Payload{}, errors.New("there has to be at least three nodes connected")
	}

	// Randomly select three unique nodes
	for i := 0; i < 3; i++ {
		currentNode := nodes[rand.Intn(len(nodes))]
		for _, prevNode := range selectedNodes {
			if currentNode == prevNode {
				i--
				continue
			}
			selectedNodes[i] = currentNode
		}
	}

	// Recursively pack payload
	currentPayload := Payload{"", []byte(url)}
	for i := 0; i < 2; i++ {
		// Convert previous payload to a JSON string and pack into new payload
		jsonPayload, err := json.Marshal(currentPayload)
		check(err)
		currentPayload = Payload{selectedNodes[i].address(), jsonPayload}
	}

	// Pack final payload to be sent from this entity
	jsonFinal, err := json.Marshal(currentPayload)
	check(err)

	return Payload{selectedNodes[2].address(), jsonFinal}, nil
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
