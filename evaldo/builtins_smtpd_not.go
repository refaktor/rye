//go:build !b_smtpd
// +build !b_smtpd

package evaldo

import (
	"rye/env"
)

var Builtins_smtpd = map[string]*env.Builtin{}

// todo - NAUK POA+''
// * mail.ReadMEssage/(bytes..)
// * msfg.header.Get(subject)
// .... attachment , text, gXSXS
