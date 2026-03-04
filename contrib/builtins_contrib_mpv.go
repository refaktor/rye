//go:build b_mpv && b_contrib
// +build b_mpv,b_contrib

package contrib

import (
	"github.com/refaktor/rye/contrib/mpv"
	"github.com/refaktor/rye/env"
)

func init() {
	// Register mpv builtins when b_mpv tag is present
	surfRegistrationFuncs = append(surfRegistrationFuncs, func(ps *env.ProgramState, builtinNames *map[string]int) {
		RegisterBuiltins2(mpv.Builtins_mpv, ps, "mpv", builtinNames)
	})
}
