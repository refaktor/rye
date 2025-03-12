//go:build !no_mail
// +build !no_mail

package evaldo

import (
	"io"

	"github.com/refaktor/rye/env"
	"github.com/thomasberger/parsemail"
)

var Builtins_mail = map[string]*env.Builtin{

	//
	// ##### Mail ##### "Email parsing functions"
	//
	// Tests:
	// equal { reader %email.eml |parse-email |type? } 'native
	// equal { reader %email.eml |parse-email |kind? } 'parsed-email
	// Args:
	// * reader: native reader object containing email data
	// Returns:
	// * native parsed-email object
	"reader//parse-email": {
		Argsn: 1,
		Doc:   "Parses email data from a reader.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch reader := arg0.(type) {
			case env.Native:
				email, err := parsemail.Parse(reader.Value.(io.Reader))
				if err != nil {
					return MakeBuiltinError(ps, err.Error(), "reader//parse-email")
				}
				return *env.NewNative(ps.Idx, email, "parsed-email")
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "reader//parse-email")
			}
		},
	},

	// Tests:
	// equal { reader %email.eml |parse-email |subject? |type? } 'string
	// Args:
	// * email: native parsed-email object
	// Returns:
	// * string containing the email subject
	"parsed-email//subject?": {
		Argsn: 1,
		Doc:   "Gets the subject from a parsed email.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch email := arg0.(type) {
			case env.Native:
				return *env.NewString(email.Value.(parsemail.Email).Subject)
			default:
				return *MakeArgError(ps, 1, []env.Type{env.NativeType}, "parsed-email//subject?")
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

	// Tests:
	// equal { reader %email.eml |parse-email |message-id? |type? } 'string
	// Args:
	// * email: native parsed-email object
	// Returns:
	// * string containing the email message ID
	"parsed-email//message-id?": {
		Argsn: 1,
		Doc:   "Gets the message ID from a parsed email.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch email := arg0.(type) {
			case env.Native:
				return *env.NewString(email.Value.(parsemail.Email).MessageID)
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "parsed-email//message-id?")
			}
		},
	},

	// Tests:
	// equal { reader %email.eml |parse-email |html-body? |type? } 'string
	// Args:
	// * email: native parsed-email object
	// Returns:
	// * string containing the HTML body of the email
	"parsed-email//html-body?": {
		Argsn: 1,
		Doc:   "Gets the HTML body from a parsed email.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch email := arg0.(type) {
			case env.Native:
				return *env.NewString(email.Value.(parsemail.Email).HTMLBody)
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "parsed-email//html-body?")
			}
		},
	},

	// Tests:
	// equal { reader %email.eml |parse-email |text-body? |type? } 'string
	// Args:
	// * email: native parsed-email object
	// Returns:
	// * string containing the plain text body of the email
	"parsed-email//text-body?": {
		Argsn: 1,
		Doc:   "Gets the plain text body from a parsed email.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch email := arg0.(type) {
			case env.Native:
				return *env.NewString(email.Value.(parsemail.Email).TextBody)
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "parsed-email//text-body?")
			}
		},
	},
	// Args:
	// * email: native parsed-email object
	// Returns:
	// * native object containing email attachments
	"parsed-email//attachments?": {
		Argsn: 1,
		Doc:   "Gets the attachments from a parsed email.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewNative(ps.Idx, arg0, "smtpd")
		},
	},
	// Args:
	// * email: native parsed-email object
	// Returns:
	// * native object containing embedded files from the email
	"parsed-email//embedded-files?": {
		Argsn: 1,
		Doc:   "Gets the embedded files from a parsed email.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewNative(ps.Idx, arg0, "smtpd")
		},
	},
}

// todo - NAUK PO
// * msfg.header.Get(subject)
// .... attachment , text, gXSXS
