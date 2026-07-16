module github.com/refaktor/rye/embed

go 1.26.1

// Replace directive points to the parent rye module in the repository.
// When used by external projects, remove this replace and reference a
// specific version: github.com/refaktor/rye v0.x.y
replace github.com/refaktor/rye => ../

// Build with: -tags=no_baseio,no_vector
// This drops terminal / OS / filesystem / vector dependencies.
//
// To regenerate vendor after source changes:
//   cd embed && GOFLAGS="-tags=no_baseio,no_vector" go mod vendor
//   (then prune vendor/ to only golang.org/x/crypto and golang.org/x/text)
//
// Do NOT run plain `go mod tidy` — it is tag-unaware and will re-add heavy deps.

require github.com/refaktor/rye v0.0.0-00010101000000-000000000000

require (
	golang.org/x/crypto v0.52.0 // indirect
	golang.org/x/text v0.37.0 // indirect
)
