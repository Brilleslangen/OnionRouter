package main

import (
	"fmt"
	"log"
	"os/exec"
	"strconv"
)

func main() {

	// Initiate router
	err := exec.Command("go", "run", "router/router.go").Start()
	check(err)
	fmt.Println("Router initiated")

	// Initiate 6 nodes, so the router can choose a random relay-pattern
	for i := 1; i <= 5; i++ {
		err := exec.Command("go", "run", "node/node.go", "808"+strconv.Itoa(i)).Start()
		check(err)
	}
	fmt.Println("All instances are running")
	err = exec.Command("go", "run", "node/node.go", "808"+strconv.Itoa(6)).Run()
	check(err)
}

func check(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}
