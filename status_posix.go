// +build darwin dragonfly freebsd linux nacl netbsd openbsd solaris windows

package main

import (
	"os/exec"
	"syscall"
)

func isExitStatus(err error, status int) bool {
	exitError, ok := err.(*exec.ExitError)
	if !ok {
		return false
	}
	ws, ok := exitError.Sys().(syscall.WaitStatus)
	if !ok {
		return false
	}
	return ws.ExitStatus() == status
}
