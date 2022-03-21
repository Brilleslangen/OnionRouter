package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
)

var payload Payload

func main() {
	http.HandleFunc("/", handler)
	err := http.ListenAndServe(":8080", nil) // setting listening port
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	//fmt.Println("method:", r.Method)
	if r.Method == "POST" {
		
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
