# Security Package

This package provides security features for Rye: 

* system call filtering *seccomp* (on Linux), 
* filesystem access control *landlock* (on Linux)
* code signing verification

## Unified Security Policy

All security features are configured through a policy system. Policies can come from:

1. **Embedded** (compiled into binary) - highest priority, cannot be bypassed
2. **System** (`/etc/rye/mandatory.yaml`) - must be root-owned
3. **Local** (`.ryesec` in script directory) - must be root-owned
4. **CLI flags** - lowest priority

### Policy File Format

```yaml
version: "1.0"
description: "My application security policy"

seccomp:
  enabled: true
  profile: strict      # strict, readonly
  action: kill         # errno, kill, trap, log

landlock:
  enabled: true
  profile: custom      # readonly, readexec, custom
  paths:
    - "/app:r"
    - "/data:rw"
    - "/tmp:rw"

codesig:
  enforced: true
  public_keys:         # Inline keys (hex-encoded Ed25519)
    - "827ba5f0..."
  public_keys_file: "/etc/rye/trusted_keys"  # Or load from file

mandatory: true        # Cannot be relaxed by CLI flags
```

### Security Requirements for Policy Files

Policy files (`.ryesec`, `/etc/rye/*.yaml`, key files) must be:
- **Owned by root** (uid 0)
- **Not writable by group or others** (mode 644 or stricter)

If these requirements are not met, Rye will **refuse to run**.

## Components

### Seccomp (System Call Filtering)

Restricts which system calls the process can make.

| Profile | Description |
|---------|-------------|
| `strict` | Minimal syscalls for Go runtime, blocks execve |
| `readonly` | Allows network, blocks file write syscalls |

| Action | Behavior |
|--------|----------|
| `errno` | Return EPERM (allows graceful handling) |
| `kill` | Terminate immediately (most secure) |
| `trap` | Send SIGSYS (debugging) |
| `log` | Log but allow (debugging only) |

### Landlock (Filesystem Access Control)

Restricts which files/directories the process can access.

| Profile | Description |
|---------|-------------|
| `readonly` | Read-only access to specified paths |
| `readexec` | Read and execute access |
| `custom` | Per-path permissions: `r` (read), `w` (write), `x` (execute) |

### Code Signing

Verifies that scripts are signed by trusted keys.

**Public keys can be specified:**
- Inline in the policy (`public_keys` array)
- In a separate file (`public_keys_file` - must be root-owned)

**Script signature format:**
```
; Your Rye script here
print "Hello"
;ryesig <hex-encoded-ed25519-signature>
```

## Usage

### Option 1: Embedded Binary (Recommended for Production)

```bash
# Create security policy
cat > security.yaml << 'EOF'
version: "1.0"
seccomp:
  enabled: true
  profile: strict
  action: kill
landlock:
  enabled: true
  profile: readonly
mandatory: true
EOF

# Generate embedded policy code
go run cmd/ryesecgen/main.go -input security.yaml

# Build with embedded policy
go build -tags "embed_security,seccomp,landlock" -o myapp
```

### Option 2: Per-Application Policy (.ryesec)

```bash
# Create policy in application directory
sudo cat > /opt/myapp/.ryesec << 'EOF'
version: "1.0"
seccomp:
  enabled: true
  profile: strict
  action: kill
mandatory: true
EOF

sudo chown root:root /opt/myapp/.ryesec
sudo chmod 644 /opt/myapp/.ryesec

# Run with policy auto-applied
rye /opt/myapp/main.rye
```

### Option 3: System-Wide Policy

```bash
sudo mkdir -p /etc/rye
sudo cat > /etc/rye/mandatory.yaml << 'EOF'
version: "1.0"
seccomp:
  enabled: true
  profile: readonly
  action: errno
mandatory: true
EOF

sudo chown root:root /etc/rye/mandatory.yaml
sudo chmod 644 /etc/rye/mandatory.yaml

# All Rye scripts now run with this policy
rye any_script.rye
```

### Option 4: CLI Flags (Development Only)

```bash
# These can be overridden by policy files
rye -seccomp-profile=strict -landlock script.rye
```

## Building with Security Support

```bash
# Full security support
go build -tags "seccomp,landlock"

# With embedded policy
go build -tags "embed_security,seccomp,landlock"

# Or use the build script
./build_secure -policy security.yaml -output myapp
```

## Signing Scripts

```rye
; Generate keys
ed25519-generate-keys |set! { pub priv }

; Save public key (add to policy)
pub |encode-to\hex |print

; Sign a script
script-content: read %myscript.rye
signature: script-content priv |sign |encode-to\hex
; Append ";ryesig " + signature to the script
```
