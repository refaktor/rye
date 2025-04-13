# Landlock in Rye

Rye now uses the `github.com/landlock-lsm/go-landlock` library to implement filesystem access control using the Linux Landlock LSM.

## What is Landlock?

Landlock is a Linux security module (LSM) that allows restricting file system access for processes. Unlike seccomp which restricts syscalls, Landlock specifically focuses on filesystem access control, providing fine-grained control over which files and directories a process can access and what operations it can perform on them.

## Benefits of Landlock

- **Fine-grained filesystem access control**: Restrict access to specific files and directories
- **Complementary to seccomp**: While seccomp restricts syscalls, Landlock restricts filesystem access
- **Defense in depth**: Use both seccomp and landlock for multiple layers of security
- **Custom access rights**: Define custom access rights for different paths

## Using Landlock

Control Landlock filtering with these flags:

```
-landlock                         # Enable landlock filesystem access control
-landlock-profile=readonly        # Use the readonly profile (default)
-landlock-profile=readexec        # Use the readexec profile (allows execution)
-landlock-profile=custom          # Use a custom profile with user-defined paths
-landlock-paths=/path1,/path2     # Comma-separated list of paths to allow access to
```

### Available Profiles

The Landlock implementation supports three profiles:

1. **readonly** (default): Allows read-only access to files and directories
   - Allows reading files and listing directories
   - Blocks write operations, file creation, deletion, etc.

2. **readexec**: Similar to readonly but also allows execution
   - Allows reading files, listing directories, and executing files
   - Useful for running scripts that need to execute other programs

3. **custom**: User-defined access rights for specific paths
   - Format: `/path:permissions` where permissions can be r (read), w (write), x (execute)
   - Example: `-landlock-paths=/home/user/data:rw,/tmp:rx`
   - Allows fine-grained control over specific directories

## Building with Landlock Support

```
go build -tags landlock
```

## Implementation Details

The Landlock implementation uses the `github.com/landlock-lsm/go-landlock` library, which provides a Go interface to the Linux Landlock LSM. This allows for filesystem access control without requiring any C dependencies.

## Using Landlock with Seccomp

Landlock and seccomp can be used together to provide comprehensive security for Rye scripts:

```
# Use both seccomp and landlock for maximum security
rye -seccomp-profile=strict -landlock -landlock-profile=readonly script.rye

# Use landlock alone for filesystem restrictions
rye -landlock -landlock-profile=custom -landlock-paths=/home/user/data:r,/tmp:rw script.rye
```

This combination provides defense in depth by restricting both syscalls (seccomp) and filesystem access (landlock).
