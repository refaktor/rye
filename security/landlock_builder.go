package security

import (
	"fmt"
	"os"
	"path/filepath"

	ll "github.com/landlock-lsm/go-landlock/landlock/syscall"
	"github.com/landlock-lsm/go-landlock/landlock"
)

type LandlockBuilder struct {
	paths    []string // rw allow list
	execPaths []string // execute allow list
	toCwd    bool
	applied  bool
}

func NewLandlockBuilder() *LandlockBuilder { return &LandlockBuilder{} }

func (b *LandlockBuilder) LimitToCwd() { b.toCwd = true }

func (b *LandlockBuilder) AllowPath(p string) { b.paths = append(b.paths, p) }
func (b *LandlockBuilder) AllowExec(p string) { b.execPaths = append(b.execPaths, p) }

func (b *LandlockBuilder) Enforce() error {
	if b.applied {
		return fmt.Errorf("landlock already enforced")
	}
	// Configure desired access rights (execute + read/write). Not directly used with helpers,
	// but kept for future extension when using low-level rule construction.
	_ = landlock.MustConfig(landlock.AccessFSSet(ll.AccessFSExecute|ll.AccessFSReadFile|ll.AccessFSWriteFile|ll.AccessFSReadDir)).BestEffort()
	// Limit to CWD: allow RW within CWD (dirs and files)
	if b.toCwd {
		base := "."
		if wd, err := os.Getwd(); err == nil { base = wd }
		if abs, err := filepath.Abs(base); err == nil { base = abs }
		if err := landlock.V1.BestEffort().RestrictPaths(landlock.RWDirs(base), landlock.RWFiles(base)); err != nil { return err }
	}
	// RW allow list
	for _, p := range b.paths {
		if fi, err := os.Stat(p); err == nil && fi.IsDir() {
			if err := landlock.V1.BestEffort().RestrictPaths(landlock.RWDirs(p), landlock.RWFiles(p)); err != nil { return err }
		} else {
			if err := landlock.V1.BestEffort().RestrictPaths(landlock.RWFiles(p)); err != nil { return err }
		}
	}
	// Exec allow list: grant execute at directory level (parent if file)
	for _, p := range b.execPaths {
		dir := p
		if fi, err := os.Stat(p); err == nil && !fi.IsDir() {
			dir = filepath.Dir(p)
		}
		// Permit execute by allowing read access set (includes execute in accessFSRead) on dir
		if err := landlock.V1.BestEffort().RestrictPaths(landlock.RODirs(dir)); err != nil { return err }
	}
	b.applied = true
	return nil
}
