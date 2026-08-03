module github.com/refaktor/rye/embed

go 1.26.1

// Development: replace directive points to the parent rye module in the repo.
// When using a released version, remove this replace and the require below
// will resolve to the tagged version from the Go module proxy.
replace github.com/refaktor/rye => ../

// The embed module only depends on rye/env, rye/evaldo, and rye/loader.
// OS I/O builtins and terminal display live in rye/baseio (not imported here).
// No build tags are needed - dependency isolation is at the package level.
//
// External dependencies pulled in transitively:
//   golang.org/x/text   – unicode case-folding (evaldo)
//   golang.org/x/crypto – PBKDF2 (util/securesave)

require github.com/refaktor/rye v0.0.0-00010101000000-000000000000

require (
	golang.org/x/crypto v0.52.0 // indirect
	golang.org/x/text v0.37.0 // indirect
)
