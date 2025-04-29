//go:build !no_mcp
// +build !no_mcp

package evaldo

import (
	"sync"

	"github.com/refaktor/rye/env"
)

// Global variables to store MCP servers and clients
var (
	mcpServerMutex sync.RWMutex
)

var Builtins_mcp = map[string]*env.Builtin{
	// Server functions
	"mcp-server//create": {
		Argsn: 2,
		Doc:   "Creates a new MCP server with the given name and version.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch name := arg0.(type) {
			case env.String:
				switch version := arg1.(type) {
				case env.String:
					mcpServerMutex.Lock()
					defer mcpServerMutex.Unlock()

					// Create a simple object to represent the server
					serverInfo := struct {
						Name    string
						Version string
					}{
						Name:    name.Value,
						Version: version.Value,
					}

					// Return the server as a native object
					return *env.NewNative(ps.Idx, serverInfo, "Rye-mcp-server")
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "mcp-server//create")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "mcp-server//create")
			}
		},
	},

	"mcp//create-resource": {
		Argsn: 3,
		Doc:   "Creates a new MCP resource with the given URI, name, and MIME type.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch uri := arg0.(type) {
			case env.String:
				switch name := arg1.(type) {
				case env.String:
					switch mimeType := arg2.(type) {
					case env.String:
						// Create a simple object to represent the resource
						resourceInfo := struct {
							URI      string
							Name     string
							MIMEType string
						}{
							URI:      uri.Value,
							Name:     name.Value,
							MIMEType: mimeType.Value,
						}

						// Return the resource as a native object
						return *env.NewNative(ps.Idx, resourceInfo, "Rye-mcp-resource")
					default:
						ps.FailureFlag = true
						return MakeArgError(ps, 3, []env.Type{env.StringType}, "mcp-resource//create")
					}
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "mcp-resource//create")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "mcp-resource//create")
			}
		},
	},

	"mcp//create-tool": {
		Argsn: 2,
		Doc:   "Creates a new MCP tool with the given name and description.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch name := arg0.(type) {
			case env.String:
				switch desc := arg1.(type) {
				case env.String:
					// Create a simple object to represent the tool
					toolInfo := struct {
						Name        string
						Description string
					}{
						Name:        name.Value,
						Description: desc.Value,
					}

					// Return the tool as a native object
					return *env.NewNative(ps.Idx, toolInfo, "Rye-mcp-tool")
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "mcp-tool//create")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "mcp-tool//create")
			}
		},
	},

	"mcp//create-prompt": {
		Argsn: 2,
		Doc:   "Creates a new MCP prompt with the given name and description.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch name := arg0.(type) {
			case env.String:
				switch desc := arg1.(type) {
				case env.String:
					// Create a simple object to represent the prompt
					promptInfo := struct {
						Name        string
						Description string
					}{
						Name:        name.Value,
						Description: desc.Value,
					}

					// Return the prompt as a native object
					return *env.NewNative(ps.Idx, promptInfo, "Rye-mcp-prompt")
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "mcp-prompt//create")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "mcp-prompt//create")
			}
		},
	},

	"mcp//protocol-version": {
		Argsn: 0,
		Doc:   "Returns the latest MCP protocol version.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return env.String{Value: "2024-11-05"} // Hardcoded latest protocol version
		},
	},

	// Server methods
	"Rye-mcp-server//get-name": {
		Argsn: 1,
		Doc:   "Gets the name of the MCP server.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch srv := arg0.(type) {
			case env.Native:
				if srv.Type() != env.NativeType {
					ps.FailureFlag = true
					return MakeBuiltinError(ps, "Expected native type", "Rye-mcp-server//get-name")
				}

				serverInfo, ok := srv.Value.(struct {
					Name    string
					Version string
				})
				if !ok {
					ps.FailureFlag = true
					return MakeBuiltinError(ps, "Failed to cast to server info", "Rye-mcp-server//get-name")
				}

				return env.String{Value: serverInfo.Name}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Rye-mcp-server//get-name")
			}
		},
	},

	"Rye-mcp-server//get-version": {
		Argsn: 1,
		Doc:   "Gets the version of the MCP server.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch srv := arg0.(type) {
			case env.Native:
				if srv.Type() != env.NativeType {
					ps.FailureFlag = true
					return MakeBuiltinError(ps, "Expected native type", "Rye-mcp-server//get-version")
				}

				serverInfo, ok := srv.Value.(struct {
					Name    string
					Version string
				})
				if !ok {
					ps.FailureFlag = true
					return MakeBuiltinError(ps, "Failed to cast to server info", "Rye-mcp-server//get-version")
				}

				return env.String{Value: serverInfo.Version}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Rye-mcp-server//get-version")
			}
		},
	},

	// Resource methods
	"Rye-mcp-resource//get-uri": {
		Argsn: 1,
		Doc:   "Gets the URI of the MCP resource.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch res := arg0.(type) {
			case env.Native:
				if res.Type() != env.NativeType {
					ps.FailureFlag = true
					return MakeBuiltinError(ps, "Expected native type", "Rye-mcp-resource//get-uri")
				}

				resourceInfo, ok := res.Value.(struct {
					URI      string
					Name     string
					MIMEType string
				})
				if !ok {
					ps.FailureFlag = true
					return MakeBuiltinError(ps, "Failed to cast to resource info", "Rye-mcp-resource//get-uri")
				}

				return env.String{Value: resourceInfo.URI}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Rye-mcp-resource//get-uri")
			}
		},
	},

	"Rye-mcp-resource//get-name": {
		Argsn: 1,
		Doc:   "Gets the name of the MCP resource.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch res := arg0.(type) {
			case env.Native:
				if res.Type() != env.NativeType {
					ps.FailureFlag = true
					return MakeBuiltinError(ps, "Expected native type", "Rye-mcp-resource//get-name")
				}

				resourceInfo, ok := res.Value.(struct {
					URI      string
					Name     string
					MIMEType string
				})
				if !ok {
					ps.FailureFlag = true
					return MakeBuiltinError(ps, "Failed to cast to resource info", "Rye-mcp-resource//get-name")
				}

				return env.String{Value: resourceInfo.Name}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Rye-mcp-resource//get-name")
			}
		},
	},

	"Rye-mcp-resource//get-mime-type": {
		Argsn: 1,
		Doc:   "Gets the MIME type of the MCP resource.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch res := arg0.(type) {
			case env.Native:
				if res.Type() != env.NativeType {
					ps.FailureFlag = true
					return MakeBuiltinError(ps, "Expected native type", "Rye-mcp-resource//get-mime-type")
				}

				resourceInfo, ok := res.Value.(struct {
					URI      string
					Name     string
					MIMEType string
				})
				if !ok {
					ps.FailureFlag = true
					return MakeBuiltinError(ps, "Failed to cast to resource info", "Rye-mcp-resource//get-mime-type")
				}

				return env.String{Value: resourceInfo.MIMEType}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Rye-mcp-resource//get-mime-type")
			}
		},
	},

	// Tool methods
	"Rye-mcp-tool//get-name": {
		Argsn: 1,
		Doc:   "Gets the name of the MCP tool.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch tool := arg0.(type) {
			case env.Native:
				if tool.Type() != env.NativeType {
					ps.FailureFlag = true
					return MakeBuiltinError(ps, "Expected native type", "Rye-mcp-tool//get-name")
				}

				toolInfo, ok := tool.Value.(struct {
					Name        string
					Description string
				})
				if !ok {
					ps.FailureFlag = true
					return MakeBuiltinError(ps, "Failed to cast to tool info", "Rye-mcp-tool//get-name")
				}

				return env.String{Value: toolInfo.Name}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Rye-mcp-tool//get-name")
			}
		},
	},

	"Rye-mcp-tool//get-description": {
		Argsn: 1,
		Doc:   "Gets the description of the MCP tool.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch tool := arg0.(type) {
			case env.Native:
				if tool.Type() != env.NativeType {
					ps.FailureFlag = true
					return MakeBuiltinError(ps, "Expected native type", "Rye-mcp-tool//get-description")
				}

				toolInfo, ok := tool.Value.(struct {
					Name        string
					Description string
				})
				if !ok {
					ps.FailureFlag = true
					return MakeBuiltinError(ps, "Failed to cast to tool info", "Rye-mcp-tool//get-description")
				}

				return env.String{Value: toolInfo.Description}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Rye-mcp-tool//get-description")
			}
		},
	},

	// Prompt methods
	"Rye-mcp-prompt//get-name": {
		Argsn: 1,
		Doc:   "Gets the name of the MCP prompt.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch prompt := arg0.(type) {
			case env.Native:
				if prompt.Type() != env.NativeType {
					ps.FailureFlag = true
					return MakeBuiltinError(ps, "Expected native type", "Rye-mcp-prompt//get-name")
				}

				promptInfo, ok := prompt.Value.(struct {
					Name        string
					Description string
				})
				if !ok {
					ps.FailureFlag = true
					return MakeBuiltinError(ps, "Failed to cast to prompt info", "Rye-mcp-prompt//get-name")
				}

				return env.String{Value: promptInfo.Name}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Rye-mcp-prompt//get-name")
			}
		},
	},

	"Rye-mcp-prompt//get-description": {
		Argsn: 1,
		Doc:   "Gets the description of the MCP prompt.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch prompt := arg0.(type) {
			case env.Native:
				if prompt.Type() != env.NativeType {
					ps.FailureFlag = true
					return MakeBuiltinError(ps, "Expected native type", "Rye-mcp-prompt//get-description")
				}

				promptInfo, ok := prompt.Value.(struct {
					Name        string
					Description string
				})
				if !ok {
					ps.FailureFlag = true
					return MakeBuiltinError(ps, "Failed to cast to prompt info", "Rye-mcp-prompt//get-description")
				}

				return env.String{Value: promptInfo.Description}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Rye-mcp-prompt//get-description")
			}
		},
	},
}
