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

func main() {
	// Alert OR client that this node i active
	request, err := http.NewRequest("POST", "http://127.0.0.1:8080/connect", bytes.NewBuffer([]byte("KEY VALUE")))
	request.Header.Set("Content-Type", "text/plain; charset=UTF-8")

	client := http.Client{}
	response, err := client.Do(request)
	check(err)

	if response.Status == "200 OK" {
		fmt.Println("Connected to TOR-nexus")
	}

	// Start listening for requests
	http.HandleFunc("/", handler)
	err = http.ListenAndServe(":8081", nil)
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

		// Get site which is requested
		resp, err := http.Get(payload.Payload)
		if err != nil {
			log.Fatalln(err)
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

func check(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}
