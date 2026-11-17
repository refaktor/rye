//go:build !no_mail
// +build !no_mail

package batteries

import (
	"io"

	"github.com/refaktor/rye/env"
	"github.com/refaktor/rye/evaldo"
	"github.com/thomasberger/parsemail"
)

var Builtins_mail = map[string]*env.Builtin{

	//
	// ##### Mail ##### "Email parsing functions"
	//
	// Example:
	// equal { reader %email.eml |parse-email |type? } 'native
	// equal { reader %email.eml |parse-email |kind? } 'parsed-email
	// Args:
	// * reader: native reader object containing email data
	// Returns:
	// * native parsed-email object
	"reader//Parse-email": {
		Effect: env.EffInput,
		Argsn:  1,
		Doc:    "Parses email data from a reader.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch reader := arg0.(type) {
			case env.Native:
				email, err := parsemail.Parse(reader.Value.(io.Reader))
				if err != nil {
					return evaldo.MakeBuiltinError(ps, err.Error(), "reader//parse-email")
				}
				return *env.NewNative(ps.Idx, email, "parsed-email")
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "reader//parse-email")
			}
		},
	},

	// Example:
	// equal { reader %email.eml |parse-email |subject? |type? } 'string
	// Args:
	// * email: native parsed-email object
	// Returns:
	// * string containing the email subject
	"parsed-email//Subject?": {
		Effect: env.EffNone,
		Argsn:  1,
		Doc:    "Gets the subject from a parsed email.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch email := arg0.(type) {
			case env.Native:
				return *env.NewString(email.Value.(parsemail.Email).Subject)
			default:
				return *evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "parsed-email//Subject?")
			}
		},
	},

	/*

		"parsed-email//From?": {
			Argsn: 1,
			Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
				switch email := arg0.(type) {
				case env.Native:
					return env.String{email.Value.(parsemail.Email).From}
				default:
					return evaldo.MakeError(ps, "Arg 1 not Native")
				}
			},
		},

		"parsed-email//Reply-to?": {
			Argsn: 1,
			Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
				return *env.NewNative(ps.Idx, arg0, "smtpd")
			},
		},

		"parsed-email//Date?": {
			Argsn: 1,
			Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
				return *env.NewNative(ps.Idx, arg0, "smtpd")
			},
		},

		"parsed-email//To?": {
			Argsn: 1,
			Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
				return *env.NewNative(ps.Idx, arg0, "smtpd")
			},
		},

	*/

	// Example:
	// equal { reader %email.eml |parse-email |message-id? |type? } 'string
	// Args:
	// * email: native parsed-email object
	// Returns:
	// * string containing the email message ID
	"parsed-email//Message-id?": {
		Effect: env.EffNone,
		Argsn:  1,
		Doc:    "Gets the message ID from a parsed email.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch email := arg0.(type) {
			case env.Native:
				return *env.NewString(email.Value.(parsemail.Email).MessageID)
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "parsed-email//message-id?")
			}
		},
	},

	// Example:
	// equal { reader %email.eml |parse-email |html-body? |type? } 'string
	// Args:
	// * email: native parsed-email object
	// Returns:
	// * string containing the HTML body of the email
	"parsed-email//Html-body?": {
		Effect: env.EffNone,
		Argsn:  1,
		Doc:    "Gets the HTML body from a parsed email.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch email := arg0.(type) {
			case env.Native:
				return *env.NewString(email.Value.(parsemail.Email).HTMLBody)
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "parsed-email//html-body?")
			}
		},
	},

	// Example:
	// equal { Reader %email.eml |parse-email |text-body? |type? } 'string
	// Args:
	// * email: native parsed-email object
	// Returns:
	// * string containing the plain text body of the email
	"parsed-email//Text-body?": {
		Effect: env.EffNone,
		Argsn:  1,
		Doc:    "Gets the plain text body from a parsed email.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch email := arg0.(type) {
			case env.Native:
				return *env.NewString(email.Value.(parsemail.Email).TextBody)
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "parsed-email//text-body?")
			}
		},
	},
	// Args:
	// * email: native parsed-email object
	// Returns:
	// * native object containing email attachments
	"parsed-email//Attachments?": {
		Effect: env.EffNone,
		Argsn:  1,
		Doc:    "Gets the attachments from a parsed email.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch email := arg0.(type) {
			case env.Native:
				attachments := email.Value.(parsemail.Email).Attachments
				items := make([]env.Object, len(attachments))
				for i, attachment := range attachments {
					items[i] = *env.NewDict(map[string]any{
						"filename":     attachment.Filename,
						"content-type": attachment.ContentType,
						"data":         *env.NewNative(ps.Idx, attachment.Data, "reader"),
					})
				}
				return *env.NewBlock(*env.NewTSeries(items))
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "parsed-email//Attachments?")
			}
		},
	},
	// Args:
	// * email: native parsed-email object
	// Returns:
	// * native object containing embedded files from the email
	"parsed-email//Embedded-files?": {
		Effect: env.EffNone,
		Argsn:  1,
		Doc:    "Gets the embedded files from a parsed email.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch email := arg0.(type) {
			case env.Native:
				files := email.Value.(parsemail.Email).EmbeddedFiles
				items := make([]env.Object, len(files))
				for i, file := range files {
					items[i] = *env.NewDict(map[string]any{
						"cid":          file.CID,
						"filename":     file.Filename,
						"content-type": file.ContentType,
						"data":         *env.NewNative(ps.Idx, file.Data, "reader"),
					})
				}
				return *env.NewBlock(*env.NewTSeries(items))
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "parsed-email//Embedded-files?")
			}
		},
	},
}

// todo - NAUK PO
// * msfg.header.Get(subject)
// .... attachment , text, gXSXS
