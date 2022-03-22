package main

import (
	"log"
	"os/exec"
)

func main() {
	err := exec.Command("cd", "router", "&&", "go", "run", "router.go").Run()
	check(err)
	err = exec.Command("cd", "../node").Run()
	check(err)

	for i := 0; i < 6; i++ {
		err = exec.Command("go").Run()
	}
}

func check(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}
