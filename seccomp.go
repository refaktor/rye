//go:build linux && seccomp
// +build linux,seccomp

// To build with seccomp support:
// go build -tags seccomp

package main

import (
	"fmt"
	"syscall"

	seccomp "github.com/seccomp/libseccomp-golang"
)

// SeccompConfig holds the configuration for seccomp filtering
type SeccompConfig struct {
	Enabled bool
	Profile string
	Action  string
}

// ParseSeccompAction converts string action to seccomp action
// Available actions:
// - "errno" (default): Return an error code (EPERM) for disallowed syscalls, allowing the program to continue
// - "kill": Terminate the process immediately when a disallowed syscall is attempted (process will appear to hang)
// - "trap": Send a SIGSYS signal to the process when a disallowed syscall is attempted (will cause crashes with stack trace)
// - "log": Log disallowed syscalls but allow them to proceed (for debugging only)
//
// For most use cases, "errno" is recommended as it provides the best balance of security and usability.
func ParseSeccompAction(action string) seccomp.ScmpAction {
	switch action {
	case "kill":
		return seccomp.ActKill
	case "trap":
		return seccomp.ActTrap
	case "log":
		return seccomp.ActLog
	case "errno", "":
		return seccomp.ActErrno.SetReturnCode(int16(syscall.EPERM))
	default:
		fmt.Printf("Warning: Unknown seccomp action '%s', using 'errno' instead\n", action)
		return seccomp.ActErrno.SetReturnCode(int16(syscall.EPERM))
	}
}

// isValidProfile checks if the specified profile is valid
func isValidProfile(profile string) bool {
	validProfiles := []string{"default", "strict", "web", "io", "readonly", "cgo"}
	for _, p := range validProfiles {
		if profile == p {
			return true
		}
	}
	return false
}

// getProfileSyscalls returns the list of syscalls for the specified profile
func getProfileSyscalls(profile string) []string {
	switch profile {
	case "strict":
		return strictProfileSyscalls
	case "web":
		return webProfileSyscalls
	case "io":
		return ioProfileSyscalls
	case "readonly":
		return readonlyProfileSyscalls
	case "cgo":
		return cgoProfileSyscalls
	default:
		return defaultProfileSyscalls
	}
}

// strictProfileSyscalls is a minimal set of syscalls for high security
// This includes syscalls needed for the Go runtime and CGO
var strictProfileSyscalls = []string{
	// Essential syscalls for basic operation
	"read", "write", "open", "close", "stat", "fstat", "lstat",
	"poll", "lseek", "mmap", "mprotect", "munmap", "brk",
	"rt_sigaction", "rt_sigprocmask", "rt_sigreturn",
	"exit", "exit_group", "nanosleep", "clock_nanosleep",
	"getpid", "getuid", "geteuid", "getgid", "getegid",
	"getcwd", "chdir", "fchdir", "readlink",
	"access", "pipe", "pipe2", "dup", "dup2", "fcntl", "select",
	"getrlimit", "getrusage", "clock_gettime", "gettimeofday",
	"futex", "futex_waitv", "sched_yield", "getrandom",

	// Epoll-related syscalls for network polling
	"epoll_create", "epoll_create1", "epoll_ctl", "epoll_wait", "epoll_pwait", "epoll_pwait2",

	// Thread-related syscalls for Go runtime and CGO
	"clone", "clone3", "set_robust_list", "set_tid_address", "gettid", "tgkill",
	"sched_getaffinity", "sched_setaffinity", "sched_getparam", "sched_setparam",
	"sched_getscheduler", "sched_setscheduler", "sched_get_priority_max",
	"sched_get_priority_min", "sched_rr_get_interval", "sigaltstack",
	"rt_sigpending", "rt_sigtimedwait", "rt_sigqueueinfo", "rt_sigsuspend",

	// Memory management for Go runtime
	"mremap", "msync", "mincore", "madvise", "mlock", "munlock", "mlockall", "munlockall",

	// Process management
	"wait4", "kill", "uname", "prctl", "arch_prctl", "ptrace", "restart_syscall",
	"process_vm_readv", "process_vm_writev", "capget", "capset",

	// IPC
	"shmget", "shmat", "shmctl", "shmdt", "semget", "semop", "semctl",
	"msgget", "msgsnd", "msgrcv", "msgctl",

	// File system operations
	"statfs", "fstatfs", "getdents", "getdents64", "readv", "writev",
	"pread64", "pwrite64", "fadvise64", "openat",

	// Time-related
	"timer_create", "timer_settime", "timer_gettime", "timer_getoverrun",
	"timer_delete", "clock_settime", "clock_getres",

	// Socket operations (needed for some Go runtime features)
	"socket", "connect", "accept", "accept4", "bind", "listen",
	"getsockname", "getpeername", "socketpair", "setsockopt", "getsockopt",
	"shutdown", "recvfrom", "sendto", "recvmsg", "sendmsg",

	// Miscellaneous
	"ioctl", "eventfd", "eventfd2", "signalfd", "signalfd4",
	"timerfd_create", "timerfd_settime", "timerfd_gettime",
	"unshare", "setns", "get_robust_list", "splice", "tee", "sync_file_range",
	"vmsplice", "move_pages", "getcpu", "rseq",
}

// webProfileSyscalls is optimized for network operations
var webProfileSyscalls = []string{
	// Include all strict syscalls
	"read", "write", "open", "close", "stat", "fstat", "lstat",
	"poll", "lseek", "mmap", "mprotect", "munmap", "brk",
	"rt_sigaction", "rt_sigprocmask", "rt_sigreturn",
	"exit", "exit_group", "nanosleep", "clock_nanosleep",
	"getpid", "getuid", "geteuid", "getgid", "getegid",
	"getcwd", "chdir", "fchdir", "readlink",
	"access", "pipe", "pipe2", "dup", "dup2", "fcntl", "select",
	"getrlimit", "getrusage", "clock_gettime", "gettimeofday",
	"futex", "futex_waitv", "sched_yield", "getrandom",

	// Thread-related syscalls for Go runtime and CGO
	"clone", "clone3", "set_robust_list", "set_tid_address", "gettid", "tgkill",

	// Network-specific syscalls
	"socket", "connect", "accept", "accept4", "bind", "listen",
	"getsockname", "getpeername", "socketpair", "setsockopt", "getsockopt",
	"shutdown", "sendto", "recvfrom", "sendmsg", "recvmsg",
	"sendmmsg", "recvmmsg", "epoll_create", "epoll_create1",
	"epoll_ctl", "epoll_wait", "epoll_pwait", "epoll_pwait2", "rseq",
}

// ioProfileSyscalls is optimized for file I/O operations
var ioProfileSyscalls = []string{
	// Include all strict syscalls
	"read", "write", "open", "close", "stat", "fstat", "lstat",
	"poll", "lseek", "mmap", "mprotect", "munmap", "brk",
	"rt_sigaction", "rt_sigprocmask", "rt_sigreturn",
	"exit", "exit_group", "nanosleep", "clock_nanosleep",
	"getpid", "getuid", "geteuid", "getgid", "getegid",
	"getcwd", "chdir", "fchdir", "readlink",
	"access", "pipe", "pipe2", "dup", "dup2", "fcntl", "select",
	"getrlimit", "getrusage", "clock_gettime", "gettimeofday",
	"futex", "futex_waitv", "sched_yield", "getrandom",

	// Thread-related syscalls for Go runtime and CGO
	"clone", "clone3", "set_robust_list", "set_tid_address", "gettid", "tgkill",

	// File I/O specific syscalls
	"pread64", "pwrite64", "readv", "writev", "fsync", "fdatasync",
	"truncate", "ftruncate", "getdents", "getdents64", "mkdir",
	"rmdir", "creat", "link", "unlink", "symlink", "rename",
	"chmod", "fchmod", "chown", "fchown", "lchown", "umask",
	"openat", "mkdirat", "mknodat", "fchownat", "futimesat",
	"newfstatat", "unlinkat", "renameat", "linkat", "symlinkat",
	"readlinkat", "fchmodat", "faccessat", "utimensat", "rseq",
}

// readonlyProfileSyscalls blocks file write operations but allows stdout/stderr
var readonlyProfileSyscalls = []string{
	// Basic process operations
	"read", "write", "close", "stat", "fstat", "lstat",
	"poll", "lseek", "mmap", "mprotect", "munmap", "brk",
	"rt_sigaction", "rt_sigprocmask", "rt_sigreturn", "ioctl",
	"access", "pipe", "pipe2", "select", "sched_yield", "mremap", "msync",
	"mincore", "madvise", "shmget", "shmat", "shmctl", "dup",
	"dup2", "pause", "nanosleep", "getitimer", "alarm", "setitimer",
	"getpid", "socket", "connect", "accept", "sendto", "recvfrom",
	"sendmsg", "recvmsg", "shutdown", "bind", "listen", "getsockname",
	"getpeername", "socketpair", "setsockopt", "getsockopt", "clone", "clone3",
	"fork", "vfork", "execve", "exit", "wait4", "kill", "uname",
	"semget", "semop", "semctl", "shmdt", "msgget", "msgsnd",
	"msgrcv", "msgctl", "fcntl", "flock", "getdents", "getcwd",
	"chdir", "fchdir", "readlink", "gettimeofday", "getrlimit",
	"getrusage", "sysinfo", "times", "ptrace", "getuid", "syslog",
	"getgid", "setuid", "setgid", "geteuid", "getegid", "setpgid",
	"getppid", "getpgrp", "setsid", "setreuid", "setregid", "getgroups",
	"setgroups", "setresuid", "getresuid", "setresgid", "getresgid",
	"getpgid", "setfsuid", "setfsgid", "getsid", "capget", "capset",
	"rt_sigpending", "rt_sigtimedwait", "rt_sigqueueinfo", "rt_sigsuspend",
	"sigaltstack", "ustat", "statfs", "fstatfs", "sysfs", "getpriority",
	"setpriority", "sched_setparam", "sched_getparam", "sched_setscheduler",
	"sched_getscheduler", "sched_get_priority_max", "sched_get_priority_min",
	"sched_rr_get_interval", "mlock", "munlock", "mlockall", "munlockall",
	"vhangup", "modify_ldt", "pivot_root", "_sysctl", "prctl", "arch_prctl",
	"adjtimex", "setrlimit", "chroot", "sync", "acct", "settimeofday",
	"sethostname", "setdomainname", "iopl", "ioperm", "create_module",
	"init_module", "delete_module", "get_kernel_syms", "query_module",
	"quotactl", "nfsservctl", "getpmsg", "putpmsg", "afs_syscall",
	"tuxcall", "security", "gettid", "readahead", "getxattr", "lgetxattr",
	"fgetxattr", "listxattr", "llistxattr", "flistxattr", "tkill", "time",
	"futex", "futex_waitv", "sched_setaffinity", "sched_getaffinity", "set_thread_area",
	"io_setup", "io_destroy", "io_getevents", "io_submit", "io_cancel",
	"get_thread_area", "lookup_dcookie", "epoll_create", "epoll_ctl_old",
	"epoll_wait_old", "remap_file_pages", "getdents64", "set_tid_address",
	"restart_syscall", "semtimedop", "fadvise64", "timer_create",
	"timer_settime", "timer_gettime", "timer_getoverrun", "timer_delete",
	"clock_settime", "clock_gettime", "clock_getres", "clock_nanosleep",
	"exit_group", "epoll_wait", "epoll_ctl", "tgkill", "utimes", "vserver",
	"mbind", "set_mempolicy", "get_mempolicy", "mq_open", "mq_unlink",
	"mq_timedsend", "mq_timedreceive", "mq_notify", "mq_getsetattr",
	"kexec_load", "waitid", "add_key", "request_key", "keyctl",
	"ioprio_set", "ioprio_get", "inotify_init", "inotify_add_watch",
	"inotify_rm_watch", "migrate_pages", "pselect6", "ppoll", "unshare",
	"set_robust_list", "get_robust_list", "splice", "tee", "sync_file_range",
	"vmsplice", "move_pages", "epoll_pwait", "signalfd", "timerfd_create",
	"eventfd", "timerfd_settime", "timerfd_gettime", "accept4", "signalfd4",
	"eventfd2", "epoll_create1", "dup3", "pipe2", "inotify_init1", "preadv",
	"rt_tgsigqueueinfo", "perf_event_open", "recvmmsg", "fanotify_init",
	"fanotify_mark", "prlimit64", "name_to_handle_at", "open_by_handle_at",
	"clock_adjtime", "syncfs", "sendmmsg", "setns", "getcpu",
	"process_vm_readv", "kcmp", "finit_module", "sched_setattr",
	"sched_getattr", "seccomp", "getrandom", "bpf", "execveat",
	"userfaultfd", "membarrier", "copy_file_range", "preadv2",
	"pkey_mprotect", "pkey_alloc", "pkey_free", "statx", "io_pgetevents",
	"rseq", "pidfd_send_signal", "io_uring_setup", "io_uring_enter",
	"io_uring_register", "open_tree", "move_mount", "fsopen", "fsconfig",
	"fsmount", "fspick", "pidfd_open", "clone3", "close_range", "openat2",
	"pidfd_getfd", "faccessat2", "process_madvise", "epoll_pwait2",
	"mount_setattr", "quotactl_fd", "landlock_create_ruleset",
	"landlock_add_rule", "landlock_restrict_self", "memfd_secret",
	"process_mrelease", "futex_waitv", "set_mempolicy_home_node",

	// Allow open with O_RDONLY
	"open", "openat",

	// Allow read operations
	"pread64", "readv",
}

// cgoProfileSyscalls is a profile that allows all syscalls needed for CGO
// but blocks some potentially dangerous operations
var cgoProfileSyscalls = []string{
	// Include all syscalls from the default profile except for dangerous ones
	// This ensures that CGO works correctly while still providing some security
	"read", "write", "open", "close", "stat", "fstat", "lstat",
	"poll", "lseek", "mmap", "mprotect", "munmap", "brk",
	"rt_sigaction", "rt_sigprocmask", "rt_sigreturn", "ioctl",
	"pread64", "pwrite64", "readv", "writev", "access", "pipe",
	"select", "sched_yield", "mremap", "msync", "mincore", "madvise",
	"shmget", "shmat", "shmctl", "dup", "dup2", "pause", "nanosleep",
	"getitimer", "alarm", "setitimer", "getpid", "sendfile",
	"clone", "fork", "vfork", "execve", "exit", "wait4", "kill", "uname",
	"semget", "semop", "semctl", "shmdt", "msgget", "msgsnd", "msgrcv", "msgctl",
	"fcntl", "flock", "fsync", "fdatasync", "truncate", "ftruncate", "getdents",
	"getcwd", "chdir", "fchdir", "rename", "mkdir", "rmdir", "creat",
	"link", "unlink", "symlink", "readlink", "chmod", "fchmod", "chown",
	"fchown", "lchown", "umask", "gettimeofday", "getrlimit", "getrusage",
	"sysinfo", "times", "ptrace", "getuid", "syslog", "getgid", "setuid",
	"setgid", "geteuid", "getegid", "setpgid", "getppid", "getpgrp",
	"setsid", "setreuid", "setregid", "getgroups", "setgroups", "setresuid",
	"getresuid", "setresgid", "getresgid", "getpgid", "setfsuid", "setfsgid",
	"getsid", "capget", "capset", "rt_sigpending", "rt_sigtimedwait",
	"rt_sigqueueinfo", "rt_sigsuspend", "sigaltstack", "utime", "mknod",
	"uselib", "personality", "ustat", "statfs", "fstatfs", "sysfs",
	"getpriority", "setpriority", "sched_setparam", "sched_getparam",
	"sched_setscheduler", "sched_getscheduler", "sched_get_priority_max",
	"sched_get_priority_min", "sched_rr_get_interval", "mlock", "munlock",
	"mlockall", "munlockall", "vhangup", "modify_ldt", "pivot_root", "_sysctl",
	"prctl", "arch_prctl", "adjtimex", "setrlimit", "chroot", "sync", "acct",
	"settimeofday", "iopl", "ioperm", "quotactl", "getpmsg", "putpmsg",
	"afs_syscall", "tuxcall", "security", "gettid", "readahead", "setxattr",
	"lsetxattr", "fsetxattr", "getxattr", "lgetxattr", "fgetxattr", "listxattr",
	"llistxattr", "flistxattr", "removexattr", "lremovexattr", "fremovexattr",
	"tkill", "time", "futex", "sched_setaffinity", "sched_getaffinity",
	"set_thread_area", "io_setup", "io_destroy", "io_getevents", "io_submit",
	"io_cancel", "get_thread_area", "lookup_dcookie", "epoll_create",
	"epoll_ctl_old", "epoll_wait_old", "remap_file_pages", "getdents64",
	"set_tid_address", "restart_syscall", "semtimedop", "fadvise64",
	"timer_create", "timer_settime", "timer_gettime", "timer_getoverrun",
	"timer_delete", "clock_settime", "clock_gettime", "clock_getres",
	"clock_nanosleep", "exit_group", "epoll_wait", "epoll_ctl", "tgkill",
	"utimes", "vserver", "mbind", "set_mempolicy", "get_mempolicy", "mq_open",
	"mq_unlink", "mq_timedsend", "mq_timedreceive", "mq_notify",
	"mq_getsetattr", "kexec_load", "waitid", "add_key", "request_key",
	"keyctl", "ioprio_set", "ioprio_get", "inotify_init", "inotify_add_watch",
	"inotify_rm_watch", "migrate_pages", "openat", "mkdirat", "mknodat",
	"fchownat", "futimesat", "newfstatat", "unlinkat", "renameat", "linkat",
	"symlinkat", "readlinkat", "fchmodat", "faccessat", "pselect6", "ppoll",
	"unshare", "set_robust_list", "get_robust_list", "splice", "tee",
	"sync_file_range", "vmsplice", "move_pages", "utimensat", "epoll_pwait",
	"signalfd", "timerfd_create", "eventfd", "fallocate", "timerfd_settime",
	"timerfd_gettime", "accept4", "signalfd4", "eventfd2", "epoll_create1",
	"dup3", "pipe2", "inotify_init1", "preadv", "pwritev", "rt_tgsigqueueinfo",
	"perf_event_open", "recvmmsg", "fanotify_init", "fanotify_mark",
	"prlimit64", "name_to_handle_at", "open_by_handle_at", "clock_adjtime",
	"syncfs", "sendmmsg", "setns", "getcpu", "process_vm_readv",
	"process_vm_writev", "kcmp", "finit_module", "sched_setattr",
	"sched_getattr", "renameat2", "seccomp", "getrandom", "memfd_create",
	"kexec_file_load", "bpf", "execveat", "userfaultfd", "membarrier",
	"mlock2", "copy_file_range", "preadv2", "pwritev2", "pkey_mprotect",
	"pkey_alloc", "pkey_free", "statx", "io_pgetevents", "rseq",
	"pidfd_send_signal", "io_uring_setup", "io_uring_enter",
	"io_uring_register", "open_tree", "move_mount", "fsopen", "fsconfig",
	"fsmount", "fspick", "pidfd_open", "clone3", "close_range", "openat2",
	"pidfd_getfd", "faccessat2", "process_madvise", "epoll_pwait2",
	"mount_setattr", "quotactl_fd", "landlock_create_ruleset",
	"landlock_add_rule", "landlock_restrict_self", "memfd_secret",
	"process_mrelease", "futex_waitv", "set_mempolicy_home_node",

	// Socket operations needed for local communication
	"socket", "connect", "accept", "accept4", "bind", "listen",
	"getsockname", "getpeername", "socketpair", "setsockopt", "getsockopt",
	"shutdown", "recvfrom", "sendto", "recvmsg", "sendmsg",
	"recvmmsg", "sendmmsg",
}

// defaultProfileSyscalls is the default list of syscalls allowed by Rye
var defaultProfileSyscalls = []string{
	"read", "write", "open", "close", "stat", "fstat", "lstat",
	"poll", "lseek", "mmap", "mprotect", "munmap", "brk",
	"rt_sigaction", "rt_sigprocmask", "rt_sigreturn", "ioctl",
	"pread64", "pwrite64", "readv", "writev", "access", "pipe",
	"select", "sched_yield", "mremap", "msync", "mincore", "madvise",
	"shmget", "shmat", "shmctl", "dup", "dup2", "pause", "nanosleep",
	"getitimer", "alarm", "setitimer", "getpid", "sendfile", "socket",
	"connect", "accept", "sendto", "recvfrom", "sendmsg", "recvmsg",
	"shutdown", "bind", "listen", "getsockname", "getpeername",
	"socketpair", "setsockopt", "getsockopt", "clone", "fork", "vfork",
	"execve", "exit", "wait4", "kill", "uname", "semget", "semop",
	"semctl", "shmdt", "msgget", "msgsnd", "msgrcv", "msgctl", "fcntl",
	"flock", "fsync", "fdatasync", "truncate", "ftruncate", "getdents",
	"getcwd", "chdir", "fchdir", "rename", "mkdir", "rmdir", "creat",
	"link", "unlink", "symlink", "readlink", "chmod", "fchmod", "chown",
	"fchown", "lchown", "umask", "gettimeofday", "getrlimit", "getrusage",
	"sysinfo", "times", "ptrace", "getuid", "syslog", "getgid", "setuid",
	"setgid", "geteuid", "getegid", "setpgid", "getppid", "getpgrp",
	"setsid", "setreuid", "setregid", "getgroups", "setgroups", "setresuid",
	"getresuid", "setresgid", "getresgid", "getpgid", "setfsuid", "setfsgid",
	"getsid", "capget", "capset", "rt_sigpending", "rt_sigtimedwait",
	"rt_sigqueueinfo", "rt_sigsuspend", "sigaltstack", "utime", "mknod",
	"uselib", "personality", "ustat", "statfs", "fstatfs", "sysfs",
	"getpriority", "setpriority", "sched_setparam", "sched_getparam",
	"sched_setscheduler", "sched_getscheduler", "sched_get_priority_max",
	"sched_get_priority_min", "sched_rr_get_interval", "mlock", "munlock",
	"mlockall", "munlockall", "vhangup", "modify_ldt", "pivot_root", "_sysctl",
	"prctl", "arch_prctl", "adjtimex", "setrlimit", "chroot", "sync", "acct",
	"settimeofday", "mount", "umount2", "swapon", "swapoff", "reboot",
	"sethostname", "setdomainname", "iopl", "ioperm", "create_module",
	"init_module", "delete_module", "get_kernel_syms", "query_module",
	"quotactl", "nfsservctl", "getpmsg", "putpmsg", "afs_syscall", "tuxcall",
	"security", "gettid", "readahead", "setxattr", "lsetxattr", "fsetxattr",
	"getxattr", "lgetxattr", "fgetxattr", "listxattr", "llistxattr",
	"flistxattr", "removexattr", "lremovexattr", "fremovexattr", "tkill",
	"time", "futex", "sched_setaffinity", "sched_getaffinity",
	"set_thread_area", "io_setup", "io_destroy", "io_getevents", "io_submit",
	"io_cancel", "get_thread_area", "lookup_dcookie", "epoll_create",
	"epoll_ctl_old", "epoll_wait_old", "remap_file_pages", "getdents64",
	"set_tid_address", "restart_syscall", "semtimedop", "fadvise64",
	"timer_create", "timer_settime", "timer_gettime", "timer_getoverrun",
	"timer_delete", "clock_settime", "clock_gettime", "clock_getres",
	"clock_nanosleep", "exit_group", "epoll_wait", "epoll_ctl", "tgkill",
	"utimes", "vserver", "mbind", "set_mempolicy", "get_mempolicy", "mq_open",
	"mq_unlink", "mq_timedsend", "mq_timedreceive", "mq_notify",
	"mq_getsetattr", "kexec_load", "waitid", "add_key", "request_key",
	"keyctl", "ioprio_set", "ioprio_get", "inotify_init", "inotify_add_watch",
	"inotify_rm_watch", "migrate_pages", "openat", "mkdirat", "mknodat",
	"fchownat", "futimesat", "newfstatat", "unlinkat", "renameat", "linkat",
	"symlinkat", "readlinkat", "fchmodat", "faccessat", "pselect6", "ppoll",
	"unshare", "set_robust_list", "get_robust_list", "splice", "tee",
	"sync_file_range", "vmsplice", "move_pages", "utimensat", "epoll_pwait",
	"signalfd", "timerfd_create", "eventfd", "fallocate", "timerfd_settime",
	"timerfd_gettime", "accept4", "signalfd4", "eventfd2", "epoll_create1",
	"dup3", "pipe2", "inotify_init1", "preadv", "pwritev", "rt_tgsigqueueinfo",
	"perf_event_open", "recvmmsg", "fanotify_init", "fanotify_mark",
	"prlimit64", "name_to_handle_at", "open_by_handle_at", "clock_adjtime",
	"syncfs", "sendmmsg", "setns", "getcpu", "process_vm_readv",
	"process_vm_writev", "kcmp", "finit_module", "sched_setattr",
	"sched_getattr", "renameat2", "seccomp", "getrandom", "memfd_create",
	"kexec_file_load", "bpf", "execveat", "userfaultfd", "membarrier",
	"mlock2", "copy_file_range", "preadv2", "pwritev2", "pkey_mprotect",
	"pkey_alloc", "pkey_free", "statx", "io_pgetevents", "rseq",
	"pidfd_send_signal", "io_uring_setup", "io_uring_enter",
	"io_uring_register", "open_tree", "move_mount", "fsopen", "fsconfig",
	"fsmount", "fspick", "pidfd_open", "clone3", "close_range", "openat2",
	"pidfd_getfd", "faccessat2", "process_madvise", "epoll_pwait2",
	"mount_setattr", "quotactl_fd", "landlock_create_ruleset",
	"landlock_add_rule", "landlock_restrict_self", "memfd_secret",
	"process_mrelease", "futex_waitv", "set_mempolicy_home_node",
}

// InitSeccomp initializes the seccomp profile for Rye based on the provided configuration
func InitSeccomp(config SeccompConfig) error {
	// Skip if seccomp is disabled
	if !config.Enabled {
		return nil
	}

	// Check if the profile is valid
	if !isValidProfile(config.Profile) {
		return fmt.Errorf("invalid seccomp profile: %s (valid profiles: default, strict, web, io, readonly, cgo)", config.Profile)
	}

	// Create a new seccomp filter with the specified action
	action := ParseSeccompAction(config.Action)
	filter, err := seccomp.NewFilter(action)
	if err != nil {
		return fmt.Errorf("failed to create seccomp filter: %w", err)
	}

	// Get the syscalls list based on the selected profile
	syscalls := getProfileSyscalls(config.Profile)

	// Special handling for readonly profile
	if config.Profile == "readonly" {
		// Remove "write" from the syscalls list if it's there
		for i, syscallName := range syscalls {
			if syscallName == "write" {
				syscalls = append(syscalls[:i], syscalls[i+1:]...)
				break
			}
		}

		// Add a special rule for write that only allows writing to stdin/stdout/stderr
		writeID, err := seccomp.GetSyscallFromName("write")
		if err == nil {
			// Allow write to file descriptors 0 (stdin), 1 (stdout), and 2 (stderr)
			err = filter.AddRuleConditional(
				writeID,
				seccomp.ActAllow,
				[]seccomp.ScmpCondition{
					{
						Argument: 0,
						Op:       seccomp.CompareEqual,
						Operand1: 0, // stdin
						Operand2: 0,
					},
				},
			)
			if err != nil {
				return fmt.Errorf("failed to add conditional rule for write to stdin: %w", err)
			}

			err = filter.AddRuleConditional(
				writeID,
				seccomp.ActAllow,
				[]seccomp.ScmpCondition{
					{
						Argument: 0,
						Op:       seccomp.CompareEqual,
						Operand1: 1, // stdout
						Operand2: 0,
					},
				},
			)
			if err != nil {
				return fmt.Errorf("failed to add conditional rule for write to stdout: %w", err)
			}

			err = filter.AddRuleConditional(
				writeID,
				seccomp.ActAllow,
				[]seccomp.ScmpCondition{
					{
						Argument: 0,
						Op:       seccomp.CompareEqual,
						Operand1: 2, // stderr
						Operand2: 0,
					},
				},
			)
			if err != nil {
				return fmt.Errorf("failed to add conditional rule for write to stderr: %w", err)
			}
		}
	}

	// Add allowed syscalls to the filter
	for _, syscallName := range syscalls {
		syscallID, err := seccomp.GetSyscallFromName(syscallName)
		if err != nil {
			// Skip syscalls that don't exist on this architecture
			continue
		}
		err = filter.AddRule(syscallID, seccomp.ActAllow)
		if err != nil {
			return fmt.Errorf("failed to add rule for %s: %w", syscallName, err)
		}
	}

	// Load the filter into the kernel
	err = filter.Load()
	if err != nil {
		return fmt.Errorf("failed to load seccomp filter: %w", err)
	}

	return nil
}

// DisableSeccompForDebug can be called to disable seccomp for debugging
func DisableSeccompForDebug() {
	// This function is intentionally empty in the release build
	// It can be modified during development to disable seccomp
}
