package loader

import (
	"strings"
	"testing"

	"github.com/refaktor/rye/env"
)

func TestLoaderNoPEGSpacing(t *testing.T) {
	// Helper: asserts result is an error whose message contains keyword
	assertSpacingError := func(t *testing.T, input, keyword string) {
		t.Helper()
		result, _ := LoadStringNoPEG(input, false)
		t.Logf("Testing input: %s", input)
		t.Logf("Result type: %T", result)
		if err, isErr := result.(env.Error); isErr {
			t.Logf("Got error: %s", err.Message)
			if !strings.Contains(err.Message, keyword) {
				t.Errorf("Expected error message to contain %q, got:\n%s", keyword, err.Message)
			}
		} else {
			t.Errorf("Expected an error, got %T", result)
		}
	}

	// Note: inputs below have NO outer braces; LoadStringNoPEG wraps them.
	t.Run("No space between number and closing brace", func(t *testing.T) {
		assertSpacingError(t, "123 123 1}", "Missing space")
	})

	t.Run("No space between numbers and operator", func(t *testing.T) {
		assertSpacingError(t, "123+123", "Missing space")
	})

	t.Run("No space before comma", func(t *testing.T) {
		assertSpacingError(t, "1234, 1231", "Missing space")
	})

	t.Run("No space before comma in group", func(t *testing.T) {
		assertSpacingError(t, "(123, 123)", "Missing space")
	})

	t.Run("No space between number and letters", func(t *testing.T) {
		assertSpacingError(t, "123abc", "Missing space")
	})
}
