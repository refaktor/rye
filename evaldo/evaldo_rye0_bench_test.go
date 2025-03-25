package evaldo_test

import (
	"testing"

	"github.com/refaktor/rye/env"
	"github.com/refaktor/rye/evaldo"
)

// setupRye0CallBuiltinBenchmark creates a program state for benchmarking Rye0_CallBuiltin
func setupRye0CallBuiltinBenchmark() (*env.ProgramState, *env.Idxs, env.Builtin) {
	// Create indices and context
	idx := env.NewIdxs()
	ctx := env.NewEnv(nil)

	// Create a simple builtin function for addition
	addBuiltin := env.NewBuiltin(
		func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			if arg0 == nil || arg1 == nil {
				return env.NewError("add requires two arguments")
			}

			// Handle integer addition
			if arg0.Type() == env.IntegerType && arg1.Type() == env.IntegerType {
				return *env.NewInteger(arg0.(env.Integer).Value + arg1.(env.Integer).Value)
			}

			return env.NewError("unsupported types for addition")
		},
		2,                // argsn
		false,            // acceptFailure
		true,             // pure
		"add two values", // doc
	)

	// Create a series with two integers
	objects := make([]env.Object, 0, 2)
	objects = append(objects, *env.NewInteger(10))
	objects = append(objects, *env.NewInteger(20))

	series := *env.NewTSeries(objects)

	// Create a program state
	ps := env.NewProgramState(series, idx)
	ps.Ctx = ctx
	ps.AllowMod = false

	return ps, idx, *addBuiltin
}

// BenchmarkRye0_CallBuiltin benchmarks the Rye0_CallBuiltin function
func BenchmarkRye0_CallBuiltin(b *testing.B) {
	ps, _, addBuiltin := setupRye0CallBuiltinBenchmark()

	// Reset the timer before the benchmark loop
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Reset the series position for each iteration
		ps.Ser.SetPos(0)

		// Call the builtin function with no pre-provided arguments
		evaldo.Rye0_CallBuiltin(addBuiltin, ps, nil, false, false, nil)
	}
}

// BenchmarkRye0_CallBuiltinWithCurried benchmarks the Rye0_CallBuiltin function with curried arguments
func BenchmarkRye0_CallBuiltinWithCurried(b *testing.B) {
	ps, _, addBuiltin := setupRye0CallBuiltinBenchmark()

	// Set curried values
	addBuiltin.Cur0 = *env.NewInteger(10)
	addBuiltin.Cur1 = *env.NewInteger(20)

	// Reset the timer before the benchmark loop
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Call the builtin function with curried arguments
		evaldo.Rye0_CallBuiltin(addBuiltin, ps, nil, false, false, nil)
	}
}
