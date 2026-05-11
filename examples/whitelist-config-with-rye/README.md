# Config Blogpost Examples

This directory contains complete Go source code and config files for each step described in the "Whitelist-based Config files with Rye and Go" blog post.

## Structure

Each step is in its own directory with:
- `main.go` - Complete Go source code
- `config.rye` - Rye configuration file  
- Content directories with sample markdown files

## Steps

### Step 1: Minimal Server (`step1-minimal/`)
- Pure data notation config
- Basic markdown server
- No functions, just key-value pairs

### Step 2: Basic Computation (`step2-computation/`)
- Adds `*` and `+` operations
- Derived values in config
- Arithmetic expressions

### Step 3: Environment Variables (`step3-environment/`)
- Custom `get-env` builtin
- `any` combinator for fallbacks
- Environment variable support

### Step 4: Route Registration (`step4-routes/`)
- Custom `route` builtin
- Conditional logic with `if` and `=`
- Dynamic route configuration

### Step 5: User Functions (`step5-functions/`)
- `fn` for user-defined functions
- `replace` and `capitalize?` string operations
- Config-defined request handlers

### Step 6: Debugging (`step6-debugging/`)
- `probe` for value inspection
- `enter-console` for live REPL debugging
- Execution limits for safety

## Usage

For any step:

1. Navigate to the step directory:
   ```bash
   cd stepN-name/
   ```

2. Initialize Go module (if not done):
   ```bash
   go mod tidy
   ```

3. Run the server:
   ```bash
   go run main.go
   ```

4. Test with curl or browser:
   ```bash
   curl http://localhost:3000
   curl http://localhost:3000/blog/
   curl http://localhost:3000/docs/
   ```

5. For step 4+, test conditional routes:
   ```bash
   DEBUG=1 go run main.go  # Enables /drafts route
   ```

6. For step 6, enable debugging by uncommenting the `enter-console` line in config.rye

## Dependencies

All steps require:
- Go 1.21+
- github.com/refaktor/rye
- github.com/yuin/goldmark

Dependencies are managed in the root `go.mod` file.

## Configuration

Each step's `config.rye` demonstrates the capabilities available at that stage:

- **Step 1**: Static values only
- **Step 2**: Basic arithmetic  
- **Step 3**: Environment variables with fallbacks
- **Step 4**: Conditional routes based on env vars
- **Step 5**: Custom page title transformation functions
- **Step 6**: Debugging and introspection capabilities

The examples show the progressive enablement of features through explicit whitelisting of Rye builtins.