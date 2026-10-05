package evaldo

import (
	"math"
	"testing"
	"unicode/utf8"

	"github.com/refaktor/rye/env"
)

func TestBaseBuiltinBoundaries(t *testing.T) {
	integer := func(v int64) env.Object { return *env.NewInteger(v) }
	decimal := func(v float64) env.Object { return *env.NewDecimal(v) }
	str := func(v string) env.Object { return *env.NewString(v) }
	secret := func(v string) env.Object { return *env.NewSecret(v) }
	block := func(v ...env.Object) env.Object { return *env.NewBlock(*env.NewTSeries(v)) }
	for _, tc := range []struct {
		name, builtin string
		args          []env.Object
		want          env.Object
	}{
		{"fractional divisor", "_//", []env.Object{integer(5), decimal(0.5)}, integer(10)},
		{"truncate quotient", "_//", []env.Object{integer(5), decimal(1.9)}, integer(2)},
		{"decimal operands", "_//", []env.Object{decimal(7.9), decimal(1.9)}, integer(4)},
		{"decimal dividend", "_//", []env.Object{decimal(7.9), integer(2)}, integer(3)},
		{"negative quotient", "_//", []env.Object{integer(-5), decimal(1.9)}, integer(-2)},
		{"negative divisor", "_//", []env.Object{integer(5), decimal(-0.5)}, integer(-10)},
		{"integer precision", "_//", []env.Object{integer(math.MaxInt64), integer(1)}, integer(math.MaxInt64)},
		{"integer overflow", "_//", []env.Object{integer(math.MinInt64), integer(-1)}, nil},
		{"zero divisor", "_//", []env.Object{integer(5), decimal(0)}, nil},
		{"decimal overflow", "_//", []env.Object{decimal(0x1p63), integer(1)}, nil},
		{"minimum result", "_//", []env.Object{decimal(-0x1p63), integer(1)}, integer(math.MinInt64)},
		{"infinite quotient", "_//", []env.Object{integer(5), decimal(math.SmallestNonzeroFloat64)}, nil},
		{"nan operand", "_//", []env.Object{decimal(math.NaN()), integer(1)}, nil},
		{"infinite operand", "_//", []env.Object{integer(5), decimal(math.Inf(1))}, nil},
		{"substring unicode", "substring", []env.Object{str("café"), integer(3), integer(4)}, str("é")},
		{"substring emoji", "substring", []env.Object{str("a🙂b"), integer(1), integer(2)}, str("🙂")},
		{"substring empty", "substring", []env.Object{str(""), integer(0), integer(0)}, str("")},
		{"substring end", "substring", []env.Object{str("abc"), integer(3), integer(3)}, str("")},
		{"substring negative", "substring", []env.Object{str("abc"), integer(-1), integer(2)}, nil},
		{"substring overrun", "substring", []env.Object{str("abc"), integer(0), integer(4)}, nil},
		{"substring reversed", "substring", []env.Object{str("abc"), integer(2), integer(1)}, nil},
		{"substring huge", "substring", []env.Object{str("abc"), integer(0), integer(math.MaxInt64)}, nil},
		{"secret unicode", "substring", []env.Object{secret("café"), integer(3), integer(4)}, secret("é")},
		{"secret negative", "substring", []env.Object{secret("abc"), integer(-1), integer(2)}, nil},
		{"secret overrun", "substring", []env.Object{secret("abc"), integer(0), integer(4)}, nil},
		{"secret reversed", "substring", []env.Object{secret("abc"), integer(2), integer(1)}, nil},
		{"last unicode", "last", []env.Object{str("café")}, str("é")},
		{"last emoji", "last", []env.Object{str("🙂")}, str("🙂")},
		{"last empty", "last", []env.Object{str("")}, nil},
		{"transpose empty", "transpose", []env.Object{block()}, block()},
		{"transpose empty rows", "transpose", []env.Object{block(block(), block())}, block()},
		{"transpose short row", "transpose", []env.Object{block(block(integer(1), integer(2)), block(integer(3)))}, nil},
		{"transpose long row", "transpose", []env.Object{block(block(integer(1)), block(integer(2), integer(3)))}, nil},
		{"transpose non-block", "transpose", []env.Object{block(block(integer(1)), integer(2))}, nil},
		{"transpose rectangular", "transpose", []env.Object{block(block(integer(1), integer(2)), block(integer(3), integer(4)))}, block(block(integer(1), integer(3)), block(integer(2), integer(4)))},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ps := env.NewProgramStateOLD(*env.NewTSeries(nil), env.NewIdxs())
			bi := builtins_numbers[tc.builtin]
			if bi == nil {
				bi = builtins_string[tc.builtin]
			}
			if bi == nil {
				bi = builtins_collection[tc.builtin]
			}
			var args [5]env.Object
			copy(args[:], tc.args)
			got := bi.Fn(ps, args[0], args[1], args[2], args[3], args[4])
			if ps.ErrorFlag {
				t.Fatal("builtin set ErrorFlag instead of returning a handleable failure")
			}
			if tc.want == nil {
				if !ps.FailureFlag || got == nil || got.Type() != env.ErrorType {
					t.Fatalf("expected failure/error value, got %v (FailureFlag=%v)", got, ps.FailureFlag)
				}
				return
			}
			if ps.FailureFlag || !tc.want.Equal(got) {
				t.Fatalf("got %v (FailureFlag=%v), want %v", got, ps.FailureFlag, tc.want)
			}
			if s, ok := got.(env.String); ok && !utf8.ValidString(s.Value) {
				t.Fatal("invalid UTF-8 result")
			}
		})
	}
}
