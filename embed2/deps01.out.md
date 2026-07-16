Let me analyze the project structure and dependencies.Now I have the complete picture. Let me compile a thorough analysis.Now I have a thorough picture. Here's the complete analysis:

---

## Dependency Analysis for `rye/embed2`

### The Core Problem: Two Issues, Not One

**Issue 1 — Missing build tag enforcement in `embed2` itself.**
Every single unwanted dependency *is already guarded* in the rye source tree with `//go:build !no_baseio` or `//go:build !no_vector`. The stubs are in place. The problem is that `embed2` never *declares* or *enforces* these tags. When a user does `go get github.com/refaktor/rye/embed2` and builds without `-tags no_baseio,no_vector`, all the IO machinery compiles in.

**Issue 2 — Two files have no build guard at all and always compile.**
Even if you pass the tags correctly, these two slip through:

- `util/securesave.go` — no `//go:build` tag → always compiled → pulls `golang.org/x/crypto/pbkdf2` (AES-GCM file encryption, pure IO, zero place in an embedded evaluator)
- `security/policy.go` — no `//go:build` tag → always compiled → pulls `gopkg.in/yaml.v3` (reads `.ryesec` policy files from the filesystem)

---

### Full Dependency Map

| Dep in go.mod | Root file | Build guard exists? | Compiled without tags? | Compiled WITH tags? |
|---|---|---|---|---|
| `github.com/fsnotify/fsnotify` | `env/env_fsnotify.go` | ✅ `!wasm && !no_baseio` | ✅ yes | ❌ no |
| `github.com/drewlanenga/govector` | `env/object_vector.go` | ✅ `!no_vector` | ✅ yes | ❌ no |
| `github.com/kr/pretty` | transitive of govector test | ✅ (govector itself gated) | ✅ yes | ❌ no |
| `github.com/refaktor/keyboard` | `util/util_wasm_not.go` | ✅ `!wasm && !no_baseio` | ✅ yes | ❌ no |
| `github.com/kopoli/go-terminal-size` | `evaldo/builtins_base_printing.go` → `rye/term` | ✅ `!no_baseio` | ✅ yes | ❌ no |
| `github.com/mattn/go-runewidth` | `rye/term` | ✅ (via printing.go) | ✅ yes | ❌ no |
| `github.com/pkg/term` | `rye/term` | ✅ (via printing.go) | ✅ yes | ❌ no |
| `github.com/mattn/go-isatty` | `rye/term` (transitive) | ✅ (via printing.go) | ✅ yes | ❌ no |
| `github.com/clipperhouse/uax29/v2` | transitive of go-runewidth | ✅ (via printing.go) | ✅ yes | ❌ no |
| `golang.org/x/term` | `evaldo/builtins_baseio.go` | ✅ `!no_baseio` | ✅ yes | ❌ no |
| `golang.org/x/sync` | `evaldo/builtins_baseio.go` | ✅ `!no_baseio` | ✅ yes | ❌ no |
| `gopkg.in/yaml.v3` | **`security/policy.go`** | ❌ **NO TAG** | ✅ yes | ✅ **yes — leaks!** |
| `golang.org/x/crypto` | **`util/securesave.go`** | ❌ **NO TAG** | ✅ yes | ✅ **yes — leaks!** |
| `github.com/elastic/go-seccomp-bpf` | `security/seccomp.go` | ✅ `linux && seccomp` | ❌ no (linux+tag) | ❌ no |
| `github.com/landlock-lsm/go-landlock` | `security/landlock.go` | ✅ `linux && landlock` | ❌ no (linux+tag) | ❌ no |
| `kernel.org/pub/linux/libs/security/libcap/psx` | transitive of landlock | ✅ (via landlock) | ❌ no | ❌ no |
| `golang.org/x/net` | nowhere in core path | — (go.mod artifact) | ❌ no | ❌ no |

---

### What Needs to Change

#### Fix 1 — `util/securesave.go`: add a build guard (or split the file)

`SaveSecure`/`ReadSecure` are filesystem I/O. They belong to `no_baseio` territory. The `util` package is imported unconditionally by nearly every base builtin, so this file always compiles and always drags in `x/crypto`.

**Option A** – simplest: add the guard directly to `securesave.go`:
```go
//go:build !no_baseio
```
And create `securesave_stub.go` with `//go:build no_baseio` providing no-op or panic stubs for `SaveSecure`/`ReadSecure` (they aren't called from any base builtin anyway, only from `builtins_baseio.go` which is already gated).

**Option B** – cleaner: move `securesave.go` out of the `util` package into a new `util/secureio` sub-package that is only imported from `builtins_baseio.go`.

#### Fix 2 — `security/policy.go`: add a build guard

`policy.go` reads `.ryesec` YAML files from disk. It's meaningless in an embedded context (there's no "filesystem" from the host's perspective — that's the whole point of embed). The stub path already exists: `loader/codesig_stub.go` stubs out `checkCodeSignature` for `no_baseio`, so the `security` package is never *called* — but it still *compiles*.

Add `//go:build !no_baseio` to `policy.go`, `policy_unix.go`, `policy_windows.go`, and `codesig.go`, and provide a minimal `security_stub.go` with `//go:build no_baseio` that exposes whatever `VerifySignature` interface the loader needs (the stub already returns `1` in `codesig_stub.go` so likely nothing is needed from security at all).

#### Fix 3 — `embed2` must enforce its required tags

The go module system has no way to require build tags. There are two viable approaches:

**Option A — Mirror `embed` and use vendoring** (what `embed/` already does correctly):
- Run `GOFLAGS="-tags=no_baseio,no_vector" go mod vendor` — this produces a `vendor/` directory that only contains the tag-filtered source
- The `go.mod` then only lists `golang.org/x/text` (once the above two fixes are done, not even `x/crypto`)
- Downstream users building with `embed2` automatically get only what's in vendor

**Option B — Use a `//go:build` constraint on `embed2/rye.go` itself** to cause a compile error without the tags:
```go
//go:build no_baseio && no_vector
```
This makes the package uncompilable without the tags, forcing users to be explicit. It's the self-documenting approach — the package literally won't build without the right tags.

**Option C — `GOFLAGS` in a `go.env` file** in the repo root (not user-friendly for downstream).

---

### Ideal End State

After the three fixes, `embed2` with `-tags no_baseio,no_vector` (or via vendoring) would have **exactly these external deps**:
- `golang.org/x/text` — unicode case-folding in `builtins_base_strings.go` (legitimate, pure computation)
- *(nothing else)* — `x/crypto` eliminated by fixing `securesave.go`

The `golang.org/x/sys` entry stays in `go.mod` as a transitive requirement of `x/text` on some platforms but adds no code.