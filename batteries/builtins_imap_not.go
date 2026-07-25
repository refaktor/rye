//go:build no_imap
// +build no_imap

package batteries

import (
	"github.com/refaktor/rye/env"
)

var Builtins_imap = map[string]*env.Builtin{}
