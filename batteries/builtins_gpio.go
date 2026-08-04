//go:build add_gpio && !add_gpio_sim
// +build add_gpio,!add_gpio_sim

package batteries

import (
	"fmt"

	"github.com/refaktor/rye/env"
	"github.com/refaktor/rye/evaldo"
	"github.com/stianeikeland/go-rpio/v4"
)

var Builtins_gpio = map[string]*env.Builtin{

	//
	// ##### GPIO ##### "Raspberry Pi GPIO control functions"
	//
	// Example:
	//  ; Blink an LED on pin 17
	//  gpio-open
	//  pin: gpio-pin 17
	//  pin .Output
	//  loop 5 {
	//    pin .Toggle
	//    sleep 500
	//  }
	//  gpio-close
	//
	// Args:
	// * none
	// Returns:
	// * true if successful
	"gpio-open": {
		Argsn: 0,
		Doc:   "Opens and initializes GPIO memory for RPi GPIO access.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			err := rpio.Open()
			if err != nil {
				ps.FailureFlag = true
				return evaldo.MakeBuiltinError(ps, err.Error(), "gpio-open")
			}
			return *env.NewBoolean(true)
		},
	},

	// Args:
	// * none
	// Returns:
	// * true if successful
	"gpio-close": {
		Argsn: 0,
		Doc:   "Closes GPIO memory and frees resources.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			err := rpio.Close()
			if err != nil {
				ps.FailureFlag = true
				return evaldo.MakeBuiltinError(ps, err.Error(), "gpio-close")
			}
			return *env.NewBoolean(true)
		},
	},

	// Tests:
	// equal { pin: gpio-pin 17 |type? } 'native
	// equal { pin: gpio-pin 17 |kind? } 'gpio-pin
	// Args:
	// * pin-number: integer representing the GPIO pin number (BCM numbering)
	// Returns:
	// * native gpio-pin object
	"gpio-pin": {
		Argsn: 1,
		Doc:   "Creates a new GPIO pin object by BCM pin number.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch n := arg0.(type) {
			case env.Integer:
				pin := rpio.Pin(int(n.Value))
				return *env.NewNative(ps.Idx, pin, "gpio-pin")
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.IntegerType}, "gpio-pin")
			}
		},
	},

	// Args:
	// * pin: native gpio-pin object
	// Returns:
	// * the pin object (for chaining)
	"gpio-pin//Input": {
		Argsn: 1,
		Doc:   "Sets the pin to input mode.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch p := arg0.(type) {
			case env.Native:
				pin, ok := p.Value.(rpio.Pin)
				if !ok {
					ps.FailureFlag = true
					return evaldo.MakeBuiltinError(ps, "Native not rpio.Pin", "gpio-pin//Input")
				}
				pin.Input()
				return arg0
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "gpio-pin//Input")
			}
		},
	},

	// Args:
	// * pin: native gpio-pin object
	// Returns:
	// * the pin object (for chaining)
	"gpio-pin//Output": {
		Argsn: 1,
		Doc:   "Sets the pin to output mode.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch p := arg0.(type) {
			case env.Native:
				pin, ok := p.Value.(rpio.Pin)
				if !ok {
					ps.FailureFlag = true
					return evaldo.MakeBuiltinError(ps, "Native not rpio.Pin", "gpio-pin//Output")
				}
				pin.Output()
				return arg0
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "gpio-pin//Output")
			}
		},
	},

	// Args:
	// * pin: native gpio-pin object
	// Returns:
	// * the pin object (for chaining)
	"gpio-pin//High": {
		Argsn: 1,
		Doc:   "Sets the pin output to high (3.3V).",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch p := arg0.(type) {
			case env.Native:
				pin, ok := p.Value.(rpio.Pin)
				if !ok {
					ps.FailureFlag = true
					return evaldo.MakeBuiltinError(ps, "Native not rpio.Pin", "gpio-pin//High")
				}
				pin.High()
				return arg0
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "gpio-pin//High")
			}
		},
	},

	// Args:
	// * pin: native gpio-pin object
	// Returns:
	// * the pin object (for chaining)
	"gpio-pin//Low": {
		Argsn: 1,
		Doc:   "Sets the pin output to low (0V).",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch p := arg0.(type) {
			case env.Native:
				pin, ok := p.Value.(rpio.Pin)
				if !ok {
					ps.FailureFlag = true
					return evaldo.MakeBuiltinError(ps, "Native not rpio.Pin", "gpio-pin//Low")
				}
				pin.Low()
				return arg0
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "gpio-pin//Low")
			}
		},
	},

	// Args:
	// * pin: native gpio-pin object
	// Returns:
	// * the pin object (for chaining)
	"gpio-pin//Toggle": {
		Argsn: 1,
		Doc:   "Toggles the pin output state (high -> low, low -> high).",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch p := arg0.(type) {
			case env.Native:
				pin, ok := p.Value.(rpio.Pin)
				if !ok {
					ps.FailureFlag = true
					return evaldo.MakeBuiltinError(ps, "Native not rpio.Pin", "gpio-pin//Toggle")
				}
				pin.Toggle()
				return arg0
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "gpio-pin//Toggle")
			}
		},
	},

	// Tests:
	// equal { pin: gpio-pin 17 .Input |Read } 0
	// Args:
	// * pin: native gpio-pin object
	// Returns:
	// * integer 0 (low) or 1 (high)
	"gpio-pin//Read": {
		Argsn: 1,
		Doc:   "Reads the pin state. Returns 0 for low, 1 for high.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch p := arg0.(type) {
			case env.Native:
				pin, ok := p.Value.(rpio.Pin)
				if !ok {
					ps.FailureFlag = true
					return evaldo.MakeBuiltinError(ps, "Native not rpio.Pin", "gpio-pin//Read")
				}
				state := pin.Read()
				return *env.NewInteger(int64(state))
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "gpio-pin//Read")
			}
		},
	},

	// Args:
	// * pin: native gpio-pin object
	// Returns:
	// * the pin object (for chaining)
	"gpio-pin//Pull-up": {
		Argsn: 1,
		Doc:   "Enables the internal pull-up resistor on the pin.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch p := arg0.(type) {
			case env.Native:
				pin, ok := p.Value.(rpio.Pin)
				if !ok {
					ps.FailureFlag = true
					return evaldo.MakeBuiltinError(ps, "Native not rpio.Pin", "gpio-pin//Pull-up")
				}
				pin.PullUp()
				return arg0
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "gpio-pin//Pull-up")
			}
		},
	},

	// Args:
	// * pin: native gpio-pin object
	// Returns:
	// * the pin object (for chaining)
	"gpio-pin//Pull-down": {
		Argsn: 1,
		Doc:   "Enables the internal pull-down resistor on the pin.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch p := arg0.(type) {
			case env.Native:
				pin, ok := p.Value.(rpio.Pin)
				if !ok {
					ps.FailureFlag = true
					return evaldo.MakeBuiltinError(ps, "Native not rpio.Pin", "gpio-pin//Pull-down")
				}
				pin.PullDown()
				return arg0
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "gpio-pin//Pull-down")
			}
		},
	},

	// Args:
	// * pin: native gpio-pin object
	// Returns:
	// * the pin object (for chaining)
	"gpio-pin//Pull-off": {
		Argsn: 1,
		Doc:   "Disables pull-up and pull-down resistors on the pin.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch p := arg0.(type) {
			case env.Native:
				pin, ok := p.Value.(rpio.Pin)
				if !ok {
					ps.FailureFlag = true
					return evaldo.MakeBuiltinError(ps, "Native not rpio.Pin", "gpio-pin//Pull-off")
				}
				pin.PullOff()
				return arg0
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "gpio-pin//Pull-off")
			}
		},
	},

	// Args:
	// * pin: native gpio-pin object
	// * frequency: integer representing the PWM frequency in Hz
	// Returns:
	// * the pin object (for chaining)
	"gpio-pin//Freq": {
		Argsn: 2,
		Doc:   "Sets the PWM frequency for a pin in PWM mode.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch p := arg0.(type) {
			case env.Native:
				pin, ok := p.Value.(rpio.Pin)
				if !ok {
					ps.FailureFlag = true
					return evaldo.MakeBuiltinError(ps, "Native not rpio.Pin", "gpio-pin//Freq")
				}
				switch freq := arg1.(type) {
				case env.Integer:
					pin.Freq(int(freq.Value))
					return arg0
				default:
					return evaldo.MakeArgError(ps, 2, []env.Type{env.IntegerType}, "gpio-pin//Freq")
				}
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "gpio-pin//Freq")
			}
		},
	},

	// Args:
	// * pin: native gpio-pin object
	// * duty: integer representing the duty cycle length (in range 0-cycleLen)
	// * cycleLen: integer representing the total cycle length
	// Returns:
	// * the pin object (for chaining)
	"gpio-pin//Duty-cycle": {
		Argsn: 3,
		Doc:   "Sets the PWM duty cycle. Usage: pin .Duty-cycle duty cycleLen.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch p := arg0.(type) {
			case env.Native:
				pin, ok := p.Value.(rpio.Pin)
				if !ok {
					ps.FailureFlag = true
					return evaldo.MakeBuiltinError(ps, "Native not rpio.Pin", "gpio-pin//Duty-cycle")
				}
				switch duty := arg1.(type) {
				case env.Integer:
					switch cycleLen := arg2.(type) {
					case env.Integer:
						pin.DutyCycle(uint32(duty.Value), uint32(cycleLen.Value))
						return arg0
					default:
						return evaldo.MakeArgError(ps, 3, []env.Type{env.IntegerType}, "gpio-pin//Duty-cycle")
					}
				default:
					return evaldo.MakeArgError(ps, 2, []env.Type{env.IntegerType}, "gpio-pin//Duty-cycle")
				}
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "gpio-pin//Duty-cycle")
			}
		},
	},

	// Args:
	// * pin: native gpio-pin object
	// * mode: word representing the pin mode ('Pwm, 'Clock, etc.)
	// Returns:
	// * the pin object (for chaining)
	"gpio-pin//Mode": {
		Argsn: 2,
		Doc:   "Sets the pin mode (e.g. 'Pwm for PWM, 'Clock for clock output).",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch p := arg0.(type) {
			case env.Native:
				pin, ok := p.Value.(rpio.Pin)
				if !ok {
					ps.FailureFlag = true
					return evaldo.MakeBuiltinError(ps, "Native not rpio.Pin", "gpio-pin//Mode")
				}
				switch mode := arg1.(type) {
				case env.Word:
					modeName := ps.Idx.GetWord(mode.Index)
					switch modeName {
					case "Pwm":
						pin.Mode(rpio.Pwm)
					case "Clock":
						pin.Mode(rpio.Clock)
					case "Input":
						pin.Mode(rpio.Input)
					case "Output":
						pin.Mode(rpio.Output)
					default:
						ps.FailureFlag = true
						return evaldo.MakeBuiltinError(ps, fmt.Sprintf("Unknown pin mode: %s. Valid modes: Pwm, Clock, Input, Output", modeName), "gpio-pin//Mode")
					}
					return arg0
				default:
					return evaldo.MakeArgError(ps, 2, []env.Type{env.WordType}, "gpio-pin//Mode")
				}
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "gpio-pin//Mode")
			}
		},
	},

	// Args:
	// * pin: native gpio-pin object
	// * edge: word representing the edge detection mode ('No, 'Rising, 'Falling, 'Both)
	// Returns:
	// * the pin object (for chaining)
	"gpio-pin//Detect": {
		Argsn: 2,
		Doc:   "Sets edge detection mode on the pin. Valid modes: 'No, 'Rising, 'Falling, 'Both.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch p := arg0.(type) {
			case env.Native:
				pin, ok := p.Value.(rpio.Pin)
				if !ok {
					ps.FailureFlag = true
					return evaldo.MakeBuiltinError(ps, "Native not rpio.Pin", "gpio-pin//Detect")
				}
				switch edge := arg1.(type) {
				case env.Word:
					edgeName := ps.Idx.GetWord(edge.Index)
					switch edgeName {
					case "No":
						pin.Detect(rpio.NoEdge)
					case "Rising":
						pin.Detect(rpio.RiseEdge)
					case "Falling":
						pin.Detect(rpio.FallEdge)
					case "Both":
						pin.Detect(rpio.AnyEdge)
					default:
						ps.FailureFlag = true
						return evaldo.MakeBuiltinError(ps, fmt.Sprintf("Unknown edge mode: %s. Valid modes: No, Rising, Falling, Both", edgeName), "gpio-pin//Detect")
					}
					return arg0
				default:
					return evaldo.MakeArgError(ps, 2, []env.Type{env.WordType}, "gpio-pin//Detect")
				}
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "gpio-pin//Detect")
			}
		},
	},

	// Args:
	// * pin: native gpio-pin object
	// Returns:
	// * true if an edge was detected
	"gpio-pin//Edge-detected?": {
		Argsn: 1,
		Doc:   "Returns true if an edge was detected on the pin since last call.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch p := arg0.(type) {
			case env.Native:
				pin, ok := p.Value.(rpio.Pin)
				if !ok {
					ps.FailureFlag = true
					return evaldo.MakeBuiltinError(ps, "Native not rpio.Pin", "gpio-pin//Edge-detected?")
				}
				detected := pin.EdgeDetected()
				return *env.NewBoolean(detected)
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "gpio-pin//Edge-detected?")
			}
		},
	},

	// Args:
	// * pin: native gpio-pin object
	// * state: integer 0 (low) or 1 (high) to write to the pin
	// Returns:
	// * the pin object (for chaining)
	"gpio-pin//Write": {
		Argsn: 2,
		Doc:   "Writes a state (0 for low, 1 for high) to the pin.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch p := arg0.(type) {
			case env.Native:
				pin, ok := p.Value.(rpio.Pin)
				if !ok {
					ps.FailureFlag = true
					return evaldo.MakeBuiltinError(ps, "Native not rpio.Pin", "gpio-pin//Write")
				}
				switch state := arg1.(type) {
				case env.Integer:
					if state.Value == 0 {
						pin.Low()
					} else {
						pin.High()
					}
					return arg0
				default:
					return evaldo.MakeArgError(ps, 2, []env.Type{env.IntegerType}, "gpio-pin//Write")
				}
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "gpio-pin//Write")
			}
		},
	},
}
