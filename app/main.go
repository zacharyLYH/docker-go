package main

import (
	"fmt"
	"os"
	"os/exec"
)

/*
Usage: your_docker.sh run <image> <command> <arg1> <arg2> ...
mydocker run ubuntu:latest /usr/local/bin/docker-explorer echo hey
*/
func main() {

	command := os.Args[3]
	args := os.Args[4:len(os.Args)]

	cmd := exec.Command(command, args...)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		fmt.Println(err)
		os.Exit(cmd.ProcessState.ExitCode())
	}
}
