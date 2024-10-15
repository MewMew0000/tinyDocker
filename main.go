package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"syscall"
)

func main() {
	switch os.Args[1] {
	//parent process
	case "run":
		fmt.Println("run pid", os.Getpid(), "ppid", os.Getppid())
		initCmd, err := os.Readlink("/proc/self/exe") // get/reflect its own path (tinydocker here)
		if err != nil {
			fmt.Println("get init process error", err)
			return
		}
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
	//child process
	case "init":
		pwd, err := os.Getwd()
		if err != nil {
			fmt.Println("pwd", err)
			return
		}
		path := pwd + "/ubuntu2204_rootfs"
		// ensure that change in parent process will not affect the child process
		syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, "")
		// when path is mounted to child process, it will in the different namespace with parent process's filesystem
		if err := syscall.Mount(path, path, "bind", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
			fmt.Println("mount fail", err)
			return
		}

		if err := os.MkdirAll(path+"/.old", 0700); err != nil {
			fmt.Println("mkdir fail", err)
			return
		}
		// give a new isolated filesystem to the child process
		err = syscall.PivotRoot(path, path+"/.old")
		if err != nil {
			fmt.Println("pivot root fail", err)
			return
		}
		syscall.Chdir("/")
		defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
		syscall.Mount("proc", "proc", "proc", uintptr(defaultMountFlags), "")
		cmd := os.Args[2]
		fmt.Println("exec cmd=", cmd)
		err = syscall.Exec(cmd, os.Args[2:], os.Environ()) // replace current process with the new one
		if err != nil {
			fmt.Println("exec proc fail", err)
			return
		}
		fmt.Println("forever exec it")
		return
	}
}
