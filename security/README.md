# Security Package

This package contains security-related functionality for the Rye language, focusing on system call filtering and filesystem access control.

## Components

### Seccomp (Secure Computing Mode)

Seccomp is a Linux kernel feature that allows filtering of system calls to reduce the attack surface of applications. This package provides:

- `seccomp.go`: Implementation of seccomp filtering for Rye
- `seccomp_stub.go`: Stub implementation for non-Linux systems
- `seccomp_README.md`: Documentation for seccomp functionality

Seccomp profiles:
- `strict`: Minimal set of system calls required for basic operation
- `readonly`: Allows read operations but blocks write operations

### Landlock

Landlock is a Linux kernel feature that provides filesystem access control. This package provides:

- `landlock.go`: Implementation of landlock filesystem access control for Rye
- `landlock_stub.go`: Stub implementation for non-Linux systems
- `landlock_README.md`: Documentation for landlock functionality
- `test_landlock.rye`: Test script for landlock functionality

Landlock profiles:
- `readonly`: Allows read-only access to specified paths
- `readexec`: Allows read and execute access to specified paths
- `custom`: Custom access control based on path specifications

## Usage

These security features are initialized in `main.go` based on command-line flags:

```go
// Initialize seccomp with configuration from command-line flags
seccompConfig := security.SeccompConfig{
    Enabled: *runner.SeccompProfile != "",
    Profile: *runner.SeccompProfile,
    Action:  *runner.SeccompAction,
}

// Initialize landlock with configuration from command-line flags
landlockConfig := security.LandlockConfig{
    Enabled: *runner.LandlockEnabled,
    Profile: *runner.LandlockProfile,
    Paths:   strings.Split(*runner.LandlockPaths, ","),
}

// Initialize seccomp
security.InitSeccomp(seccompConfig)

// Initialize landlock
security.InitLandlock(landlockConfig)
```

Command-line flags:
- `--seccomp-profile=strict|readonly`: Enable seccomp with the specified profile
- `--seccomp-action=errno|kill|trap|log`: Action to take on restricted syscalls
- `--landlock`: Enable landlock filesystem access control
- `--landlock-profile=readonly|readexec|custom`: Landlock profile to use
- `--landlock-paths=/path1,/path2`: Comma-separated list of paths to allow access to
