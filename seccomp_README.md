# Seccomp in Rye

Rye now uses a pure Go seccomp implementation using `github.com/elastic/go-seccomp-bpf`.

## What is Seccomp?

Seccomp (secure computing mode) is a Linux kernel feature that restricts the system calls a process can make, enhancing security by limiting potential damage from compromised processes.

## Benefits of Pure Go Implementation

- **No C Dependencies**: No need for the libseccomp C library
- **Simplified Deployment**: Pure Go solution with no external dependencies
- **Easier Cross-Compilation**: Pure Go code is easier to cross-compile for different architectures

## Using Seccomp

Control seccomp filtering with these flags:

```
-seccomp=true                    # Enable seccomp filtering (enabled by default)
-seccomp-profile=strict          # Use the strict profile (only option available)
-seccomp-action=errno            # Action for restricted syscalls (default)
```

### Available Profiles

The seccomp implementation supports two profiles:

1. **strict** (default): Allows essential syscalls for Go programs, including read and write operations
   - Provides a minimal, secure set of allowed syscalls
   - Blocks dangerous syscalls like `execve` that could be used to execute external commands

2. **readonly**: Similar to strict but blocks write operations
   - Allows read operations but blocks write operations to files
   - Useful for running scripts that should only read from the filesystem
   - Provides an additional layer of security for untrusted scripts

### Available Actions

- `errno` (default): Return EPERM for disallowed syscalls
- `kill`: Terminate the process when a disallowed syscall is attempted
- `trap`: Send SIGSYS signal on disallowed syscalls (causes crashes with stack trace)
- `log`: Log disallowed syscalls but allow them (for debugging)

## Building with Seccomp Support

```
go build -tags seccomp
```

## Implementation Details

The seccomp implementation uses the `github.com/elastic/go-seccomp-bpf` library, which provides a pure Go interface to the Linux seccomp BPF system. This allows for seccomp filtering without requiring any C dependencies, making it easier to build and deploy Rye across different Linux environments.
