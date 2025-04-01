//go:build linux && seccomp
// +build linux,seccomp

// To build with seccomp support:
// go build -tags seccomp

package main

import (
	"fmt"

	"github.com/elastic/go-seccomp-bpf"
)

// Seccomp2Config holds the configuration for seccomp2 filtering
type Seccomp2Config struct {
	Enabled bool
	Profile string
	Action  string
}

// ParseSeccomp2Action converts string action to seccomp action
// Available actions:
// - "errno" (default): Return an error code (EPERM) for disallowed syscalls
// - "kill": Terminate the process immediately when a disallowed syscall is attempted
// - "trap": Send a SIGSYS signal to the process when a disallowed syscall is attempted
// - "log": Log disallowed syscalls but allow them to proceed (for debugging only)
func ParseSeccomp2Action(action string) seccomp.Action {
	switch action {
	case "kill":
		return seccomp.ActionKillThread
	case "trap":
		return seccomp.ActionTrap
	case "log":
		return seccomp.ActionLog
	case "errno", "":
		return seccomp.ActionErrno
	default:
		fmt.Printf("Warning: Unknown seccomp2 action '%s', using 'errno' instead\n", action)
		return seccomp.ActionErrno
	}
}

// InitSeccomp2 initializes the seccomp2 profile for Rye based on the provided configuration
func InitSeccomp2(config Seccomp2Config) error {
	// Skip if seccomp2 is disabled
	if !config.Enabled {
		return nil
	}

	// Only strict profile is supported
	if config.Profile != "strict" && config.Profile != "" {
		return fmt.Errorf("invalid seccomp2 profile: %s (only 'strict' is supported)", config.Profile)
	}

	// Essential syscalls needed for Go runtime
	allowedSyscalls := []string{
		// Basic file operations
		"read", "write", "open", "close", "stat", "fstat", "lstat",
		"poll", "lseek", "mmap", "mprotect", "munmap", "brk",

		// Signal handling
		"rt_sigaction", "rt_sigprocmask", "rt_sigreturn",

		// Process/thread operations
		"exit", "exit_group", "nanosleep", "clock_nanosleep",
		"getpid", "getuid", "geteuid", "getgid", "getegid",
		"getcwd", "chdir", "fchdir", "readlink",
		"access", "pipe", "pipe2", "dup", "dup2", "fcntl", "select",
		"getrlimit", "getrusage", "clock_gettime", "gettimeofday",
		"futex", "sched_yield", "getrandom",

		// Network operations
		"socket", "connect", "accept", "accept4", "bind", "listen",
		"getsockname", "getpeername", "socketpair", "setsockopt", "getsockopt",
		"shutdown", "recvfrom", "sendto", "recvmsg", "sendmsg",

		// Thread management
		"clone", "set_robust_list", "set_tid_address", "gettid", "tgkill",

		// Epoll for network polling
		"epoll_create", "epoll_create1", "epoll_ctl", "epoll_wait", "epoll_pwait",
	}

	// Create a new filter
	filter := seccomp.Filter{
		NoNewPrivs: true,
		Flag:       seccomp.FilterFlagTSync,
		Policy: seccomp.Policy{
			DefaultAction: ParseSeccomp2Action(config.Action),
			Syscalls: []seccomp.SyscallGroup{
				{
					Action: seccomp.ActionAllow,
					Names:  allowedSyscalls,
				},
			},
		},
	}

	// Load the filter
	if err := seccomp.LoadFilter(filter); err != nil {
		return fmt.Errorf("failed to load seccomp2 filter: %w", err)
	}

	return nil
}

// DisableSeccomp2ForDebug can be called to disable seccomp2 for debugging
func DisableSeccomp2ForDebug() {
	// This function is intentionally empty in the release build
}
