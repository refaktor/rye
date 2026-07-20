package batteries

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"

	"github.com/refaktor/rye/env"
	"github.com/refaktor/rye/evaldo"
)

// Global counter for process-unique IDs - thread-safe using atomic operations
var processUniqueCounter int64

// processStartTime stores when this process started for unique ID generation
var processStartTime int64

func init() {
	// Initialize the start time when the package loads
	processStartTime = time.Now().UnixNano()
}

// generateRandomString creates a cryptographically secure random string of specified length
func generateRandomString(length int) (string, error) {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes)[:length], nil
}

// generateProcessUniqueID creates a unique ID for this process run
func generateProcessUniqueID() (string, error) {
	// Get process ID
	pid := os.Getpid()

	// Get atomic counter value and increment
	counter := atomic.AddInt64(&processUniqueCounter, 1)

	// Get current nanosecond timestamp
	nowNano := time.Now().UnixNano()

	// Generate a short random component for extra uniqueness
	randomStr, err := generateRandomString(6)
	if err != nil {
		return "", err
	}

	// Combine all components: rye_[PID]_[COUNTER]_[NANOS]_[RANDOM]
	uniqueID := fmt.Sprintf("rye_%d_%d_%d_%s", pid, counter, nowNano, randomStr)
	return uniqueID, nil
}

var builtins_unique = map[string]*env.Builtin{

	//
	// ##### Unique ID Generation ##### "Process-unique identifier generation functions"
	//

	// Tests:
	// different { Rye-itself//unique-id } { Rye-itself//unique-id }
	// does { Rye-itself//unique-id } |type? |= 'string
	// Args:
	// * rye-ctx: Rye runtime context (automatically provided)
	// Returns:
	// * string: Process-unique identifier
	"Rye-itself//Unique-id?": {
		Argsn: 1,
		Doc:   "Generates a process-unique identifier string. Each call within the same process run is guaranteed to return a different ID.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			// arg0 should be Rye context, but we don't need to validate it for this implementation
			uniqueID, err := generateProcessUniqueID()
			if err != nil {
				return evaldo.MakeBuiltinError(ps, fmt.Sprintf("Failed to generate unique ID: %v", err), "Rye-itself//unique-id")
			}
			return *env.NewString(uniqueID)
		},
	},

	// Tests:
	// different { Rye-itself//unique-temp-dir } { Rye-itself//unique-temp-dir }
	// does { Rye-itself//unique-temp-dir } |type? |= 'uri
	// Args:
	// * rye-ctx: Rye runtime context (automatically provided)
	// Returns:
	// * uri: File URI pointing to a unique temporary directory
	"Rye-itself//Temp-dir?": {
		Argsn: 1,
		Doc:   "Creates a unique temporary directory and returns its file URI. The directory is guaranteed to be unique for this process run.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			// Generate unique ID for directory name
			uniqueID, err := generateProcessUniqueID()
			if err != nil {
				return evaldo.MakeBuiltinError(ps, fmt.Sprintf("Failed to generate unique ID: %v", err), "Rye-itself//unique-temp-dir")
			}

			// Create the full path in system temp directory
			tempDir := os.TempDir()
			uniqueDirPath := filepath.Join(tempDir, uniqueID)

			// Create the directory
			err = os.MkdirAll(uniqueDirPath, 0755)
			if err != nil {
				return evaldo.MakeBuiltinError(ps, fmt.Sprintf("Failed to create temp directory: %v", err), "Rye-itself//unique-temp-dir")
			}

			// Return as file URI
			return *env.NewFileUri(ps.Idx, uniqueDirPath)
		},
	},

	// Tests:
	// different { Rye-itself//unique-temp-file } { Rye-itself//unique-temp-file }
	// does { Rye-itself//unique-temp-file } |type? |= 'uri
	// Args:
	// * rye-ctx: Rye runtime context (automatically provided)
	// Returns:
	// * uri: File URI pointing to a unique temporary file path
	"Rye-itself//Temp-file?": {
		Argsn: 1,
		Doc:   "Generates a unique temporary file path and returns its file URI. The file path is guaranteed to be unique for this process run. Note: the file is not created, only the path is generated.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			// Generate unique ID for file name
			uniqueID, err := generateProcessUniqueID()
			if err != nil {
				return evaldo.MakeBuiltinError(ps, fmt.Sprintf("Failed to generate unique ID: %v", err), "Rye-itself//unique-temp-file")
			}

			// Create the full path in system temp directory with .tmp extension
			tempDir := os.TempDir()
			uniqueFilePath := filepath.Join(tempDir, uniqueID+".tmp")

			// Return as file URI (don't create the file, just return the path)
			return *env.NewFileUri(ps.Idx, uniqueFilePath)
		},
	},
}

