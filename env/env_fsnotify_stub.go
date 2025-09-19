//go:build wasm || b_wasm

package env

import (
	"sync"
)

// LiveEnv -- a experiment in live realoading (WASM stub version)

type LiveEnv struct {
	Active  bool
	Watcher interface{} // Placeholder since fsnotify.Watcher not available in WASM
	PsMutex sync.Mutex
	Updates []string
}

func NewLiveEnv() *LiveEnv {
	// Return nil for WASM builds since file watching is not supported
	return nil
}

func (le *LiveEnv) Add(file string) {
	// No-op for WASM builds
}

func (le *LiveEnv) ClearUpdates() {
	if le != nil {
		le.Updates = make([]string, 0)
	}
}
