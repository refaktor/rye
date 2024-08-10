//go:build !windows
// +build !windows

package util

import "syscall"

func KillProcess(pid int) error {
	return syscall.Kill(pid, syscall.SIGTERM)
}
