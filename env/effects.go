package env

import "sync"

// EffectKind describes whether a builtin touches the outside world through a
// declared input, output, or both. It is an enum rather than independent bits:
// a builtin has one primary effect class.
type EffectKind uint8

const (
	EffNone   EffectKind = iota // no declared I/O effect
	EffInput                    // reads from the outside world
	EffOutput                   // writes to the outside world
	EffIO                       // both reads and writes, or not worth splitting
)

func (k EffectKind) String() string {
	switch k {
	case EffInput:
		return "input"
	case EffOutput:
		return "output"
	case EffIO:
		return "io"
	default:
		return "none"
	}
}

// EffectEntry is one recorded declared effect.
//
// Kind is either "input" or "output". Value holds the validated input value for
// inputs, or the payload for outputs. Result holds the optional stand-in value
// returned when an output effect is replayed.
type EffectEntry struct {
	Kind   string
	Name   string
	Value  Object
	Result Object
}

// EffectTape is an ordered recording of declared effects, split by kind.
//
// Inputs and outputs are consumed independently because a script may interleave
// them. Each list keeps execution order, so repeated names (loops) replay in the
// same order they were recorded.
type EffectTape struct {
	Inputs  []EffectEntry
	Outputs []EffectEntry

	InputPos  int
	OutputPos int

	Mismatches []string
}

// NextInput consumes the next recorded input. It returns false when the tape is
// exhausted or the next entry has a different name. The caller decides whether a
// miss is an error or triggers fallback to the provider.
func (t *EffectTape) NextInput(name string) (EffectEntry, bool) {
	if t == nil || t.InputPos >= len(t.Inputs) {
		return EffectEntry{}, false
	}
	entry := t.Inputs[t.InputPos]
	t.InputPos++
	if entry.Name != name {
		return EffectEntry{}, false
	}
	return entry, true
}

// NextOutput consumes the next recorded output and compares its name and payload
// with the actual call. It returns the recorded entry and whether the name and
// payload matched. The caller still decides what value the skipped effect returns.
func (t *EffectTape) NextOutput(name string, payload Object, idx *Idxs) (EffectEntry, bool, bool) {
	if t == nil || t.OutputPos >= len(t.Outputs) {
		t.addMismatch("no recorded output for '" + name + "'")
		return EffectEntry{}, false, false
	}
	entry := t.Outputs[t.OutputPos]
	t.OutputPos++

	matched := entry.Name == name
	if matched && payload != nil && entry.Value != nil {
		matched = payload.Equal(entry.Value)
	} else if payload == nil || entry.Value == nil {
		matched = payload == nil && entry.Value == nil
	}

	if !matched {
		expected := "<nil>"
		if entry.Value != nil {
			expected = entry.Value.Dump(*idx)
		}
		actual := "<nil>"
		if payload != nil {
			actual = payload.Dump(*idx)
		}
		t.addMismatch("output '" + name + "': expected " + expected + ", got " + actual)
	}
	return entry, true, matched
}

// RemainingOutputs returns the output entries that were never consumed.
func (t *EffectTape) RemainingOutputs() []EffectEntry {
	if t == nil || t.OutputPos >= len(t.Outputs) {
		return nil
	}
	return t.Outputs[t.OutputPos:]
}

func (t *EffectTape) addMismatch(msg string) {
	if t == nil {
		return
	}
	t.Mismatches = append(t.Mismatches, msg)
}

// EffectRecorder collects effects while a script runs. Flush serializes them in
// execution order (see evaldo/effects_tape.go).
type EffectRecorder struct {
	mu      sync.Mutex
	Entries []EffectEntry
}

// Add appends an effect entry in execution order.
func (r *EffectRecorder) Add(entry EffectEntry) {
	if r == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Entries = append(r.Entries, entry)
}

// Snapshot returns a copy of the recorded entries.
func (r *EffectRecorder) Snapshot() []EffectEntry {
	if r == nil {
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]EffectEntry, len(r.Entries))
	copy(out, r.Entries)
	return out
}
