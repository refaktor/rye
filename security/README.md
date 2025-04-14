# Security Package

This package contains security-related functionality for the Rye language, focusing on system call filtering, filesystem access control, and code signing verification.

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

### Code Signing

Code signing provides a way to verify that Rye scripts have not been tampered with and come from trusted sources. This package provides:

- `codesig.go`: Implementation of code signature verification for Rye

Code signing features:
- Verification of Ed25519 signatures in Rye scripts
- Loading trusted public keys from a `.codepks` file
- Integration with the loader package for script verification

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

// Initialize code signing with configuration from command-line flags
codeSigConfig := security.CodeSigConfig{
    Enabled: *runner.CodeSigEnabled,
    PubKeys: *runner.CodeSigPubKeys,
}

// Initialize seccomp
security.InitSeccomp(seccompConfig)

// Initialize landlock
security.InitLandlock(landlockConfig)

// Initialize code signing
security.InitCodeSig(codeSigConfig)
```

Command-line flags:
- `--seccomp-profile=strict|readonly`: Enable seccomp with the specified profile
- `--seccomp-action=errno|kill|trap|log`: Action to take on restricted syscalls
- `--landlock`: Enable landlock filesystem access control
- `--landlock-profile=readonly|readexec|custom`: Landlock profile to use
- `--landlock-paths=/path1,/path2`: Comma-separated list of paths to allow access to
- `--codesig`: Force code signature verification even if no .codepks file is present
- `--codesig-pubkeys=.codepks`: Path to the file containing trusted public keys

### Auto-Enforcement of Code Signing

Code signing is automatically enforced when a `.codepks` file is present in the same directory as the Rye script being executed. This provides "security by default" without requiring users to remember to add the `--codesig` flag.

Key aspects of this approach:
1. If a `.codepks` file exists in the script's directory, code signing is automatically enforced
2. The `.codepks` file must be owned by root and not writable by group or others
3. The `--codesig` flag can be used to force code signing even when no `.codepks` file is present
4. Each application/script directory can have its own `.codepks` file with specific trusted keys

## Code Signing Format

Rye scripts can be signed by adding a signature at the end of the file in the following format:

```
;ryesig <hex-encoded-signature>
```

The signature is an Ed25519 signature of the script content (excluding the signature line itself).

### Public Keys File Format and Security

The `.codepks` file should contain one public key per line in hexadecimal format. Lines starting with `#` are treated as comments and empty lines are ignored.

Example:
```
# Trusted public keys for Rye scripts
# Developer 1's key
827ba5f0904227678bf33446abbca8bf6a3a5333815920741eb475582a4c31dd
# Developer 2's key
a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2
```

#### Security Requirements

For security reasons, the `.codepks` file must meet the following requirements:

1. **Ownership**: On Unix-like systems, the file must be owned by the root user (uid 0)
2. **Permissions**: The file must not be writable by group or others (no write permission for anyone except the owner)

These requirements are enforced by the code signing system. If the file doesn't meet these requirements, code signing verification will fail with an appropriate error message.

To set the correct ownership and permissions, you can use the following commands:
```bash
# Change ownership to root
sudo chown root:root .codepks

# Set permissions to read-only for group and others
sudo chmod 644 .codepks
```

### Signing a Script

To sign a Rye script, you need to:
1. Generate an Ed25519 key pair using the `ed25519-generate-keys` function in Rye
2. Sign the script content with the private key using the `Ed25519-priv-key//sign` function
3. Add the signature to the end of the script in the format `;ryesig <hex-encoded-signature>`

Example Rye code to sign a script:
```
{ 
  script-content: "your script content here"
  ed25519-generate-keys |set! { pub priv }
  script-content priv |sign |encode-to\hex |print
  # Add the output as ;ryesig <signature> at the end of your script
}
```
