package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"syscall"
)

// mount --make-rprivate /
// mount -t proc proc /proc
// ipcmk -Q
// sudo ./tinyDocker  run /bin/sh
func main() {
	switch os.Args[1] {
	case "run":
		fmt.Println("run pid", os.Getpid(), "ppid", os.Getppid())
		initCmd, err := os.Readlink("/proc/self/exe")
		if err != nil {
			fmt.Println("get init process error", err)
			return
		}
		os.Args[1] = "init"
		cmd := exec.Command(initCmd, os.Args[1:]...)
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWNET | syscall.CLONE_NEWIPC,
		}
		cmd.Env = os.Environ()
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		if err != nil {
			log.Fatal(err)
		}
	case "init":
		syscall.Chroot("./ubuntu2204_rootfs")
		syscall.Chdir("/")
		defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
		syscall.Mount("proc", "proc", "proc", uintptr(defaultMountFlags), "")
		cmd := os.Args[2]
		fmt.Println("exec cmd=", cmd)
		err := syscall.Exec(cmd, os.Args[2:], os.Environ())
		if err != nil {
			fmt.Println("exec proc fail", err)
			return
		}
		fmt.Println("forever exec it")
		return
	}
}
