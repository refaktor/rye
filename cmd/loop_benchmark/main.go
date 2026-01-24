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
	// Define command line flags
	ryeFlag := flag.Bool("rye", false, "Run only the standard evaluator")
	rye0Flag := flag.Bool("rye0", false, "Run only the standard evaluator")
	vmFlag := flag.Bool("vm", false, "Run only the VM evaluator")
	cpuProfile := flag.String("cpuprofile", "", "Write cpu profile to file")
	memProfile := flag.String("memprofile", "", "Write memory profile to file")
	flag.Parse()

	// Start CPU profiling if requested
	if *cpuProfile != "" {
		f, err := os.Create(*cpuProfile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not create CPU profile: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()
		if err := pprof.StartCPUProfile(f); err != nil {
			fmt.Fprintf(os.Stderr, "Could not start CPU profile: %v\n", err)
			os.Exit(1)
		}
		defer pprof.StopCPUProfile()
	}

	// Create a simple program that adds numbers in a loop
	idx := env.NewIdxs()

	// Create a context with the necessary functions
	ctx := env.NewEnv(nil)

	// Get indices for words we'll need
	addIdx, _ := idx.GetIndex("add")
	xIdx, _ := idx.GetIndex("x")

	// Create a builtin for addition
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
	// Create a builtin for addition

	add00Builtin := env.NewBuiltin(
		func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			if arg0 == nil || arg1 == nil {
				// return env.NewError("add requires two arguments")
			}

			// Handle integer addition
			//if arg0.Type() == env.IntegerType && arg1.Type() == env.IntegerType {
			return *env.NewInteger(123 + 123)
			// }

			//return env.NewError("unsupported types for addition")
		},
		0,                // argsn
		false,            // acceptFailure
		true,             // pure
		"add two values", // doc
	)
	fmt.Println(addBuiltin)
	fmt.Println(add00Builtin)

	// Set the builtins in the context
	ctx.Set(addIdx, *addBuiltin)

	// Number of iterations
	iterations := int64(10000000)
	fmt.Printf("Running benchmark with %d iterations...\n", iterations)

	var standardResult env.Object
	var standardDuration time.Duration
	var fastResult env.Object
	var fastDuration time.Duration

	// Run standard evaluator if requested or if no specific flag is set
	if *ryeFlag || (!*ryeFlag && !*vmFlag) {
		fmt.Println("Running with standard evaluator...")
		evaldo.DisableFastEvaluator() // Make sure it's disabled

		standardStart := time.Now()

		// Create a simple series with just a single integer
		standardSeries := *env.NewTSeries([]env.Object{
			// TEST adding builtin directly into block and evaluating that, no lookup
			// *env.NewModword(xIdx),
			*addBuiltin,
			*env.NewInteger(123),
			*env.NewInteger(123),
		})

		// Create a new program state for each iteration
		ps := env.NewProgramState(standardSeries, idx)
		ps.Ctx = ctx
		ps.Dialect = env.Rye2Dialect

		//i = env.Integer{0}
		// Run the loop with the standard evaluator
		for i := int64(0); i < iterations; i++ {
			// Reset the series position
			ps.Ser.SetPos(0)

			// Evaluate the block
			evaldo.EvalBlockInj(ps, nil, true)
			//evaldo.Rye0_EvalBlockInj(ps, nil, false)
		}

		// Store the result
		standardResult = ps.Res

		standardDuration = time.Since(standardStart)
		fmt.Printf("Standard evaluator result: %v (took %v)\n", standardResult.Print(*idx), standardDuration)
	}

	// Run standard evaluator if requested or if no specific flag is set
	if *rye0Flag || (!*rye0Flag && !*vmFlag) {
		fmt.Println("Running with rye0 evaluator...")
		evaldo.DisableFastEvaluator() // Make sure it's disabled

		standardStart := time.Now()

		// Create a simple series with just a single integer
		standardSeries := *env.NewTSeries([]env.Object{
			// TEST adding builtin directly into block and evaluating that, no lookup
			//			*env.NewModword(xIdx),
			*addBuiltin,
			*env.NewInteger(123),
			*env.NewInteger(123),
		})

		// Create a new program state for each iteration
		ps := env.NewProgramState(standardSeries, idx)
		ps.Ctx = ctx
		ps.Dialect = env.Rye0Dialect

		// Run the loop with the standard evaluator
		for i := int64(0); i < iterations; i++ {
			// Reset the series position
			ps.Ser.SetPos(0)

			// Evaluate the block
			evaldo.Rye0_EvalBlockInj(ps, nil, false)
		}

		// Store the result
		standardResult = ps.Res

		standardDuration = time.Since(standardStart)
		fmt.Printf("Standard evaluator result: %v (took %v)\n", standardResult.Print(*idx), standardDuration)
	}

	// Run VM evaluator if requested or if no specific flag is set
	if *vmFlag || (!*ryeFlag && !*vmFlag) {
		fmt.Println("Running with fast evaluator...")
		evaldo.EnableFastEvaluator() // Enable it

		fastStart := time.Now()

		// Create a simple series with just a single integer
		series := *env.NewTSeries([]env.Object{
			*env.NewSetword(xIdx),
			*env.NewInteger(123),
			*env.NewInteger(123),
			*env.NewWord(addIdx),
		})

		// Compile the block once
		psCompile := env.NewProgramState(series, idx)
		psCompile.Ctx = ctx
		psCompile.Dialect = env.Rye0Dialect
		program := evaldo.Rye0_CompileBlock(psCompile)

		// Create a single VM and reuse it
		vm := evaldo.NewRye0VM(ctx, nil, nil, idx)

		// Run the loop with the compiled program
		for i := int64(0); i < iterations; i++ {
			// Reset the VM state
			vm.Sp = 0

			// Execute the program
			result, err := vm.Execute(program)
			if err != nil {
				fmt.Println("Error in fast evaluator:", err)
				return
			}
			fastResult = result
		}

		fastDuration = time.Since(fastStart)
		fmt.Printf("Fast evaluator result: %v (took %v)\n", fastResult.Print(*idx), fastDuration)
	}

	// Print comparison if both evaluators were run
	if *ryeFlag && *vmFlag || (!*ryeFlag && !*vmFlag) {
		fmt.Printf("Speed improvement: %.2fx\n", float64(standardDuration)/float64(fastDuration))
	}

	// Write memory profile if requested
	if *memProfile != "" {
		f, err := os.Create(*memProfile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not create memory profile: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()
		if err := pprof.WriteHeapProfile(f); err != nil {
			fmt.Fprintf(os.Stderr, "Could not write memory profile: %v\n", err)
			os.Exit(1)
		}
	}
}
