# ryesig - Rye Script Signing Tool

A command-line tool for signing and verifying Rye scripts using Ed25519 signatures.

## Overview

`ryesig` is a tool for managing code signatures for Rye scripts. It provides functionality to:

- Generate Ed25519 key pairs for signing
- Sign Rye scripts with a private key
- Verify Rye scripts using a public key
- Check if a script has a signature

This tool is designed to work with Rye's built-in code signing verification system, which can automatically verify signatures when running scripts.

## Installation

The `ryesig` tool is included in the Rye repository under `cmd/ryesig/`. To use it:

```bash
# Run directly
rye cmd/ryesig/main.rye [options] [script-path]

# Or build it as an executable
rye build cmd/ryesig/main.rye -o ryesig
./ryesig [options] [script-path]
```

## Usage

### Generate a Key Pair

```bash
rye cmd/ryesig/main.rye --generate [output-path]
```

This generates a new Ed25519 key pair and saves them to `[output-path].pub` and `[output-path].priv`. If no output path is specified, it defaults to `keys`.

Example:
```bash
rye cmd/ryesig/main.rye --generate my_keys
# Creates my_keys.pub and my_keys.priv
```

### Sign a Script

```bash
rye cmd/ryesig/main.rye --sign [private-key-path] [script-path]
```

This signs the script at `[script-path]` using the private key at `[private-key-path]`. The signature is added to the end of the script file.

Example:
```bash
rye cmd/ryesig/main.rye --sign my_keys.priv my_script.rye
```

### Verify a Script

```bash
rye cmd/ryesig/main.rye --verify [public-key-path] [script-path]
```

This verifies the signature of the script at `[script-path]` using the public key at `[public-key-path]`.

Example:
```bash
rye cmd/ryesig/main.rye --verify my_keys.pub my_script.rye
```

### Check if a Script Has a Signature

```bash
rye cmd/ryesig/main.rye --check [script-path]
```

This checks if the script at `[script-path]` has a signature. It does not verify the signature.

Example:
```bash
rye cmd/ryesig/main.rye --check my_script.rye
```

## Options

- `--help`, `-h`: Show help message
- `--generate`, `-g [path]`: Generate a new Ed25519 key pair and save to `[path]`
- `--sign`, `-s [key-path]`: Sign the script using the private key at `[key-path]`
- `--verify`, `-v [key-path]`: Verify the script using the public key at `[key-path]`
- `--check`, `-c`: Check if the script has a valid signature
- `--output`, `-o [path]`: Output path for generated files (default: same as input)
- `--force`, `-f`: Force overwrite of existing files

## Signature Format

The signature is added to the end of the script file in the following format:

```
;ryesig <hex-encoded-signature>
```

This format is compatible with Rye's built-in code signing verification system.

## Security Considerations

- Keep your private key secure. Anyone with access to your private key can sign scripts that will be verified as coming from you.
- The public key can be shared freely and should be distributed to users who need to verify your scripts.
- For system-wide script verification, place the public key in a `.codepks` file in the same directory as the script.

## Integration with Rye's Code Signing System

This tool is designed to work with Rye's built-in code signing verification system. When a script is run, Rye can automatically verify its signature if:

1. The `--codesig` flag is used when running the script, or
2. A `.codepks` file containing trusted public keys is present in the script's directory

For more information on Rye's code signing system, see the [security documentation](../../security/README.md).
