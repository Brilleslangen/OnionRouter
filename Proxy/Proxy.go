package main

import (
	"bytes"
	"encoding/json"
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
	http.HandleFunc("/", handler)
	err := http.ListenAndServe(":8080", nil) // setting listening port
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	//fmt.Println("method:", r.Method)
	if r.Method == "GET" {
		t, _ := template.ParseFiles("../html/index.html")
		err := t.Execute(w, nil)
		check(err)
	} else {
		err := r.ParseForm()
		check(err)

		linkAddress := r.Form["code"][0]

		/*
			resp, err := http.Get("https://" + linkAddress)
			if err != nil {
				log.Fatalln(err)
			}
		*/

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
		panic(err)
	}
}
