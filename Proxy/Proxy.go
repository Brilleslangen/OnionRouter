package main

import (
	"bytes"
	"html/template"
	"io"
	"log"
	"net/http"
)

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

		resp, err := http.Get("https://" + linkAddress)
		if err != nil {
			log.Fatalln(err)
		}

		t, _ := template.ParseFiles("../html/blank.html")
		err = t.Execute(w, nil)
		check(err)

		// Print to client
		_, err = io.Copy(w, resp.Body)
		check(err)
	}
}

func sendThroughNodes(url string) {
	request, error := http.NewRequest("POST", httpposturl, bytes.NewBuffer(jsonData))
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
