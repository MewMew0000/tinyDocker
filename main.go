package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"syscall"
	"tinyDocker/workspace"
)

func main() {
	switch os.Args[1] {
	//parent process
	case "run":
		initCmd, err := os.Readlink("/proc/self/exe") // get/reflect its own path (tinydocker here)
		if err != nil {
			fmt.Println("get init process error", err)
			return
		}
		containerName := os.Args[2]
		os.Args[1] = "init"
		cmd := exec.Command(initCmd, os.Args[1:]...) // run itself with init
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
		workspace.DelMntNamespace(containerName)
		return
	//child process
	case "init":
		var (
			containerName = os.Args[2]
			cmd           = os.Args[3]
		)
		if err := workspace.SetMntNamespace(containerName); err != nil {
			fmt.Println("set mnt namespace fail", err)
			return
		}
		syscall.Chdir("/")
		defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
		syscall.Mount("proc", "proc", "proc", uintptr(defaultMountFlags), "")

		err := syscall.Exec(cmd, os.Args[3:], os.Environ()) // replace current process with the new one
		if err != nil {
			fmt.Println("exec proc fail", err)
			return
		}
		fmt.Println("forever exec it")
		return
	default:
		fmt.Println("command not found")
	}
}
