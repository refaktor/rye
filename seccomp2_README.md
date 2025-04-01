# Seccomp2 in Rye

Rye now supports a pure Go seccomp implementation using `github.com/elastic/go-seccomp-bpf` alongside the original C-based implementation.

## What is Seccomp?

Seccomp (secure computing mode) is a Linux kernel feature that restricts the system calls a process can make, enhancing security by limiting potential damage from compromised processes.

## Why a Pure Go Implementation?

- **No C Dependencies**: The original implementation (`-seccomp` flags) requires the libseccomp C library.
- **Simplified Deployment**: The new implementation (`-seccomp2` flags) is a pure Go solution with no external dependencies.
- **Easier Cross-Compilation**: Pure Go code is easier to cross-compile for different architectures.

## Using Seccomp2

Enable the pure Go seccomp implementation with these flags:

```
-seccomp2=true                   # Enable seccomp2 filtering (pure Go implementation)
-seccomp2-profile=strict         # Use the strict profile (only option available)
-seccomp2-action=errno           # Action for restricted syscalls (default)
```

### Simplified Design

The seccomp2 implementation is intentionally simpler than the original:
- Only supports a single `strict` profile with essential syscalls for Go programs
- Focuses on providing a minimal, secure set of allowed syscalls

### Available Actions

- `errno` (default): Return EPERM for disallowed syscalls
- `kill`: Terminate the process when a disallowed syscall is attempted
- `trap`: Send SIGSYS signal on disallowed syscalls (causes crashes with stack trace)
- `log`: Log disallowed syscalls but allow them (for debugging)

## Building with Seccomp Support

```
go build -tags seccomp
```

## Notes

- If both seccomp implementations are enabled, only seccomp2 will be used
- Seccomp2 is disabled by default and must be explicitly enabled
- The original implementation supports multiple profiles (`default`, `strict`, `web`, `io`, `readonly`, `cgo`) while seccomp2 only supports `strict`
