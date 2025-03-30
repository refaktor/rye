package evaldo

import (
	"testing"

	"github.com/refaktor/rye/env"
	"github.com/refaktor/rye/loader"
)

// TestRye0Benchmark runs a benchmark of the Rye0 interpreter
// To run this test with profiling:
// go test -cpuprofile cpu.prof -bench=. ./evaldo
func TestRye0Benchmark(t *testing.T) {
	// Create a Rye environment
	idx := env.NewIdxs()
	ctx := env.NewEnv(nil)
	pctx := env.NewEnv(nil)
	gen := env.NewGen()

	// Create the benchmark code
	benchmarkCode := `
	rye0 { 
		loop 10000000 { 
			_+ 1 1 
		} 
	}
	`

	// Parse the code
	block := loader.LoadStringNEW(benchmarkCode, false, &env.ProgramState{Idx: idx})

	// Set up the program state
	ps := env.NewProgramState(block.(env.Block).Series, idx)
	ps.Ctx = ctx
	ps.PCtx = pctx
	ps.Gen = gen

	// Register builtins
	RegisterBuiltins(ps)

	// Run the benchmark
	EvalBlock(ps)
}

// BenchmarkRye0 benchmarks the Rye0 interpreter
func BenchmarkRye0(b *testing.B) {
	// Create a Rye environment
	idx := env.NewIdxs()
	ctx := env.NewEnv(nil)
	pctx := env.NewEnv(nil)
	gen := env.NewGen()

	// Create the benchmark code
	benchmarkCode := `
	rye0 { 
		loop 1000 { 
			_+ 1 1 
		} 
	}
	`

	// Parse the code
	block := loader.LoadStringNEW(benchmarkCode, false, &env.ProgramState{Idx: idx})

	// Set up the program state
	ps := env.NewProgramState(block.(env.Block).Series, idx)
	ps.Ctx = ctx
	ps.PCtx = pctx
	ps.Gen = gen

	// Register builtins
	RegisterBuiltins(ps)

	// Reset the timer before the benchmark loop
	b.ResetTimer()

	// Run the benchmark
	for i := 0; i < b.N; i++ {
		// Reset the series position for each iteration
		ps.Ser.SetPos(0)
		EvalBlock(ps)
	}
}
