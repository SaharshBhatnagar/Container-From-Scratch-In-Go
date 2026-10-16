package main

import (
	"os"
	"fmt"
	"os/exec"
	"syscall"
)

func main() {

	if len(os.Args) < 3 {
		panic("usage: go run main.go run <command>")
	}

	switch os.Args[1] {
		case "run":
			run()
		case "child":
			child()
		default:
			panic("bad command")
	}

}

func run() {
	fmt.Printf("Running %v as %d\n", os.Args[2:], os.Getpid())

	cmd := exec.Command("/proc/self/exe", append([]string{"child"}, os.Args[2:]...)...)


	cmd.SysProcAttr = &syscall.SysProcAttr {
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID,
	}

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()

	errHandle(err)

}

func child() {
	fmt.Printf("Running %v as %d\n", os.Args[2:], os.Getpid())

	err := syscall.Sethostname([]byte("my-container"))

	errHandle(err)

	err = syscall.Chroot("alpine-fs")

	errHandle(err)

	err = syscall.Chdir("/")

	errHandle(err)

	err = syscall.Mount("proc", "proc", "proc", uintptr(0), "")

	errHandle(err)

	cmd := exec.Command(os.Args[2], os.Args[3:]...)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()

	errHandle(err)

	err = syscall.Unmount("proc", 0)

	errHandle(err)
}

func errHandle(err error) {
	if err != nil {
		panic(err)
	}

}