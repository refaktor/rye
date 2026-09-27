// Package finder extends go-find with result filters used by Rye's OS builtins.
package finder

import (
	"os"

	upstream "github.com/refaktor/go-find"
)

// Find retains go-find's traversal and built-in predicates. Additional filters
// select results after traversal; they do not prune directory traversal.
type Find struct {
	*upstream.Find
	filters []func(string, os.FileInfo) bool
}

func NewFind(paths ...string) *Find {
	return &Find{Find: upstream.NewFind(paths...)}
}

func (finder *Find) FilterFunc(predicate func(string, os.FileInfo) bool) {
	finder.filters = append(finder.filters, predicate)
}

func (finder *Find) Evaluate() ([]string, error) {
	paths, err := finder.Find.Evaluate()
	if err != nil || len(finder.filters) == 0 {
		return paths, err
	}
	results := make([]string, 0, len(paths))
	for _, path := range paths {
		// Match filepath.Walk's metadata semantics, including for symbolic links.
		info, err := os.Lstat(path)
		if err != nil {
			return nil, err
		}
		keep := true
		for _, predicate := range finder.filters {
			if !predicate(path, info) {
				keep = false
				break
			}
		}
		if keep {
			results = append(results, path)
		}
	}
	return results, nil
}
