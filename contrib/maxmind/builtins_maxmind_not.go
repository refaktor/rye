//go:build !b_maxmind
// +build !b_maxmind

package maxmind

import (
	"github.com/refaktor/rye/env"
)

// Builtins_maxmind is empty when Rye is built without the b_maxmind tag.
var Builtins_maxmind = map[string]*env.Builtin{}
