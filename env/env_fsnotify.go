//go:build !wasm && !b_wasm

package env

import (
	"fmt"
	"sync"

	"github.com/fsnotify/fsnotify"
)

// LiveEnv -- a experiment in live realoading

type LiveEnv struct {
	Active  bool
	Watcher *fsnotify.Watcher
	PsMutex sync.Mutex
	Updates []string
}

func NewLiveEnv() *LiveEnv {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		// TODO -- temporary removed for WASM and we don't use this at the moment ... solve at build time with flags
		// fmt.Println("Error creating watcher:", err) // TODO -- if this fails show error in red, but make it so that rye runs anyway (check if null at repl for starters)
		return nil
	}

	// defer watcher.Close()

	// Watch current directory for changes in any Go source file (*.go)

	liveEnv := &LiveEnv{true, watcher, sync.Mutex{}, make([]string, 0)}

	go func() {
		for {
			select {
			case event := <-watcher.Events:
				if event.Op&fsnotify.Write == fsnotify.Write {
					// fmt.Println("LiveEnv file changed:", event.Name)
					liveEnv.PsMutex.Lock()
					liveEnv.Updates = append(liveEnv.Updates, event.Name)
					liveEnv.PsMutex.Unlock()
				}
			case err := <-watcher.Errors:
				fmt.Println("LiveEnv error watching files:", err)
			}
		}
	}()

	return liveEnv
}

func (le *LiveEnv) Add(file string) {
	err := le.Watcher.Add(".")
	if err != nil {
		fmt.Println("LiveEnv: Error adding directory to watch:", err)
	}
}

func (le *LiveEnv) ClearUpdates() {
	le.Updates = make([]string, 0)
}
