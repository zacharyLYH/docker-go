package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

/*
Usage: your_docker.sh run <image> <command> <arg1> <arg2> ...
mydocker run ubuntu:latest /usr/local/bin/docker-explorer echo hey
*/
func main() {

	command := os.Args[3]
	args := os.Args[4:len(os.Args)]

	tempDir, err := os.MkdirTemp("", "mydocker")
	if err != nil {
		fmt.Printf("Error creating temp directory: %s\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tempDir)

	// Make sure the directory is executable
	if err := os.Chmod(tempDir, 0755); err != nil {
		fmt.Println("Error setting permissions:", err)
		os.Exit(1)
	}

	// The path where the binary should be copied
	binaryPath := filepath.Join(tempDir, filepath.Base(command))

	//copy command (name of file with functionality) into binaryPath (path that points to the tmp directory)
	if err := copyFile(command, binaryPath); err != nil {
		fmt.Println("Failed to copy binary:", err)
		os.Exit(1)
	}

	// /tmp/mydocker1515083054/docker-explorer
	// if _, err := os.Stat(binaryPath); err != nil {
	// 	fmt.Println("Failed to find executable before chroot:", err)
	// 	os.Exit(1)
	// } else {
	// 	fmt.Println("Executable confirmed before chroot at:", binaryPath)
	// }

	//now that copying of functionality is done, ready to change root
	if err := syscall.Chroot(tempDir); err != nil {
		fmt.Println("Failed to chroot:", err)
		os.Exit(1)
	}

	// Change working directory to the root after chroot because chroot doesn't change directory for us.
	if err := os.Chdir("/"); err != nil {
		fmt.Println("Failed to change directory after chroot:", err)
		os.Exit(1)
	}

	// viewFS()

	// if _, err := os.Stat(filepath.Base(command)); err != nil {
	// 	fmt.Println("Failed to find executable in new root:", err)
	// 	os.Exit(1)
	// } else {
	// 	fmt.Println("Executable found, proceeding with execution.")
	// }

	//Command spawns a new child process. But since it is a child process, it inherits the parents(the current) process' root directory (the temp directory)
	cmd := exec.Command("/"+filepath.Base(command), args...)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Println("Command execution failed:", err)
		os.Exit(cmd.ProcessState.ExitCode())
	}
}

func copyFile(src, dst string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	// Ensure executable permissions are correctly set
	return os.WriteFile(dst, input, 0755)
}

// func readToTerm(name string) {
// 	input, _ := os.ReadFile(name)
// 	fmt.Println(string(input))
// }

// func viewFS() {
// 	root := "/" // This refers to the new root after chroot
// 	fmt.Println("Listing files in chroot environment:")
// 	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
// 		if err != nil {
// 			fmt.Println("Unable to open: ", path)
// 		} else {
// 			fmt.Println(path) // Print each path
// 		}
// 		return nil
// 	})

// 	if err != nil {
// 		fmt.Printf("Error walking the path %q: %v\n", root, err)
// 	}
// }
