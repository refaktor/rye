//go:build linux || darwin || windows
// +build linux darwin windows

package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/jwalton/go-supportscolor"
	"github.com/refaktor/rye/env"
	"github.com/refaktor/rye/runner"
	"github.com/refaktor/rye/security"
)

// Global variables to store the current security profiles
// These can be accessed from builtins to enforce additional restrictions
// These are now defined in the security package

type TagType int
type RjType int
type Series []any

type anyword struct {
	kind RjType
	idx  int
}

type node struct {
	kind  RjType
	value any
}

var CODE []any

//
// main function. Dispatches to appropriate mode function
//

func main() {
	// Initialize security profiles
	// These are no-ops on non-Linux systems or when built without the appropriate tags
	// The actual configuration will be set in runner.DoMain based on command-line flags

	supportscolor.Stdout()
	runner.DoMain(func(ps *env.ProgramState) error {
		// Get script directory from runner
		scriptDir := runner.GetScriptDirectory()

		// Build CLI-based policy (lowest priority)
		cliPolicy := buildCLIPolicy()

		// Load security policy with proper precedence:
		// 1. Embedded policy (compiled into binary) - highest priority
		// 2. System policy (/etc/rye/mandatory.yaml)
		// 3. Local policy (.ryesec in script directory)
		// 4. CLI flags - lowest priority
		policy, err := security.LoadSecurityPolicy(scriptDir, cliPolicy)
		if err != nil {
			// Policy file exists but is insecure or malformed - this is FATAL
			fmt.Fprintf(os.Stderr, "FATAL: Security policy error: %v\n", err)
			os.Exit(1)
		}

		// If we have an embedded or system policy, validate the script path
		if policy.Source == security.PolicySourceEmbedded || policy.Source == security.PolicySourceSystem {
			if ps.ScriptPath != "" {
				if err := policy.ValidateScriptPath(ps.ScriptPath); err != nil {
					fmt.Fprintf(os.Stderr, "FATAL: %v\n", err)
					os.Exit(1)
				}
			}
		}

		// Apply the security policy
		if err := security.ApplySecurityPolicy(policy); err != nil {
			// For embedded/mandatory policies, failure is fatal
			if policy.Mandatory || policy.Source == security.PolicySourceEmbedded {
				fmt.Fprintf(os.Stderr, "FATAL: Failed to apply mandatory security policy: %v\n", err)
				os.Exit(1)
			}
			// For other policies, warn but continue
			fmt.Fprintf(os.Stderr, "Warning: Failed to apply security policy: %v\n", err)
		}

		return nil
	})
}

// buildCLIPolicy creates a security policy from command-line flags
func buildCLIPolicy() *security.SecurityPolicy {
	policy := &security.SecurityPolicy{
		Mandatory: false, // CLI policies can always be overridden by file-based policies
	}

	// Seccomp from CLI flags
	if *runner.SeccompProfile != "" {
		policy.Seccomp.Enabled = true
		policy.Seccomp.Profile = *runner.SeccompProfile
		policy.Seccomp.Action = *runner.SeccompAction
		if policy.Seccomp.Action == "" {
			policy.Seccomp.Action = "errno"
		}
	}

	// Landlock from CLI flags
	if *runner.LandlockEnabled {
		policy.Landlock.Enabled = true
		policy.Landlock.Profile = *runner.LandlockProfile
		if *runner.LandlockPaths != "" {
			paths := strings.Split(*runner.LandlockPaths, ",")
			// Clean up empty paths
			for _, p := range paths {
				if p != "" {
					policy.Landlock.Paths = append(policy.Landlock.Paths, p)
				}
			}
		}
	}

	// Code signing from CLI flags
	if *runner.CodeSigEnforced {
		policy.CodeSig.Enforced = true
	}

	return policy
}

// ErrNoPath is returned when 'cd' was called without a second argument.
var ErrNoPath = errors.New("path required")
