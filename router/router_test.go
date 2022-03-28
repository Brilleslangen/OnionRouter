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
	initRouter()

	// Wait for router and nodes to start
	// Increase this if the test fails
	time.Sleep(10 * time.Second)

	// Build test request
	testUrl := "http://127.0.0.1:8080/"
	testMethod := "POST"
	testPayload := strings.NewReader("code=nginx.org")
	testClient := &http.Client{}
	testReq, err := http.NewRequest(testMethod, testUrl, testPayload)
	testReq.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	// Send test request
	testRes, err := testClient.Do(testReq)
	defer testRes.Body.Close()
	check(err)

	// Read the body of the response for comparison later
	testBody, err := ioutil.ReadAll(testRes.Body)

	// Build request to compare with the test request
	url := "http://nginx.org"
	method := "GET"
	payload := strings.NewReader("code=nginx.org")
	req, err := http.NewRequest(method, url, payload)
	client := &http.Client{}

	// Send request and read the body of the response
	res, err := client.Do(req)
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)

	killProcesses()
	// Compare the two bodies
	if !reflect.DeepEqual(testBody, body) {
		t.Errorf("FAILED, GOT %x, EXPECTED %x", testBody, body)
	}

}

// initRouter spins up a router with num nodes to choose from
func initRouter() {
	// Initiate router
	cmd := exec.Command("go", "run", "router/router.go")
	cmd.Dir = "../"
	err := cmd.Start()
	check(err)
	time.Sleep(1 * time.Second)

	fmt.Println("Router initiated")

	// Initiate num nodes, so the router can choose a random relay-pattern
	for i := 1; i <= 3; i++ {
		cmd := exec.Command("go", "run", "node/node.go", "808"+strconv.Itoa(i))
		cmd.Dir = "../"
		err := cmd.Start()
		check(err)
	}
	fmt.Println("All instances are running")
}

func killProcesses() {
	err := exec.Command("pkill", "router").Run()
	err = exec.Command("pkill", "node").Run()
	err = exec.Command("pkill", "node").Run()
	err = exec.Command("pkill", "node").Run()

	check(err)
	fmt.Println("All processes killed")
}
