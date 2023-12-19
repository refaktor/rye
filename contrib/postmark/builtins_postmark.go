//go:build b_postmark
// +build b_postmark

package postmark

import (
	"bufio"
	"encoding/base64"
	"io/ioutil"
	"mime"
	"os"
	"path/filepath"
	"strings"

	// "bytes"
	"context"

	"github.com/refaktor/rye/env"
	"github.com/refaktor/rye/evaldo"

	// "fmt"

	"github.com/mrz1836/postmark"
)

var Builtins_postmark = map[string]*env.Builtin{

	"open-postmark": {
		Argsn: 1,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch token := arg0.(type) {
			case env.String:
				pm := postmark.NewClient(token.Value, "")
				return *env.NewNative(ps.Idx, pm, "postmark")
			default:
				return evaldo.MakeError(ps, "Arg 1 not String")
			}
		},
	},

	"new-postmark-email": {
		Argsn: 0,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			pe := postmark.Email{}
			return *env.NewNative(ps.Idx, &pe, "postmark-email")
		},
	},

	"postmark-email//from<-": {
		Argsn: 2,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch pe := arg0.(type) {
			case env.Native:
				switch email := arg1.(type) {
				case env.Email:
					pe.Value.(*postmark.Email).From = email.Address
					return arg0
				default:
					return evaldo.MakeError(ps, "Arg 2 not email.")
				}
			default:
				return evaldo.MakeError(ps, "Arg 1 not native.")
			}
		},
	},

	"postmark-email//to<-": {
		Argsn: 2,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch pe := arg0.(type) {
			case env.Native:
				switch email := arg1.(type) {
				case env.Email:
					pe.Value.(*postmark.Email).To = email.Address
					return arg0
				default:
					return evaldo.MakeError(ps, "Arg 2 not email.")
				}
			default:
				return evaldo.MakeError(ps, "Arg 1 not native.")
			}
		},
	},

	"postmark-email//subject<-": {
		Argsn: 2,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch pe := arg0.(type) {
			case env.Native:
				switch txt := arg1.(type) {
				case env.String:
					pe.Value.(*postmark.Email).Subject = txt.Value
					return arg0
				default:
					return evaldo.MakeError(ps, "Arg 2 not string.")
				}
			default:
				return evaldo.MakeError(ps, "Arg 1 not native.")
			}
		},
	},

	"postmark-email//text-body<-": {
		Argsn: 2,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch pe := arg0.(type) {
			case env.Native:
				switch txt := arg1.(type) {
				case env.String:
					pe.Value.(*postmark.Email).TextBody = txt.Value
					return arg0
				default:
					return evaldo.MakeError(ps, "Arg 2 not string.")
				}
			default:
				return evaldo.MakeError(ps, "Arg 1 not native.")
			}
		},
	},

	"postmark-email//attach!": {
		Argsn: 2,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch pe := arg0.(type) {
			case env.Native:
				switch attachment := arg1.(type) {
				case env.Uri:
					path := strings.Split(attachment.Path, "://")
					path1 := path[1]
					// Open file on disk.
					file, err := os.Open(path1)
					if err != nil {
						return evaldo.MakeError(ps, err.Error())
					}

					// Read entire JPG into byte slice.
					reader := bufio.NewReader(file)
					content, _ := ioutil.ReadAll(reader)

					// Encode as base64.
					encoded := base64.StdEncoding.EncodeToString(content)

					extension := filepath.Ext(path1)
					mimeType := mime.TypeByExtension(extension)
					att := postmark.Attachment{Name: path1, Content: encoded, ContentType: mimeType}
					pem := pe.Value.(*postmark.Email)
					pem.Attachments = append(pem.Attachments, att)
					return arg0
				default:
					return evaldo.MakeError(ps, "Arg 2 not string.")
				}
			default:
				return evaldo.MakeError(ps, "Arg 1 not native.")
			}
		},
	},

	"postmark-email//send": {
		Argsn: 2,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch pe := arg0.(type) {
			case env.Native:
				switch pm := arg1.(type) {
				case env.Native:
					email := pe.Value.(*postmark.Email)
					_, err := pm.Value.(*postmark.Client).SendEmail(context.Background(), *email)
					if err != nil {
						return evaldo.MakeError(ps, err.Error())
					}
					return arg0
				default:
					return evaldo.MakeError(ps, "Arg 2 not native.")
				}
			default:
				return evaldo.MakeError(ps, "Arg 1 not native.")
			}
		},
	},
}
