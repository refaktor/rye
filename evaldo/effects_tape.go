package evaldo

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/refaktor/rye/env"
	"github.com/refaktor/rye/loader"
)

// LoadEffectTape loads an effect tape file. The file is parsed as Rye data and
// walked token by token; it is never evaluated.
func LoadEffectTape(ps *env.ProgramState, path string) (*env.EffectTape, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, ok := loader.LoadString(string(content), false, ps).(env.Block)
	if !ok {
		return nil, fmt.Errorf("effect tape %s is not a block", path)
	}
	tokens := block.Series.GetAll()
	tape := &env.EffectTape{}

	i := 0
	// Optional header: rye-effects 1
	if i < len(tokens) && tapeWordName(ps, tokens[i]) == "rye-effects" {
		i += 2
	}

	for i < len(tokens) {
		kind := tapeWordName(ps, tokens[i])
		if kind != "input" && kind != "output" {
			i++
			continue
		}
		i++
		if i >= len(tokens) {
			return nil, fmt.Errorf("effect tape %s: missing name after %s", path, kind)
		}
		name := tapeName(ps, tokens[i])
		if name == "" {
			return nil, fmt.Errorf("effect tape %s: invalid %s name", path, kind)
		}
		i++
		if i >= len(tokens) {
			return nil, fmt.Errorf("effect tape %s: missing value for %s '%s'", path, kind, name)
		}
		value, next, err := parseTapeValue(ps, tokens, i)
		if err != nil {
			return nil, fmt.Errorf("effect tape %s: %s '%s': %w", path, kind, name, err)
		}
		i = next

		entry := env.EffectEntry{Kind: kind, Name: name, Value: value}
		if kind == "output" && i < len(tokens) {
			nextKind := tapeWordName(ps, tokens[i])
			if nextKind != "input" && nextKind != "output" {
				result, next, err := parseTapeValue(ps, tokens, i)
				if err != nil {
					return nil, fmt.Errorf("effect tape %s: stand-in for output '%s': %w", path, name, err)
				}
				i = next
				entry.Result = result
			}
		}

		if kind == "input" {
			tape.Inputs = append(tape.Inputs, entry)
		} else {
			tape.Outputs = append(tape.Outputs, entry)
		}
	}
	return tape, nil
}

func tapeWordName(ps *env.ProgramState, obj env.Object) string {
	switch o := obj.(type) {
	case env.Word:
		return ps.Idx.GetWord(o.Index)
	case env.Tagword:
		return ps.Idx.GetWord(o.Index)
	case env.String:
		return o.Value
	default:
		return ""
	}
}

func tapeName(ps *env.ProgramState, obj env.Object) string {
	switch o := obj.(type) {
	case env.Word:
		return ps.Idx.GetWord(o.Index)
	case env.Tagword:
		return ps.Idx.GetWord(o.Index)
	case env.String:
		return o.Value
	default:
		return ""
	}
}

func parseTapeValue(ps *env.ProgramState, tokens []env.Object, i int) (env.Object, int, error) {
	if i >= len(tokens) {
		return nil, i, fmt.Errorf("missing value")
	}
	obj := tokens[i]
	i++

	// Compound values emitted by Object.Dump: dict { ... }, list { ... },
	// context { ... }. The keyword itself must not be evaluated.
	if name := tapeWordName(ps, obj); name != "" {
		switch name {
		case "true":
			return *env.NewBoolean(true), i, nil
		case "false":
			return *env.NewBoolean(false), i, nil
		case "dict":
			if i >= len(tokens) {
				return nil, i, fmt.Errorf("dict without block")
			}
			blk, ok := tokens[i].(env.Block)
			if !ok {
				return nil, i, fmt.Errorf("dict without block")
			}
			i++
			items := blk.Series.GetAll()
			data := make(map[string]any)
			j := 0
			for j < len(items) {
				key := tapeName(ps, items[j])
				if key == "" {
					return nil, i, fmt.Errorf("dict key is not a word or string")
				}
				j++
				if j >= len(items) {
					return nil, i, fmt.Errorf("dict key %q has no value", key)
				}
				val, next, err := parseTapeValue(ps, items, j)
				if err != nil {
					return nil, i, err
				}
				j = next
				data[key] = val
			}
			return *env.NewDict(data), i, nil
		case "list":
			if i >= len(tokens) {
				return nil, i, fmt.Errorf("list without block")
			}
			blk, ok := tokens[i].(env.Block)
			if !ok {
				return nil, i, fmt.Errorf("list without block")
			}
			i++
			items := blk.Series.GetAll()
			data := make([]any, 0, len(items))
			j := 0
			for j < len(items) {
				val, next, err := parseTapeValue(ps, items, j)
				if err != nil {
					return nil, i, err
				}
				j = next
				data = append(data, val)
			}
			return *env.NewList(data), i, nil
		case "context":
			if i >= len(tokens) {
				return nil, i, fmt.Errorf("context without block")
			}
			blk, ok := tokens[i].(env.Block)
			if !ok {
				return nil, i, fmt.Errorf("context without block")
			}
			i++
			ctx := env.NewEnv(nil)
			items := blk.Series.GetAll()
			j := 0
			for j < len(items) {
				sw, ok := items[j].(env.Setword)
				if !ok {
					return nil, i, fmt.Errorf("context item is not a set-word")
				}
				j++
				if j >= len(items) {
					return nil, i, fmt.Errorf("context word %s has no value", ps.Idx.GetWord(sw.Index))
				}
				val, next, err := parseTapeValue(ps, items, j)
				if err != nil {
					return nil, i, err
				}
				j = next
				ctx.Set(sw.Index, val)
			}
			return ctx, i, nil
		}
	}
	return obj, i, nil
}

// FlushEffectRecord writes all recorded effects to path.
func FlushEffectRecord(ps *env.ProgramState, path string) error {
	if ps.EffectMode.Record == nil {
		return nil
	}
	entries := ps.EffectMode.Record.Snapshot()
	var b strings.Builder
	b.WriteString("rye-effects 1\n")
	b.WriteString("; recorded " + time.Now().Format(time.RFC3339) + "\n")
	for _, entry := range entries {
		name := dumpTapeName(entry.Name)
		if entry.Value == nil {
			b.WriteString(entry.Kind + " " + name + " nil\n")
			continue
		}
		b.WriteString(entry.Kind + " " + name + " " + dumpEffectValue(entry.Value, ps.Idx))
		if entry.Kind == "output" && entry.Result != nil {
			b.WriteString(" " + dumpEffectValue(entry.Result, ps.Idx))
		}
		b.WriteString("\n")
	}
	return os.WriteFile(path, []byte(b.String()), 0644)
}

func dumpTapeName(name string) string {
	if name == "" {
		return `""`
	}
	for _, r := range name {
		if !(r == '-' || r == '_' || r == '?' || r == '!' || r == '\'' ||
			(r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')) {
			return strconv.Quote(name)
		}
	}
	return "'" + name
}

// FinishEffects flushes a recording, reports replay differences, and prints the
// dry-run summary. It returns a process exit code (0 or 1).
func FinishEffects(ps *env.ProgramState) int {
	code := 0
	if ps.EffectMode.Record != nil && ps.EffectMode.RecordPath != "" {
		if err := FlushEffectRecord(ps, ps.EffectMode.RecordPath); err != nil {
			fmt.Fprintf(effectLogWriter(ps), "effect record: %v\n", err)
			code = 1
		}
	}

	if tape := ps.EffectMode.Replay; tape != nil {
		for _, msg := range tape.Mismatches {
			fmt.Fprintf(effectLogWriter(ps), "replay mismatch: %s\n", msg)
			code = 1
		}
		if remaining := tape.RemainingOutputs(); len(remaining) > 0 {
			for _, entry := range remaining {
				fmt.Fprintf(effectLogWriter(ps), "replay mismatch: recorded output '%s' was not produced\n", entry.Name)
			}
			code = 1
		}
	}

	if ps.EffectMode.DryRun {
		fmt.Fprintf(effectLogWriter(ps), "declared effects: %d inputs, %d outputs skipped (dry-run)\n",
			ps.EffectMode.Inputs, ps.EffectMode.OutputsSkipped)
	}
	return code
}

// dumpEffectValue serializes effect values. Dict and List values may hold raw Go
// values when they came from a builtin, so normalize those before Dump.
func dumpEffectValue(obj env.Object, idx *env.Idxs) string {
	if obj == nil {
		return "nil"
	}
	switch v := obj.(type) {
	case env.Dict:
		keys := make([]string, 0, len(v.Data))
		for k := range v.Data {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		var b strings.Builder
		b.WriteString("dict { ")
		for _, k := range keys {
			b.WriteString(k)
			b.WriteString(" ")
			b.WriteString(dumpEffectValue(env.ToRyeValue(v.Data[k]), idx))
			b.WriteString(" ")
		}
		b.WriteString("}")
		return b.String()
	case env.List:
		var b strings.Builder
		b.WriteString("list { ")
		for _, raw := range v.Data {
			b.WriteString(dumpEffectValue(env.ToRyeValue(raw), idx))
			b.WriteString(" ")
		}
		b.WriteString("}")
		return b.String()
	default:
		return obj.Dump(*idx)
	}
}
