// Command loop_benchmark measures the supported Rye evaluator.
package main

import (
	"flag"
	"fmt"
	"os"
	"runtime/pprof"
	"time"

	"github.com/refaktor/rye/env"
	"github.com/refaktor/rye/evaldo"
)

func main() {
	iterations := flag.Int("iterations", 10000000, "Number of evaluations")
	cpuProfile := flag.String("cpuprofile", "", "Write CPU profile")
	memProfile := flag.String("memprofile", "", "Write heap profile")
	flag.Parse()
	if *iterations < 1 {
		fmt.Fprintln(os.Stderr, "iterations must be positive")
		os.Exit(2)
	}
	if *cpuProfile != "" {
		f, err := os.Create(*cpuProfile)
		if err != nil {
			panic(err)
		}
		defer f.Close()
		if err := pprof.StartCPUProfile(f); err != nil {
			panic(err)
		}
		defer pprof.StopCPUProfile()
	}
	idx := env.NewIdxs()
	ps := env.NewProgramStateOLD(*env.NewTSeries(nil), idx)
	add := env.NewBuiltin(func(_ *env.ProgramState, a, b, _, _, _ env.Object) env.Object {
		return *env.NewInteger(a.(env.Integer).Value + b.(env.Integer).Value)
	}, 2, false, true, "benchmark addition")
	ps.Ser = *env.NewTSeries([]env.Object{*add, *env.NewInteger(123), *env.NewInteger(123)})
	start := time.Now()
	for i := 0; i < *iterations; i++ {
		ps.Ser.SetPos(0)
		evaldo.EvalBlockInj(ps, nil, false)
		if ps.ErrorFlag || ps.FailureFlag {
			panic(ps.Res)
		}
	}
	fmt.Printf("%d evaluations: %s (%s)\n", *iterations, ps.Res.Print(*idx), time.Since(start))
	if *memProfile != "" {
		f, err := os.Create(*memProfile)
		if err != nil {
			panic(err)
		}
		defer f.Close()
		if err := pprof.WriteHeapProfile(f); err != nil {
			panic(err)
		}
	}
}
