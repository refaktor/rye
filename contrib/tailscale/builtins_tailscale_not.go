//go:build !b_tailscale
// +build !b_tailscale

package tailscale

import (
	"github.com/refaktor/rye/env"
)

// Builtins_tailscale is empty when Rye is built without the b_tailscale tag.
var Builtins_tailscale = map[string]*env.Builtin{}
