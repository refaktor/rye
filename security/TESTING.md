# Security Testing Framework for Rye

This directory contains comprehensive tests for Rye's security features, including:

1. **Landlock** - Linux filesystem access control
2. **Seccomp** - Linux system call filtering
3. **Code Signing** - Script authenticity verification

## Test Files

- `test_landlock_comprehensive.rye` - Tests for Landlock filesystem access control
- `test_seccomp_comprehensive.rye` - Tests for Seccomp system call filtering
- `test_codesig_comprehensive.rye` - Tests for code signing verification

## Running the Tests

### Landlock Tests

The Landlock tests verify that filesystem access control is working correctly. Run the tests with:

```bash
# Test readonly profile
rye -landlock -landlock-profile=readonly security/test_landlock_comprehensive.rye readonly

# Test readexec profile
rye -landlock -landlock-profile=readexec security/test_landlock_comprehensive.rye readexec

# Test custom profile
rye -landlock -landlock-profile=custom -landlock-paths=landlock_test/readable:r,landlock_test/writable:rw security/test_landlock_comprehensive.rye custom
```

The tests create a temporary directory structure with different types of files and directories, then attempt various operations (read, write, create, delete, execute) to verify that the Landlock profiles are enforcing the expected access controls.

### Seccomp Tests

The Seccomp tests verify that system call filtering is working correctly. Run the tests with:

```bash
# Test strict profile
rye -seccomp-profile=strict security/test_seccomp_comprehensive.rye strict

# Test readonly profile
rye -seccomp-profile=readonly security/test_seccomp_comprehensive.rye readonly
```

The tests attempt various operations (file operations, network operations, process operations, system operations) to verify that the Seccomp profiles are allowing or blocking the expected system calls.

### Code Signing Tests

The code signing tests verify that script authenticity verification is working correctly. Run the tests with:

```bash
# Run with code signing enabled
rye -codesig security/test_codesig_comprehensive.rye
```

The tests create temporary script files with valid signatures, invalid signatures, and no signatures, then verify that the code signing system correctly accepts or rejects them.

## Test Coverage

### Landlock Tests

- **File Read Operations**: Tests that read operations are allowed in all profiles
- **File Write Operations**: Tests that write operations are blocked in readonly and readexec profiles, but allowed in custom profile with write permission
- **File Creation**: Tests that file creation is blocked in readonly and readexec profiles, but allowed in custom profile with write permission
- **File Deletion**: Tests that file deletion is blocked in readonly and readexec profiles, but allowed in custom profile with write permission
- **File Execution**: Tests that file execution is blocked in readonly profile, allowed in readexec profile, and blocked in custom profile without execute permission

### Seccomp Tests

- **File Operations**: Tests that file read operations are allowed in both profiles, but write operations are only allowed in the strict profile
- **Network Operations**: Tests that network operations are allowed in both profiles
- **Process Operations**: Tests that process creation is blocked in both profiles, but process termination is allowed
- **System Operations**: Tests that system information retrieval is allowed in both profiles, but system configuration changes are blocked

### Code Signing Tests

- **Valid Signatures**: Tests that scripts with valid signatures from trusted keys are accepted
- **Invalid Signatures**: Tests that scripts with invalid signatures, malformed signatures, or no signatures are rejected
- **Tampered Scripts**: Tests that scripts where the content has been modified after signing are rejected

## Adding New Tests

To add new tests for security features:

1. Create a new test file in the `security` directory
2. Follow the pattern of the existing test files, using the error handling utilities and test functions
3. Add the new test file to this README

## Security Best Practices

When running these tests, keep in mind:

1. **Landlock and Seccomp are Linux-only features**: These tests will only work on Linux systems
2. **Code signing requires proper key management**: The tests simulate key verification, but in a real environment, you need to properly manage your keys
3. **Defense in depth**: Use multiple security features together for the best protection
4. **Test in isolation**: Run these tests in a controlled environment to avoid interfering with your system

## Troubleshooting

If the tests fail, check:

1. **Linux kernel version**: Landlock requires Linux 5.13+ and Seccomp requires Linux 3.5+
2. **Build tags**: Ensure Rye was built with the appropriate tags (`landlock`, `seccomp`)
3. **Permissions**: Code signing tests may fail if the `.codepks` file has incorrect permissions
4. **System capabilities**: Some tests may require specific capabilities or permissions
