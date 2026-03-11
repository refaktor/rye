//go:build linux

package runner

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"syscall"
)

// Environment variable names used to communicate between parent and child processes.
const (
	envUnshareChild = "RYE_UNSHARE_CHILD" // "1" when we are the sandboxed child
	envUnshareFs    = "RYE_UNSHARE_FS"    // "1" to set up filesystem jail in child
	envUnshareNet   = "RYE_UNSHARE_NET"   // "1" to isolate network in child
	envUnsharePid   = "RYE_UNSHARE_PID"   // "1" to isolate PID namespace in child
	envUnshareUts   = "RYE_UNSHARE_UTS"   // "1" to isolate UTS/hostname in child
)

// UnshareConfig holds the namespace isolation options collected from CLI flags
// or the security policy file.
type UnshareConfig struct {
	Fs  bool // Isolate filesystem (bind-mount CWD, pivot_root)
	Net bool // Isolate network namespace
	Pid bool // Isolate PID namespace
	Uts bool // Isolate UTS (hostname) namespace
}

// IsUnshareChild reports whether the current process is the sandboxed child
// spawned by a parent Rye process.
func IsUnshareChild() bool {
	return os.Getenv(envUnshareChild) == "1"
}

// ReadUnshareChildConfig returns the UnshareConfig that was passed to this
// child via environment variables. Only meaningful when IsUnshareChild() is true.
func ReadUnshareChildConfig() UnshareConfig {
	return UnshareConfig{
		Fs:  os.Getenv(envUnshareFs) == "1",
		Net: os.Getenv(envUnshareNet) == "1",
		Pid: os.Getenv(envUnsharePid) == "1",
		Uts: os.Getenv(envUnshareUts) == "1",
	}
}

// envBool converts a bool to the "1"/"0" string used for env vars.
func envBool(b bool) string {
	if b {
		return "1"
	}
	return "0"
}

// DoReexecInUnshare re-execs the current Rye binary with the requested Linux
// namespace clone flags. The parent process forwards stdin/stdout/stderr and
// waits for the child to exit, then exits with the child's exit code.
//
// This must only be called from the parent (i.e. when IsUnshareChild() is false).
func DoReexecInUnshare(cfg UnshareConfig) {
	// Build the clone flags. CLONE_NEWUSER is always required so that an
	// unprivileged user can create the other namespaces.
	var cloneFlags uintptr = syscall.CLONE_NEWUSER
	if cfg.Fs {
		cloneFlags |= syscall.CLONE_NEWNS
	}
	if cfg.Net {
		cloneFlags |= syscall.CLONE_NEWNET
	}
	if cfg.Pid {
		cloneFlags |= syscall.CLONE_NEWPID
	}
	if cfg.Uts {
		cloneFlags |= syscall.CLONE_NEWUTS
	}

	// Map current UID/GID to root inside the new user namespace so that
	// bind-mount and pivot_root calls succeed without real root privileges.
	uid := os.Getuid()
	gid := os.Getgid()

	cmd := exec.Command("/proc/self/exe", os.Args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: cloneFlags,
		UidMappings: []syscall.SysProcIDMap{
			{ContainerID: 0, HostID: uid, Size: 1},
		},
		GidMappings: []syscall.SysProcIDMap{
			{ContainerID: 0, HostID: gid, Size: 1},
		},
	}

	// Pass the configuration to the child via environment variables.
	cmd.Env = append(os.Environ(),
		envUnshareChild+"=1",
		envUnshareFs+"="+envBool(cfg.Fs),
		envUnshareNet+"="+envBool(cfg.Net),
		envUnsharePid+"="+envBool(cfg.Pid),
		envUnshareUts+"="+envBool(cfg.Uts),
	)

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		fmt.Fprintf(os.Stderr, "rye --unshare: failed to start sandbox: %v\n", err)
		os.Exit(1)
	}
	os.Exit(0)
}

// SetupUnshareFilesystem sets up the filesystem jail inside the child process.
// It:
//  1. Locks the goroutine to its OS thread (required for mount namespace changes).
//  2. Makes all existing mounts private so nothing leaks back to the host.
//  3. Creates a tmpfs jail in /tmp/rye_jail_<pid>.
//  4. Bind-mounts the current working directory (read-only) as /app inside the jail.
//  5. Performs pivot_root so / becomes the jail.
//  6. Chdir to /app so relative script paths continue to work.
//  7. Unmounts the old root.
//
// Must be called early in the child process before any interpreter state is set up.
func SetupUnshareFilesystem() error {
	// Pin this goroutine to one OS thread. Mount namespace changes are per-thread
	// in Linux; without this the Go scheduler could move us to a different thread
	// that still sees the host filesystem.
	runtime.LockOSThread()

	// Capture CWD before we start changing mounts.
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("unshare: getwd: %w", err)
	}
	// --- 1. Make all current mounts private ---
	// This prevents our bind-mounts from propagating back to the host.
	if err := syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, ""); err != nil {
		return fmt.Errorf("unshare: make mounts private: %w", err)
	}

	// --- 2. Create a tmpfs jail root in RAM ---
	jailRoot := fmt.Sprintf("/tmp/rye_jail_%s", strconv.Itoa(os.Getpid()))
	if err := os.MkdirAll(jailRoot, 0o755); err != nil {
		return fmt.Errorf("unshare: mkdir jail: %w", err)
	}
	if err := syscall.Mount("tmpfs", jailRoot, "tmpfs", 0, "size=64m,mode=0755"); err != nil {
		return fmt.Errorf("unshare: mount tmpfs: %w", err)
	}

	// --- 3. Bind-mount CWD as /app (read-only) ---
	appDir := filepath.Join(jailRoot, "app")
	if err := os.MkdirAll(appDir, 0o755); err != nil {
		return fmt.Errorf("unshare: mkdir app: %w", err)
	}
	if err := syscall.Mount(cwd, appDir, "", syscall.MS_BIND, ""); err != nil {
		return fmt.Errorf("unshare: bind-mount cwd: %w", err)
	}
	// Remount read-only.
	if err := syscall.Mount("none", appDir, "", syscall.MS_REMOUNT|syscall.MS_BIND|syscall.MS_RDONLY, ""); err != nil {
		// Non-fatal: we proceed with read-write if the remount fails (e.g. some kernels
		// require different flags). A warning is enough — the script still runs, just
		// with write access to its own directory.
		fmt.Fprintf(os.Stderr, "rye --unshare: warning: could not remount /app read-only: %v\n", err)
	}

	// --- 4. Prepare for pivot_root ---
	// pivot_root requires that the new root is a mount point. We bind-mount
	// jailRoot to itself recursively so that the app sub-mount is included
	// in the view seen after pivot_root.
	if err := syscall.Mount(jailRoot, jailRoot, "", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
		return fmt.Errorf("unshare: bind jail to itself: %w", err)
	}
	oldRoot := filepath.Join(jailRoot, ".old_root")
	if err := os.MkdirAll(oldRoot, 0o700); err != nil {
		return fmt.Errorf("unshare: mkdir old_root: %w", err)
	}

	// --- 5. Pivot root ---
	if err := syscall.PivotRoot(jailRoot, oldRoot); err != nil {
		return fmt.Errorf("unshare: pivot_root: %w", err)
	}
	if err := os.Chdir("/"); err != nil {
		return fmt.Errorf("unshare: chdir /: %w", err)
	}

	// --- 6. Unmount old root ---
	if err := syscall.Unmount("/.old_root", syscall.MNT_DETACH); err != nil {
		return fmt.Errorf("unshare: unmount old_root: %w", err)
	}
	// Best-effort removal of the mount-point directory.
	_ = os.Remove("/.old_root")

	// --- 7. Move into the project directory ---
	if err := os.Chdir("/app"); err != nil {
		return fmt.Errorf("unshare: chdir /app: %w", err)
	}

	return nil
}
