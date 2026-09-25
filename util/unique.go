package util

import (
	"reflect"

	"github.com/refaktor/rye/env"
)

// RemoveDuplicateValues preserves first-occurrence order and original values.
// Comparable scalars retain the existing exact-key semantics and linear fast
// path. Collections use structural comparison rather than unsafe Go map keys.
func RemoveDuplicateValues[T any](values []T) []T {
	result := make([]T, 0, len(values))
	seen := make(map[any]struct{})
	var collections []T
	for _, value := range values {
		rv := reflect.ValueOf(value)
		if !rv.IsValid() || rv.Comparable() {
			if _, ok := seen[value]; ok {
				continue
			}
			seen[value] = struct{}{}
		} else {
			duplicate := false
			for _, previous := range collections {
				if sameUniqueValue(previous, value) {
					duplicate = true
					break
				}
			}
			if duplicate {
				continue
			}
			collections = append(collections, value)
		}
		result = append(result, value)
	}
	return result
}

func sameUniqueValue(a, b any) bool {
	switch a := a.(type) {
	case env.Block:
		b, ok := b.(env.Block)
		if !ok || a.Mode != b.Mode || a.Series.Len() != b.Series.Len() {
			return false
		}
		// Source locations and cursor positions are not part of block value equality.
		for i := 0; i < a.Series.Len(); i++ {
			if !sameUniqueValue(a.Series.Get(i), b.Series.Get(i)) {
				return false
			}
		}
		return true
	case env.List:
		b, ok := b.(env.List)
		return ok && a.Kind.Equal(b.Kind) && sameUniqueValue(a.Data, b.Data)
	case env.Dict:
		b, ok := b.(env.Dict)
		return ok && a.Kind.Equal(b.Kind) && sameUniqueValue(a.Data, b.Data)
	case []any:
		b, ok := b.([]any)
		if !ok || len(a) != len(b) {
			return false
		}
		for i := range a {
			if !sameUniqueValue(a[i], b[i]) {
				return false
			}
		}
		return true
	case map[string]any:
		b, ok := b.(map[string]any)
		if !ok || len(a) != len(b) {
			return false
		}
		for key, value := range a {
			other, exists := b[key]
			if !exists || !sameUniqueValue(value, other) {
				return false
			}
		}
		return true
	default:
		return reflect.DeepEqual(a, b)
	}
}
