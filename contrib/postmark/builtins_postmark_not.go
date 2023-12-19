//go:build !b_postmark
// +build !b_postmark

package postmark

import (
	"github.com/refaktor/rye/env"
)

var Builtins_postmark = map[string]*env.Builtin{}
