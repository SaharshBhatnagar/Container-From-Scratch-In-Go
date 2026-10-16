package main

import (
	"os"
	"fmt"
	"os/exec"
)

func main() {

	if len(os.Args) < 3 {
		panic("usage: go run main.go run <command>")
	}

	switch os.Args[1] {
		case "run":
			run()
		default:
			panic("bad command")
	}

}

func run() {
	fmt.Printf("Running %v \n", os.Args[2:])

	cmd := exec.Command(os.Args[2], os.Args[3:]...)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()

	if err != nil {
		panic(err)
	}

}