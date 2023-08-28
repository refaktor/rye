//go:build b_mail
// +build b_mail

package evaldo

import (
	"github.com/thomasberger/parsemail"
	"io"
	"rye/env"
)

var Builtins_mail = map[string]*env.Builtin{

	"rye-reader//parse-email": {
		Argsn: 1,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch reader := arg0.(type) {
			case env.Native:
				email, err := parsemail.Parse(reader.Value.(io.Reader))
				if err != nil {
					return makeError(ps, err.Error())
				}
				return *env.NewNative(ps.Idx, email, "parsed-email")
			default:
				return makeError(ps, "Arg 1 not Native")
			}
		},
	},

	"parsed-email//subject?": {
		Argsn: 1,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch email := arg0.(type) {
			case env.Native:
				return env.String{email.Value.(parsemail.Email).Subject}
			default:
				return makeError(ps, "Arg 1 not Native")
			}
		},
	},

	/*

		"parsed-email//from?": {
			Argsn: 1,
			Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
				switch email := arg0.(type) {
				case env.Native:
					return env.String{email.Value.(parsemail.Email).From}
				default:
					return makeError(ps, "Arg 1 not Native")
				}
			},
		},

		"parsed-email//reply-to?": {
			Argsn: 1,
			Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
				return *env.NewNative(ps.Idx, arg0, "smtpd")
			},
		},

		"parsed-email//date?": {
			Argsn: 1,
			Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
				return *env.NewNative(ps.Idx, arg0, "smtpd")
			},
		},

		"parsed-email//to?": {
			Argsn: 1,
			Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
				return *env.NewNative(ps.Idx, arg0, "smtpd")
			},
		},

	*/

	"parsed-email//message-id?": {
		Argsn: 1,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch email := arg0.(type) {
			case env.Native:
				return env.String{email.Value.(parsemail.Email).MessageID}
			default:
				return makeError(ps, "Arg 1 not Native")
			}
		},
	},

	"parsed-email//html-body?": {
		Argsn: 1,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch email := arg0.(type) {
			case env.Native:
				return env.String{email.Value.(parsemail.Email).HTMLBody}
			default:
				return makeError(ps, "Arg 1 not Native")
			}
		},
	},

	"parsed-email//text-body?": {
		Argsn: 1,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch email := arg0.(type) {
			case env.Native:
				return env.String{email.Value.(parsemail.Email).TextBody}
			default:
				return makeError(ps, "Arg 1 not Native")
			}
		},
	},
	"parsed-email//attachments?": {
		Argsn: 1,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewNative(ps.Idx, arg0, "smtpd")
		},
	},
	"parsed-email//embedded-files?": {
		Argsn: 1,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewNative(ps.Idx, arg0, "smtpd")
		},
	},
}

// todo - NAUK PO
// * msfg.header.Get(subject)
// .... attachment , text, gXSXS
