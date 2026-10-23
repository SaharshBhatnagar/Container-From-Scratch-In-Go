package main

import (
	"os"
	"fmt"
	"os/exec"
	"syscall"
	"path/filepath"
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
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
		Unshareflags: syscall.CLONE_NEWNS,
	}

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	defer os.Remove(filepath.Join("/sys/fs/cgroup", "my-container"))

	err := cmd.Run()

	errHandle(err)

}

func child() {
	fmt.Printf("Running %v as %d\n", os.Args[2:], os.Getpid())

	err := syscall.Sethostname([]byte("my-container"))

	errHandle(err)

	cg()

	err = syscall.Chroot("alpine-fs")

	errHandle(err)

	err = syscall.Chdir("/")

	errHandle(err)

	err = syscall.Mount("proc", "proc", "proc", uintptr(0), "")

	errHandle(err)

	defer syscall.Unmount("proc", 0)

	cmd := exec.Command(os.Args[2], os.Args[3:]...)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()

	errHandle(err)

}

func errHandle(err error) {
	if err != nil {
		panic(err)
	}

}

func cg() {
	newpath := filepath.Join("/sys/fs/cgroup", "my-container")

	err := os.Mkdir(newpath, 0755)

	if err != nil && !os.IsExist(err) {
		errHandle(err)
	}

	err = os.WriteFile(filepath.Join(newpath, "pids.max"), []byte("20"), 0700)

	errHandle(err)

	
	err = os.WriteFile(filepath.Join(newpath, "memory.max"), []byte("20M"), 0700)
	
	errHandle(err)
	
	err = os.WriteFile(filepath.Join(newpath, "memory.swap.max"), []byte("0"), 0700)
	
	errHandle(err)

	err = os.WriteFile(filepath.Join(newpath, "cgroup.procs"), []byte("0"), 0700)

	errHandle(err)

}