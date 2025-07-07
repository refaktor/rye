package evaldo

import (
	"math/cmplx"

	"github.com/refaktor/rye/env"
)

var builtins_complex = map[string]*env.Builtin{

	// Tests:
	// equal { complex 3 4 |type? } 'complex
	// equal { complex 3 4 |print } "3.000000+4.000000i"
	// equal { complex 0 0 |print } "0.000000+0.000000i"
	// equal { complex -1 -2 |print } "-1.000000-2.000000i"
	// Args:
	// * real: Real part of the complex number (integer or decimal)
	// * imag: Imaginary part of the complex number (integer or decimal)
	// Returns:
	// * a new complex number with the given real and imaginary parts
	"complex": {
		Argsn: 2,
		Doc:   "Creates a complex number from real and imaginary parts.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			var real, imag float64

			// Handle real part
			switch r := arg0.(type) {
			case env.Integer:
				real = float64(r.Value)
			case env.Decimal:
				real = r.Value
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType, env.DecimalType}, "complex")
			}

			// Handle imaginary part
			switch i := arg1.(type) {
			case env.Integer:
				imag = float64(i.Value)
			case env.Decimal:
				imag = i.Value
			default:
				return MakeArgError(ps, 2, []env.Type{env.IntegerType, env.DecimalType}, "complex")
			}

			return *env.NewComplexFromParts(real, imag)
		},
	},

	// Tests:
	// equal { complex? complex 3 4 } true
	// equal { complex? 5 } false
	// equal { complex? "hello" } false
	// Args:
	// * value: Value to check
	// Returns:
	// * boolean true if the value is a complex number, false otherwise
	"complex?": {
		Argsn: 1,
		Doc:   "Checks if a value is a complex number.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			_, ok := arg0.(env.Complex)
			return *env.NewBoolean(ok)
		},
	},

	// Tests:
	// equal { real complex 3 4 } 3.0
	// equal { real complex -1.5 2.5 } -1.5
	// error { real 5 }
	// Args:
	// * value: Complex number
	// Returns:
	// * decimal value representing the real part of the complex number
	"real": {
		Argsn: 1,
		Doc:   "Returns the real part of a complex number.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch c := arg0.(type) {
			case env.Complex:
				return *env.NewDecimal(real(c.Value))
			default:
				return MakeArgError(ps, 1, []env.Type{env.ComplexType}, "real")
			}
		},
	},

	// Tests:
	// equal { imag complex 3 4 } 4.0
	// equal { imag complex -1.5 2.5 } 2.5
	// error { imag 5 }
	// Args:
	// * value: Complex number
	// Returns:
	// * decimal value representing the imaginary part of the complex number
	"imag": {
		Argsn: 1,
		Doc:   "Returns the imaginary part of a complex number.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch c := arg0.(type) {
			case env.Complex:
				return *env.NewDecimal(imag(c.Value))
			default:
				return MakeArgError(ps, 1, []env.Type{env.ComplexType}, "imag")
			}
		},
	},

	// Tests:
	// equal { complex-conj complex 3 4 |print } "3.000000-4.000000i"
	// equal { complex-conj complex 3 -4 |print } "3.000000+4.000000i"
	// error { complex-conj 5 }
	// Args:
	// * z: Complex number
	// Returns:
	// * complex number representing the complex conjugate of z
	"complex-conj": {
		Argsn: 1,
		Doc:   "Returns the complex conjugate of a complex number.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch z := arg0.(type) {
			case env.Complex:
				return *env.NewComplex(cmplx.Conj(z.Value))
			default:
				return MakeArgError(ps, 1, []env.Type{env.ComplexType}, "complex-conj")
			}
		},
	},

	// Tests:
	// equal { complex-phase complex 1 1 } 0.7853981633974483
	// equal { complex-phase complex -1 0 } 3.141592653589793
	// error { complex-phase 5 }
	// Args:
	// * z: Complex number
	// Returns:
	// * decimal value representing the phase (argument) of the complex number
	"complex-phase": {
		Argsn: 1,
		Doc:   "Returns the phase (argument) of a complex number.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch z := arg0.(type) {
			case env.Complex:
				return *env.NewDecimal(cmplx.Phase(z.Value))
			default:
				return MakeArgError(ps, 1, []env.Type{env.ComplexType}, "complex-phase")
			}
		},
	},
}
