//go:build !unix

package term

import "fmt"

// SuspendProcess implements fallback process suspension for non-Unix systems
func SuspendProcess() error {
	fmt.Println("[ Process suspended with Ctrl+Z - Use 'fg' to resume ]")
	return fmt.Errorf("process suspended with Ctrl+Z")
}
