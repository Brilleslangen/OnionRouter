package main

import (
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
	http.HandleFunc("/", handler)
	err := http.ListenAndServe(":8081", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		fmt.Println(r.Header)
		var payload Payload
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&payload)
		check(err)
		resp, err := http.Get(payload.Payload)
		if err != nil {
			log.Fatalln(err)
		}

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
