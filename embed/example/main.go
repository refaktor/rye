//go:build ignore
// +build ignore

// Example program showing how to embed Rye in a Go application using the
// minimal embed sub-module.
//
// Build & run:
//
//	cd embed/example
//	go run -tags "no_persistent no_table no_vector" main.go
package main

import (
	"fmt"
	"os"

	"github.com/refaktor/rye/embed"
	"github.com/refaktor/rye/env"
)

// ---------------------------------------------------------------------------
// 1. Register a simple Go function as a Rye builtin
// ---------------------------------------------------------------------------

func multiply(ps *env.ProgramState, a0, a1, a2, a3, a4 env.Object) env.Object {
	if a0.Type() != env.IntegerType || a1.Type() != env.IntegerType {
		return *env.NewError("multiply: expected two integers")
	}
	x := a0.(env.Integer).Value
	y := a1.(env.Integer).Value
	return *env.NewInteger(x * y)
}

func greet(ps *env.ProgramState, a0, a1, a2, a3, a4 env.Object) env.Object {
	if a0.Type() != env.StringType {
		return *env.NewError("greet: expected a string")
	}
	name := a0.(env.String).Value
	return *env.NewString("Hello, " + name + "!")
}

func main() {
	// ---------------------------------------------------------------------------
	// 2. Create an engine and register builtins
	// ---------------------------------------------------------------------------
	engine := embed.New()

	engine.RegisterBuiltinDoc("multiply", 2, "multiply a b — returns a × b", multiply)
	engine.RegisterBuiltinDoc("greet", 1, "greet name — returns a greeting string", greet)

	// ---------------------------------------------------------------------------
	// 3. Inject a Go value as a Rye word
	// ---------------------------------------------------------------------------
	engine.SetWord("app-name", *env.NewString("MyApp"))

	// ---------------------------------------------------------------------------
	// 4. Evaluate Rye code and read back results
	// ---------------------------------------------------------------------------
	ryeCode := `
		; use the injected word
		greeting: greet app-name

		; arithmetic
		a: 6
		b: 7
		product: multiply a b

		; native Rye arithmetic
		total: a + b + product

		; return a block so we can inspect values
		total
	`

	result, err := engine.Eval(ryeCode)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Rye error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("total =", result) // 55

	// Read the greeting back as a typed Go string
	greeting, err := engine.EvalString(`greet "World"`)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Rye error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(greeting) // Hello, World!

	// Read an integer result
	n, err := engine.EvalInteger("42")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Rye error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("integer =", n) // 42

	// ---------------------------------------------------------------------------
	// 5. Low-level: work with env.Object directly
	// ---------------------------------------------------------------------------
	obj, err := engine.EvalGetObject(`{ 1 2 3 }`)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Rye error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("block type: %T\n", obj) // env.Block

	// ---------------------------------------------------------------------------
	// 6. Use GetWord to read back a word that was set inside a script
	// ---------------------------------------------------------------------------
	engine.Eval(`answer: 21 + 21`)
	if val, ok := engine.GetWord("answer"); ok {
		fmt.Println("answer =", val.Print(*engine.ProgramState().Idx)) // 42
	}

	fmt.Println("Done.")
}
