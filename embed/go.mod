module github.com/refaktor/rye/embed

go 1.26.1

// Replace directive points to the parent rye module in the repository.
// When used by external projects, remove this replace and reference a
// specific version: github.com/refaktor/rye/embed v0.x.y
replace github.com/refaktor/rye => ../

require github.com/refaktor/rye v0.0.0-00010101000000-000000000000

require (
	github.com/clipperhouse/uax29/v2 v2.7.0 // indirect
	github.com/drewlanenga/govector v0.0.0-20220726163947-b958ac08bc93 // indirect
	github.com/elastic/go-seccomp-bpf v1.6.0 // indirect
	github.com/fsnotify/fsnotify v1.10.1 // indirect
	github.com/kopoli/go-terminal-size v0.0.0-20170219200355-5c97524c8b54 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/landlock-lsm/go-landlock v0.8.1 // indirect
	github.com/mattn/go-isatty v0.0.22 // indirect
	github.com/mattn/go-runewidth v0.0.24 // indirect
	github.com/pkg/term v1.2.0-beta.2.0.20211217091447-1a4a3b719465 // indirect
	github.com/refaktor/keyboard v0.0.0-20260517095250-755a59d30156 // indirect
	github.com/rogpeppe/go-internal v1.14.1 // indirect
	github.com/stretchr/testify v1.11.1 // indirect
	golang.org/x/crypto v0.52.0 // indirect
	golang.org/x/net v0.55.0 // indirect
	golang.org/x/sync v0.20.0 // indirect
	golang.org/x/sys v0.45.0 // indirect
	golang.org/x/term v0.43.0 // indirect
	golang.org/x/text v0.37.0 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	kernel.org/pub/linux/libs/security/libcap/psx v1.2.77 // indirect
)
