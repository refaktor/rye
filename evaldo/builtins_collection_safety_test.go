package evaldo

import (
	"math"
	"reflect"
	"testing"

	"github.com/refaktor/rye/env"
)

func TestCollectionSafety(t *testing.T) {
	i := func(v int64) env.Object { return *env.NewInteger(v) }
	b := func(v ...env.Object) env.Object { return *env.NewBlock(*env.NewTSeries(v)) }
	l := func(v ...any) env.Object { return *env.NewList(v) }
	for _, tc := range []struct {
		name, builtin string
		args          []env.Object
		want          env.Object
	}{
		{"random zero", "random\\integer", []env.Object{i(0)}, nil},
		{"random negative", "random\\integer", []env.Object{i(-1)}, nil},
		{"random one", "random\\integer", []env.Object{i(1)}, i(0)},
		{"decimal zero", "random\\decimal", []env.Object{*env.NewDecimal(0)}, nil},
		{"decimal negative", "random\\decimal", []env.Object{*env.NewDecimal(-1)}, nil},
		{"decimal int zero", "random\\decimal", []env.Object{i(0)}, nil},
		{"decimal int negative", "random\\decimal", []env.Object{i(-1)}, nil},
		{"decimal nan", "random\\decimal", []env.Object{*env.NewDecimal(math.NaN())}, nil},
		{"decimal inf", "random\\decimal", []env.Object{*env.NewDecimal(math.Inf(1))}, nil},
		{"range reversed", "range", []env.Object{i(5), i(1)}, nil},
		{"range adjacent reversed", "range", []env.Object{i(2), i(1)}, nil},
		{"range max", "range", []env.Object{i(math.MaxInt64), i(math.MaxInt64)}, b(i(math.MaxInt64))},
		{"range near max", "range", []env.Object{i(math.MaxInt64 - 1), i(math.MaxInt64)}, b(i(math.MaxInt64-1), i(math.MaxInt64))},
		{"range min", "range", []env.Object{i(math.MinInt64), i(math.MinInt64 + 1)}, b(i(math.MinInt64), i(math.MinInt64+1))},
		{"range overflow", "range", []env.Object{i(math.MinInt64), i(math.MaxInt64)}, nil},
		{"range allocation limit", "range", []env.Object{i(0), i(maxRangeElements)}, nil},
		{"range crosses zero", "range", []env.Object{i(-1), i(1)}, b(i(-1), i(0), i(1))},
		{"rest block negative", "rest\\from", []env.Object{b(i(1)), i(-1)}, nil},
		{"rest list negative", "rest\\from", []env.Object{l(1), i(-1)}, nil},
		{"rest string negative", "rest\\from", []env.Object{*env.NewString("abc"), i(-1)}, nil},
		{"rest secret negative", "rest\\from", []env.Object{*env.NewSecret("abc"), i(-1)}, nil},
		{"rest vector negative", "rest\\from", []env.Object{*env.NewVector([]float64{1}), i(-1)}, nil},
		{"rest matrix negative", "rest\\from", []env.Object{*env.NewMatrixWithData(1, 1, []float64{1}), i(-1)}, nil},
		{"rest huge index", "rest\\from", []env.Object{l(1), i(math.MaxInt64)}, nil},
		{"rest list value", "rest\\from", []env.Object{l(1, 2, 3), i(1)}, l(2, 3)},
		{"unique stable", "unique", []env.Object{l(3, 1, 2, 3, 1)}, l(3, 1, 2)},
		{"unique blocks", "unique", []env.Object{b(b(i(1)), b(i(1)), b(i(2)))}, b(b(i(1)), b(i(2)))},
		{"unique lists", "unique", []env.Object{l(l(1), l(1), l(2))}, l(l(1), l(2))},
		{"unique raw lists", "unique", []env.Object{l([]any{1}, []any{1}, []any{2})}, l([]any{1}, []any{2})},
		{"unique raw dicts", "unique", []env.Object{l(map[string]any{"a": []any{1}}, map[string]any{"a": []any{1}}, map[string]any{"b": []any{1}})}, l(map[string]any{"a": []any{1}}, map[string]any{"b": []any{1}})},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ps := env.NewProgramStateOLD(*env.NewTSeries(nil), env.NewIdxs())
			bi := builtins_collection[tc.builtin]
			if bi == nil {
				bi = builtins_numbers[tc.builtin]
			}
			var args [5]env.Object
			copy(args[:], tc.args)
			got := bi.Fn(ps, args[0], args[1], args[2], args[3], args[4])
			if ps.ErrorFlag {
				t.Fatal("expected handleable failure, not ErrorFlag")
			}
			if tc.want == nil {
				if !ps.FailureFlag || got == nil || got.Type() != env.ErrorType {
					t.Fatalf("expected failure, got %v", got)
				}
				return
			}
			if ps.FailureFlag || !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("got %#v, want %#v (failure=%v)", got, tc.want, ps.FailureFlag)
			}
		})
	}
}

func TestRandomDecimalBounds(t *testing.T) {
	for _, bound := range []float64{math.SmallestNonzeroFloat64, 0.5, 1, math.MaxFloat64} {
		ps := env.NewProgramStateOLD(*env.NewTSeries(nil), env.NewIdxs())
		for n := 0; n < 100; n++ {
			got := builtins_numbers["random\\decimal"].Fn(ps, *env.NewDecimal(bound), nil, nil, nil, nil)
			value, ok := got.(env.Decimal)
			if !ok || ps.FailureFlag || value.Value < 0 || value.Value >= bound {
				t.Fatalf("bound=%v, got=%v", bound, got)
			}
		}
	}
}
