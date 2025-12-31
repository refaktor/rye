//go:build !no_smtpd
// +build !no_smtpd

package evaldo

import (
	"bytes"
	"fmt"
	"net"

	"github.com/refaktor/rye/env"

	"github.com/jinzhu/copier"
	"github.com/mhale/smtpd"
)

var Builtins_smtpd = map[string]*env.Builtin{

	//
	// ##### SMTP Server Functions ##### "Creating and running SMTP mail servers."
	//

	// Tests:
	// equal { smtp-server 2525 |type? } 'native
	// equal { smtp-server ":2525" |kind? } 'smtpd
	// Args:
	// * address: String containing the server address (e.g., ":2525", "localhost:587")
	// Returns:
	// * native smtpd object configured to listen on the specified address
	"smtp-server": {
		Argsn: 1,
		Doc:   "Creates a new SMTP server that can receive incoming email messages on the specified address.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			// Wrap the address string in a native smtpd object
			return *env.NewNative(ps.Idx, arg0, "smtpd")
		},
	},

	// Example:
	// handler: fn { reader from to origin } { print "Got email from:" from }
	// smtp-server ":2525"
	// |Serve handler "TestSMTP" ""
	// Args:
	// * server: Native smtpd object created by smtp-server
	// * handler: Function that processes incoming emails (reader from to origin -> ...)
	// * appname: String name for the SMTP server application
	// * password: String password for SMTP authentication (empty string for no auth)
	// Returns:
	// * the server object after starting to listen, or error if unable to serve
	"smtpd//Serve": {
		Argsn: 4,
		Doc:   "Starts the SMTP server listening for incoming emails and calls the handler function for each received message.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch server := arg0.(type) {
			case env.Native:
				switch handler := arg1.(type) {
				case env.Function:
					switch name := arg2.(type) {
					case env.String:
						// Start SMTP server with custom mail handler function
						err := smtpd.ListenAndServe(server.Value.(env.String).Value,
							func(origin net.Addr, from string, to []string, data []byte) error {
								// Reset program state flags for clean handler execution
								ps.FailureFlag = false
								ps.ErrorFlag = false
								ps.ReturnFlag = false
								// Create temporary program state to avoid conflicts
								psTemp := env.ProgramState{}
								err := copier.Copy(&psTemp, &ps)
								if err != nil {
									fmt.Println(err.Error()) // TODO: proper error handling
								}
								// Convert recipient list to Rye list format
								lstTo := make([]any, len(to))
								for i, v := range to {
									lstTo[i] = v
								}
								// Call Rye handler function with email data:
								// - reader: bytes reader for raw email data
								// - from: sender email address string
								// - to: list of recipient email addresses
								// - origin: network address of sender
								CallFunctionArgs4(handler, ps,
									*env.NewNative(ps.Idx, bytes.NewReader(data), "reader"),
									env.NewString(from),
									*env.NewList(lstTo),
									*env.NewNative(ps.Idx, origin, "new-addr"), nil)
								return nil
							}, name.Value, "")
						if err != nil {
							return makeError(ps, err.Error())
						}
						return arg0
					default:
						return MakeArgError(ps, 3, []env.Type{env.StringType}, "smtpd//serve")
					}
				default:
					return MakeArgError(ps, 2, []env.Type{env.FunctionType}, "smtpd//serve")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "smtpd//serve")
			}
		},
	},
}

// todo - NAUK PO
// * msfg.header.Get(subject)
// .... attachment , text, gXSXS
