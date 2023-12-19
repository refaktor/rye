//go:build !b_aws
// +build !b_aws

package aws

import (
	"github.com/refaktor/rye/env"
)

var Builtins_aws = map[string]*env.Builtin{}
