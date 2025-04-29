//go:build no_mcp
// +build no_mcp

package evaldo

import (
	"github.com/refaktor/rye/env"
)

// Builtins_mcp is a placeholder for when MCP functionality is not available
var Builtins_mcp = map[string]*env.Builtin{}
