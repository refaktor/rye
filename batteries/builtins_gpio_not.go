//go:build !add_gpio && !add_gpio_sim
// +build !add_gpio,!add_gpio_sim

package batteries

import "github.com/refaktor/rye/env"

var Builtins_gpio = map[string]*env.Builtin{}
