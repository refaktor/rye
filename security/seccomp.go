//go:build linux && seccomp
// +build linux,seccomp

// To build with seccomp support:
// go build -tags seccomp

package security

import (
	"fmt"
	"os"

	"github.com/elastic/go-seccomp-bpf"
)

// SeccompConfig holds the configuration for seccomp filtering
type SeccompConfig struct {
	Enabled bool
	Profile string
	Action  string
}

// CurrentSeccompProfile stores the active seccomp profile
var CurrentSeccompProfile string

// isValidSeccompProfile checks if the specified profile is valid
func isValidSeccompProfile(profile string) bool {
	validProfiles := []string{"strict", "readonly"}
	for _, p := range validProfiles {
		if profile == p {
			return true
		}
	}
	return false
}

// ParseSeccompAction converts string action to seccomp action
// Available actions:
// - "errno" (default): Return an error code (EPERM) for disallowed syscalls
// - "kill": Terminate the process immediately when a disallowed syscall is attempted
// - "trap": Send a SIGSYS signal to the process when a disallowed syscall is attempted
// - "log": Log disallowed syscalls but allow them to proceed (for debugging only)
func ParseSeccompAction(action string) seccomp.Action {
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
		fmt.Printf("Warning: Unknown seccomp action '%s', using 'errno' instead\n", action)
		return seccomp.ActionErrno
	}
}

// InitSeccomp initializes the seccomp profile for Rye based on the provided configuration
func InitSeccomp(config SeccompConfig) error {
	// Skip if seccomp is disabled
	if !config.Enabled {
		return nil
	}

	// Check if the profile is valid
	if !isValidSeccompProfile(config.Profile) {
		return fmt.Errorf("invalid seccomp profile: %s (valid profiles: strict, readonly)", config.Profile)
	}

	var allowedSyscalls []string

	// Select the appropriate syscalls based on the profile
	if config.Profile == "readonly" {
		allowedSyscalls = []string{
			// Basic file operations (read-only)
			"read", "close", "stat", "fstat", "lstat",
			"poll", "lseek", "mmap", "mprotect", "munmap", "brk",

			// Allow write for stdout/stderr
			"write",

			// Signal handling
			"rt_sigaction", "rt_sigprocmask", "rt_sigreturn", "sigaltstack",

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

			// Allow open with O_RDONLY
			"open", "openat",

			// Memory management
			"madvise", "mincore",

			// System info
			"uname", "arch_prctl",

			// I/O operations
			"writev", "pread64", "pwrite64", "ioctl",

			// Additional syscalls that might be needed with CGO_ENABLED=0
			"rseq", "prlimit64", "statx", "newfstatat",
		}
	} else {
		// Default to strict profile
		allowedSyscalls = []string{

			// --- Absolute minimum for Go runtime ---
			"read", "write", "open", "close", "fstat",
			"mmap", "munmap", "mprotect", "brk",
			"rt_sigaction", "rt_sigprocmask", "rt_sigreturn",
			"sched_yield", "clone", "futex",
			"gettid", "exit", "exit_group",

			// --- Extended epoll requirements ---
			"epoll_create1", "epoll_ctl", "epoll_pwait",
			"epoll_wait", "poll", "select",

			// --- Memory management ---
			"madvise", "mincore",

			// --- Time/clock ---
			"clock_gettime", "clock_nanosleep", "nanosleep",
			"gettimeofday",

			// --- I/O ---
			"writev", "pread64", "pwrite64",
			"ioctl", // Required for terminal handling

			// --- Filesystem (basic) ---
			"lseek", "access", "stat", //"fstat",

			// --- Signal handling ---
			"sigaltstack",

			// --- System info ---
			"uname", "arch_prctl",

			// --- Pipe/socketpair (used internally) ---
			"pipe", "pipe2",

			// --- Go runtime essentials ---
			/*	"read", "write", "open", "close", "fstat",
				"mmap", "munmap", "mprotect", "brk", "rt_sigaction",
				"rt_sigprocmask", "sched_yield", "clone", "execve", // `execve` for Go's os/exec
				"gettid", "futex", "exit", "exit_group",

				// --- I/O (allow stdout/stderr) ---
				"writev", "ioctl", // `ioctl` for terminal handling

				// --- Time/clock ---
				"clock_gettime", "time", "nanosleep",

				// --- Essential for Go runtime ---
				"epoll_create", "epoll_ctl", "epoll_wait", // Linux
				"poll", "select", // Fallbacks
				// --- Previous minimal syscalls (from earlier) ---
				//"read", "write", "clone", "futex", "mmap",
				//"mprotect", "rt_sigaction", "exit_group",

				// --- Network (optional, block if unused) ---
				// "socket", "connect", "accept",

				// Basic file operations
				/* "read", "write", "open", "close", "stat", "fstat", "lstat",
				"poll", "lseek", "mmap", "mprotect", "munmap", "brk",

				// Signal handling
				"rt_sigaction", "rt_sigprocmask", "rt_sigreturn",

				// Process/thread operations
				"exit", "exit_group", "nanosleep", "clock_nanosleep",
				/* "getpid", "getuid", "geteuid", "getgid", "getegid",
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
			*/
		}
	}

	// Create a new filter
	filter := seccomp.Filter{
		NoNewPrivs: true,
		Flag:       seccomp.FilterFlagTSync,
		Policy: seccomp.Policy{
			DefaultAction: ParseSeccompAction(config.Action),
			Syscalls: []seccomp.SyscallGroup{
				{
					Action: seccomp.ActionAllow,
					Names:  allowedSyscalls,
				},
			},
		},
	}

	// Log seccomp initialization details
	fmt.Printf("\033[2;37mInitializing seccomp with profile: %s, action: %s\033[0m\n", config.Profile, config.Action)
	// DEBUG: fmt.Printf("Allowing %d syscalls\n", len(allowedSyscalls))

	// Load the filter
	if err := seccomp.LoadFilter(filter); err != nil {
		return fmt.Errorf("failed to load seccomp filter: %w", err)
	}

	// DEBUG fmt.Printf("Seccomp filter loaded successfully\n")

	// Set the global CurrentSeccompProfile variable
	CurrentSeccompProfile = config.Profile

	// Set an environment variable that can be checked by builtins
	os.Setenv("RYE_SECCOMP_PROFILE", config.Profile)

	return nil
}

// DisableSeccompForDebug can be called to disable seccomp for debugging
func DisableSeccompForDebug() {
	// This function is intentionally empty in the release build
}

// SetupSeccompTrapHandler sets up a handler for SIGSYS signals that are triggered
// when using the "trap" action with seccomp. This helps identify which syscalls
// are being blocked and causing issues.
func SetupSeccompTrapHandler() {
	// This is a no-op in the release build
	// In a debug build, this would register a signal handler for SIGSYS
	fmt.Println("Seccomp trap handler would be set up here in debug mode")
}
