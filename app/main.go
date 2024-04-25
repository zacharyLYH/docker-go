package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"syscall"
)

// Usage: your_docker.sh run <image> <command> <arg1> <arg2> ...
func main() {
	image := os.Args[2]
	imageDir := fmt.Sprintf("./images/%s", image)

	if _, err := os.Stat(imageDir); err != nil {
		if os.IsNotExist(err) {
			imageDir, err = ImagePull(image, "./images")
			if err != nil {
				logError(err, "Error while pulling the image")
				os.Exit(255)
			}
		}
	}

	command := os.Args[3]
	args := os.Args[4:len(os.Args)]

	// change root filesystem for the child process using chroot
	// this is necessary to make the child process believe it is running in a different root filesystem
	// not needed during final task since we are downloading our image
	IsolatedProcess()

	cmd := exec.Command(command, args...)

	//It will create a process Isolation by creating a new Namespace
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWPID,
		Chroot:     imageDir, //changed the
		// this CLONE_NEWPID Unshare the PID namespace, so that the calling
		// process has a new PID namespace for its children which is
		// not shared with any previously existing process.  The
		// calling process is not moved into the new namespace.  The
		// first child created by the calling process will have the
		// process ID 1 and will assume the role of init(1) in the
		// new namespace
	}

	// bind the standard input, output and error to the parent process
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()

	// exit with the same exit code as the child process
	if err != nil {
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			os.Exit(exitError.ExitCode())
		}
	}
}

func IsolatedProcess() {
	rootFsPath, err := os.MkdirTemp("", "temp_")
	if err != nil {
		logError(err, "Failed to create temporary directory")
	}
	err = os.Chmod(rootFsPath, 0755)
	if err != nil {
		logError(err, "Failed to change permissions of temporary directory")
	}

	defer os.Remove(rootFsPath)

	binPath := "/usr/local/bin"

	err = os.MkdirAll(path.Join(rootFsPath, binPath), 0755)
	if err != nil {
		logError(err, "Failed to create bin directory")
	}

	//Link command is used to link the existing path with the new path in this case /tmp/temp_*/usr/local/bin/docker-explorer
	os.Link("/usr/local/bin/docker-explorer", path.Join(rootFsPath, "/usr/local/bin/docker-explorer"))
	if err != nil {
		logError(err, "Failed to copy binaries to root file system")
	}

	err = syscall.Chroot(rootFsPath)
	if err != nil {
		logError(err, "Failed to change root filesystem")
	}
}

func logError(err error, errorMessage string) {
	log.Fatalf("%s: %v", errorMessage, err)
	os.Exit(1)
}
