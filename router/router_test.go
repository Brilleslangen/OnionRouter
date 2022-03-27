package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestRouter(t *testing.T) {
	// Initiate router
	go func() {
		cmd := exec.Command("go", "run", "router/router.go")
		cmd.Dir = "../"
		err := cmd.Run()
		check(err)
	}()

	fmt.Println("Router initiated")

	// Initiate 6 nodes, so the router can choose a random relay-pattern
	for i := 1; i <= 6; i++ {
		go func() {
			err := exec.Command("go", "run", "node/node.go", "808"+strconv.Itoa(i)).Start()
			check(err)
		}()
	}
	fmt.Println("All instances are running")

	testUrl := "http://127.0.0.1:8080/"
	testMethod := "POST"
	testPayload := strings.NewReader("code=nginx.org")

	testClient := &http.Client{}
	testReq, err := http.NewRequest(testMethod, testUrl, testPayload)

	testReq.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	testRes, err := testClient.Do(testReq)

	check(err)

	defer testRes.Body.Close()

	testBody, err := ioutil.ReadAll(testRes.Body)

	time.Sleep(time.Second * 5)

	url := "http://nginx.org"
	method := "GET"
	payload := strings.NewReader("code=nginx.org")

	req, err := http.NewRequest(method, url, payload)

	check(err)

	client := &http.Client{}

	res, err := client.Do(req)

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)

	if !reflect.DeepEqual(testBody, body) {
		t.Errorf("FAILED, GOT %x, EXPECTED %x", testBody, body)
	}
}
