//go:build b_smtpd
// +build b_smtpd

package evaldo

import (
	"bytes"
	"net"
	"rye/env"

	"github.com/jinzhu/copier"
	"github.com/mhale/smtpd"
)

var Builtins_smtpd = map[string]*env.Builtin{

	"new-smtpd": {
		Argsn: 1,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewNative(ps.Idx, arg0, "smtpd")
		},
	},

	"smtpd//serve": {
		Argsn: 4,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch server := arg0.(type) {
			case env.Native:
				switch handler := arg1.(type) {
				case env.Function:
					switch name := arg2.(type) {
					case env.String:
						smtpd.ListenAndServe(server.Value.(env.String).Value,
							func(origin net.Addr, from string, to []string, data []byte) error {
								ps.FailureFlag = false
								ps.ErrorFlag = false
								ps.ReturnFlag = false
								psTemp := env.ProgramState{}
								copier.Copy(&psTemp, &ps)
								// msg, _ := mail.ReadMessage(bytes.NewReader(data))
								lstTo := make([]interface{}, len(to))
								for i, v := range to {
									lstTo[i] = v
								}
								CallFunctionArgs4(handler, ps, *env.NewNative(ps.Idx, bytes.NewReader(data), "rye-reader"), env.String{from}, *env.NewList(lstTo), *env.NewNative(ps.Idx, origin, "new-addr"), nil)
								//msg, _ := mail.ReadMessage(bytes.NewReader(data))
								//subject := msg.Header.Get("Subject")
								//log.Printf("Received mail from %s for %s with subject %s", from, to[0], subject)
								return nil
							}, name.Value, "")
						return arg0
					default:
						return makeError(ps, "arg 3 should be string")
					}
				default:
					return makeError(ps, "arg 2 should be string")
				}
			default:
				return makeError(ps, "arg 1 should be native")
			}

		},
	},
}

// todo - NAUK PO
// * msfg.header.Get(subject)
// .... attachment , text, gXSXS
