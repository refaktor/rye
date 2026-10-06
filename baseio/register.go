// Package baseio provides the OS-level I/O builtins for Rye (file access,
// shell commands, stdin, os.Args, os.Exit), HTTP and the interactive terminal
// display builtins (display, explore, _.., explore\custom, print\ssv, print\csv).
//
// This package is separate from evaldo so that the embed module does not
// pull in heavy OS/terminal dependencies.  Call baseio.Register(ps) after
// evaldo.RegisterBuiltins(ps) to get the full CLI runner feature set.
package baseio

import (
	"github.com/refaktor/rye/env"
	"github.com/refaktor/rye/evaldo"
)

// Register registers all base I/O builtins and the interactive terminal
// display builtins into the program state.  Call after evaldo.RegisterBuiltins.
func Register(ps *env.ProgramState) {
	evaldo.RegisterBuiltins2(builtins_baseio, ps, "baseio")
	evaldo.RegisterBuiltins2(builtins_printing_extra, ps, "baseio-printing")
	evaldo.RegisterBuiltins2(Builtins_io, ps, "io")
	evaldo.RegisterBuiltins2(Builtins_cmd, ps, "cmd")
	evaldo.RegisterBuiltins2(Builtins_http, ps, "http")
	evaldo.RegisterBuiltins2(Builtins_table_files, ps, "table-files")
	evaldo.RegisterBuiltinsInContext(Builtins_os, ps, "os")
}
