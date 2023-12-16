//go:build b_mail
// +build b_mail

package evaldo

import (
	"io"

	"github.com/refaktor/rye/env"

	"github.com/thomasberger/parsemail"
)

var Builtins_mail = map[string]*env.Builtin{

	"rye-reader//parse-email": {
		Argsn: 1,
		Doc:   "Parsing email.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch reader := arg0.(type) {
			case env.Native:
				email, err := parsemail.Parse(reader.Value.(io.Reader))
				if err != nil {
					return MakeBuiltinError(ps, err.Error(), "rye-reader//parse-email")
				}
				return *env.NewNative(ps.Idx, email, "parsed-email")
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "rye-reader//parse-email")
			}
		},
	},

	"parsed-email//subject?": {
		Argsn: 1,
		Doc:   "TODODOC",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch email := arg0.(type) {
			case env.Native:
				return env.String{email.Value.(parsemail.Email).Subject}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "parsed-email//subject?")
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
		Doc:   "TODODOC",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch email := arg0.(type) {
			case env.Native:
				return env.String{email.Value.(parsemail.Email).MessageID}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "parsed-email//message-id?")
			}
		},
	},

	"parsed-email//html-body?": {
		Argsn: 1,
		Doc:   "TODODOC",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch email := arg0.(type) {
			case env.Native:
				return env.String{email.Value.(parsemail.Email).HTMLBody}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "parsed-email//html-body?")
			}
		},
	},

	"parsed-email//text-body?": {
		Argsn: 1,
		Doc:   "TODODOC",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch email := arg0.(type) {
			case env.Native:
				return env.String{email.Value.(parsemail.Email).TextBody}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "parsed-email//text-body?")
			}
		},
	},
	"parsed-email//attachments?": {
		Argsn: 1,
		Doc:   "TODODOC",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewNative(ps.Idx, arg0, "smtpd")
		},
	},
	"parsed-email//embedded-files?": {
		Argsn: 1,
		Doc:   "TODODOC",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewNative(ps.Idx, arg0, "smtpd")
		},
	},
}

// todo - NAUK PO
// * msfg.header.Get(subject)
// .... attachment , text, gXSXS
