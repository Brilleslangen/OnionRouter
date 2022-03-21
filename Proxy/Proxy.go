package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
)

type Payload struct {
	NextNode string
	Payload  string
}

var nodes map[string]string

func main() {
	// Initiate map to hold nodes
	nodes = make(map[string]string)

	// Set handlers
	http.HandleFunc("/connect", connectNode)
	http.HandleFunc("/", handler)
	err := http.ListenAndServe(":8080", nil) // setting listening port
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

		resp := sendThroughNodes("https://" + linkAddress)

		t, _ := template.ParseFiles("../html/blank.html")
		err = t.Execute(w, nil)
		check(err)
		
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
		// Extract address and key
		nodeAddress := r.RemoteAddr
		encryptionKey, err := io.ReadAll(r.Body)
		check(err)
		fmt.Println("Address: ", nodeAddress, " Key: "+string(encryptionKey))

		// Add to available nodes
		nodes[nodeAddress] = string(encryptionKey)
	}
}

func sendThroughNodes(url string) *http.Response {
	// Create payload
	payload := Payload{"", url}

	// Convert to JSON-string
	jsonData, err := json.Marshal(payload)
	check(err)

	// Create request
	request, err := http.NewRequest("POST", "http://127.0.0.1:8081/", bytes.NewBuffer(jsonData))
	check(err)
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")

	// Initiate http request and collect response
	client := http.Client{}
	resp, err := client.Do(request)
	check(err)

	return resp
}

func check(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}
