# Seccomp Integration for Rye

This document describes the seccomp integration in Rye, which provides an additional layer of security by restricting the system calls that the Rye runtime and applications can make.

## Overview

[Seccomp (Secure Computing Mode)](https://en.wikipedia.org/wiki/Seccomp) is a Linux kernel feature that allows a process to restrict the system calls it can make. By limiting the available system calls, seccomp reduces the attack surface of the application, making it more difficult for attackers to exploit vulnerabilities.

The seccomp integration in Rye is implemented at build time, which means the seccomp profile is embedded directly in the binary. This approach was chosen because:

1. It provides the strongest security guarantees - the seccomp profile cannot be bypassed at runtime
2. It requires no configuration from end users
3. It ensures consistent behavior across deployments
4. It reduces the attack surface from the start

## Building with Seccomp Support

To build Rye with seccomp support, you need to:

1. Install the libseccomp development package on your system:
   ```
   # Debian/Ubuntu
   sudo apt-get install libseccomp-dev

   # Fedora/RHEL
   sudo dnf install libseccomp-devel

   # Arch Linux
   sudo pacman -S libseccomp
   ```

2. Build Rye with the seccomp tag:
   ```
   go build -tags seccomp
   ```

If you build without the seccomp tag or on a non-Linux system, the seccomp integration will be a no-op (it won't do anything).

## Seccomp Profile

The seccomp profile in Rye allows a wide range of system calls that are necessary for normal operation, including:

- File operations (open, read, write, close, etc.)
- Network operations (socket, connect, bind, listen, etc.)
- Process management (fork, execve, wait4, etc.)
- Memory management (mmap, munmap, mprotect, etc.)
- Time-related operations (clock_gettime, nanosleep, etc.)
- And many others

System calls that are not explicitly allowed will be blocked, and the process will receive an EPERM error if it attempts to make such a call.

## Customizing the Seccomp Profile

If you need to customize the seccomp profile for your specific use case, you can modify the `seccomp.go` file. The profile is defined in the `InitSeccomp` function, which creates a seccomp filter and adds rules for allowed system calls.

To add or remove system calls from the profile, modify the `syscalls` slice in the `InitSeccomp` function.

## Debugging

If you encounter issues with the seccomp integration, you can:

1. Build without the seccomp tag to disable seccomp:
   ```
   go build
   ```

2. Use strace to see which system calls are being blocked:
   ```
   strace -f ./rye your_script.rye
   ```

3. Modify the `DisableSeccompForDebug` function in `seccomp.go` to disable seccomp at runtime for debugging purposes.

## Future Improvements

Potential future improvements to the seccomp integration include:

1. **Process Start Time Integration**: Add command-line flags or environment variables to enable/disable seccomp or select different profiles at runtime.

2. **Install Time Integration**: Store seccomp profiles in system-wide configuration directories with restricted permissions.

3. **Module-Specific Profiles**: Create different seccomp profiles for different Rye modules based on their required system calls.

4. **Capability-Based Approach**: Implement a capability-based approach where modules declare their required system calls.
