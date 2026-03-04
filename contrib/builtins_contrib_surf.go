//go:build b_surf && b_contrib
// +build b_surf,b_contrib

package contrib

import (
	"github.com/refaktor/rye/contrib/surf"
	"github.com/refaktor/rye/env"
)

func init() {
	// Register surf builtins when b_surf tag is present
	surfRegistrationFuncs = append(surfRegistrationFuncs, func(ps *env.ProgramState, builtinNames *map[string]int) {
		RegisterBuiltins2(surf.Builtins_surf, ps, "surf", builtinNames)
	})
}
