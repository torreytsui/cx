// +build darwin linux freebsd

package main

import (
	"os"
	"syscall"
)

func sysExec(path string, args []string, env []string) error {
	return syscall.Exec(path, args, env)
}

func homePath() string {
	return os.Getenv("HOME")
}
