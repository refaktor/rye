//go:build add_gpio_sim
// +build add_gpio_sim

package batteries

import (
	"fmt"
	"sync"

	"github.com/refaktor/rye/env"
	"github.com/refaktor/rye/evaldo"
)

// simPin is an in-memory simulated GPIO pin for testing on non-RPi hardware.
// It stores direction, state, pull, mode, edge detection etc. in memory.
type simPin struct {
	mu           sync.Mutex
	pinNum       int
	isOutput     bool
	state        int // 0 = low, 1 = high
	pull         string
	mode         string
	edge         string
	edgeDetected bool
	freq         int
	dutyLen      uint32
	cycleLen     uint32
}

func (p *simPin) Input() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.isOutput = false
	p.mode = "Input"
}

func (p *simPin) Output() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.isOutput = true
	p.mode = "Output"
}

func (p *simPin) High() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.state = 1
}

func (p *simPin) Low() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.state = 0
}

func (p *simPin) Toggle() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.state ^= 1
}

func (p *simPin) Read() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.state
}

func (p *simPin) Write(state int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.state = state
}

func (p *simPin) PullUp() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.pull = "up"
}

func (p *simPin) PullDown() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.pull = "down"
}

func (p *simPin) PullOff() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.pull = "off"
}

func (p *simPin) SetFreq(freq int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.freq = freq
}

func (p *simPin) SetDutyCycle(dutyLen, cycleLen uint32) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.dutyLen = dutyLen
	p.cycleLen = cycleLen
}

func (p *simPin) SetMode(mode string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.mode = mode
	if mode == "Input" {
		p.isOutput = false
	} else {
		p.isOutput = true
	}
}

func (p *simPin) SetEdgeDetect(edge string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.edge = edge
}

func (p *simPin) EdgeDetected() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	detected := p.edgeDetected
	p.edgeDetected = false // clear after read
	return detected
}

// SimulateEdge simulates an edge event on the pin (for testing).
func (p *simPin) SimulateEdge() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.edgeDetected = true
}

func (p *simPin) PinNumber() int {
	return p.pinNum
}

func (p *simPin) GetState() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.state
}

func (p *simPin) GetMode() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.mode
}

func (p *simPin) GetPull() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.pull
}

func (p *simPin) GetFreq() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.freq
}

func (p *simPin) GetDutyLen() uint32 {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.dutyLen
}

func (p *simPin) GetCycleLen() uint32 {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.cycleLen
}

// Global store for simulated pins so they persist across builtin calls.
var simPinStore = make(map[int]*simPin)

func getOrCreateSimPin(pinNum int) *simPin {
	if p, ok := simPinStore[pinNum]; ok {
		return p
	}
	p := &simPin{pinNum: pinNum}
	simPinStore[pinNum] = p
	return p
}

var Builtins_gpio = map[string]*env.Builtin{

	//
	// ##### GPIO (Simulated) ##### "Simulated GPIO for testing on desktop"
	//
	// Example:
	//  gpio-open
	//  pin: gpio-pin 17
	//  pin .Output .High
	//  print pin .Read   ; prints 1
	//  pin .Toggle
	//  print pin .Read   ; prints 0
	//  gpio-close

	// Args:
	// * none
	// Returns:
	// * true (simulated)
	"gpio-open": {
		Argsn: 0,
		Doc:   "[SIM] Opens simulated GPIO (no hardware access).",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewBoolean(true)
		},
	},

	// Args:
	// * none
	// Returns:
	// * true (simulated)
	"gpio-close": {
		Argsn: 0,
		Doc:   "[SIM] Closes simulated GPIO.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewBoolean(true)
		},
	},

	// Args:
	// * pin-number: integer (BCM numbering)
	// Returns:
	// * native sim-pin object
	"gpio-pin": {
		Argsn: 1,
		Doc:   "[SIM] Creates a simulated GPIO pin by BCM pin number.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch n := arg0.(type) {
			case env.Integer:
				p := getOrCreateSimPin(int(n.Value))
				return *env.NewNative(ps.Idx, p, "gpio-pin")
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.IntegerType}, "gpio-pin")
			}
		},
	},

	// Args:
	// * pin: native sim-pin object
	// Returns:
	// * the pin object (for chaining)
	"gpio-pin//Input": {
		Argsn: 1,
		Doc:   "[SIM] Sets the pin to input mode.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch p := arg0.(type) {
			case env.Native:
				pin, ok := p.Value.(*simPin)
				if !ok {
					ps.FailureFlag = true
					return evaldo.MakeBuiltinError(ps, "Native not simPin", "gpio-pin//Input")
				}
				pin.Input()
				return arg0
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "gpio-pin//Input")
			}
		},
	},

	// Args:
	// * pin: native sim-pin object
	// Returns:
	// * the pin object (for chaining)
	"gpio-pin//Output": {
		Argsn: 1,
		Doc:   "[SIM] Sets the pin to output mode.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch p := arg0.(type) {
			case env.Native:
				pin, ok := p.Value.(*simPin)
				if !ok {
					ps.FailureFlag = true
					return evaldo.MakeBuiltinError(ps, "Native not simPin", "gpio-pin//Output")
				}
				pin.Output()
				return arg0
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "gpio-pin//Output")
			}
		},
	},

	// Args:
	// * pin: native sim-pin object
	// Returns:
	// * the pin object (for chaining)
	"gpio-pin//High": {
		Argsn: 1,
		Doc:   "[SIM] Sets the pin output to high.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch p := arg0.(type) {
			case env.Native:
				pin, ok := p.Value.(*simPin)
				if !ok {
					ps.FailureFlag = true
					return evaldo.MakeBuiltinError(ps, "Native not simPin", "gpio-pin//High")
				}
				pin.High()
				return arg0
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "gpio-pin//High")
			}
		},
	},

	// Args:
	// * pin: native sim-pin object
	// Returns:
	// * the pin object (for chaining)
	"gpio-pin//Low": {
		Argsn: 1,
		Doc:   "[SIM] Sets the pin output to low.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch p := arg0.(type) {
			case env.Native:
				pin, ok := p.Value.(*simPin)
				if !ok {
					ps.FailureFlag = true
					return evaldo.MakeBuiltinError(ps, "Native not simPin", "gpio-pin//Low")
				}
				pin.Low()
				return arg0
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "gpio-pin//Low")
			}
		},
	},

	// Args:
	// * pin: native sim-pin object
	// Returns:
	// * the pin object (for chaining)
	"gpio-pin//Toggle": {
		Argsn: 1,
		Doc:   "[SIM] Toggles the pin output state.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch p := arg0.(type) {
			case env.Native:
				pin, ok := p.Value.(*simPin)
				if !ok {
					ps.FailureFlag = true
					return evaldo.MakeBuiltinError(ps, "Native not simPin", "gpio-pin//Toggle")
				}
				pin.Toggle()
				return arg0
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "gpio-pin//Toggle")
			}
		},
	},

	// Args:
	// * pin: native sim-pin object
	// Returns:
	// * integer 0 (low) or 1 (high)
	"gpio-pin//Read": {
		Argsn: 1,
		Doc:   "[SIM] Reads the pin state. Returns 0 for low, 1 for high.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch p := arg0.(type) {
			case env.Native:
				pin, ok := p.Value.(*simPin)
				if !ok {
					ps.FailureFlag = true
					return evaldo.MakeBuiltinError(ps, "Native not simPin", "gpio-pin//Read")
				}
				return *env.NewInteger(int64(pin.Read()))
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "gpio-pin//Read")
			}
		},
	},

	// Args:
	// * pin: native sim-pin object
	// Returns:
	// * the pin object (for chaining)
	"gpio-pin//Pull-up": {
		Argsn: 1,
		Doc:   "[SIM] Enables pull-up resistor.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch p := arg0.(type) {
			case env.Native:
				pin, ok := p.Value.(*simPin)
				if !ok {
					ps.FailureFlag = true
					return evaldo.MakeBuiltinError(ps, "Native not simPin", "gpio-pin//Pull-up")
				}
				pin.PullUp()
				return arg0
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "gpio-pin//Pull-up")
			}
		},
	},

	// Args:
	// * pin: native sim-pin object
	// Returns:
	// * the pin object (for chaining)
	"gpio-pin//Pull-down": {
		Argsn: 1,
		Doc:   "[SIM] Enables pull-down resistor.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch p := arg0.(type) {
			case env.Native:
				pin, ok := p.Value.(*simPin)
				if !ok {
					ps.FailureFlag = true
					return evaldo.MakeBuiltinError(ps, "Native not simPin", "gpio-pin//Pull-down")
				}
				pin.PullDown()
				return arg0
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "gpio-pin//Pull-down")
			}
		},
	},

	// Args:
	// * pin: native sim-pin object
	// Returns:
	// * the pin object (for chaining)
	"gpio-pin//Pull-off": {
		Argsn: 1,
		Doc:   "[SIM] Disables pull resistors.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch p := arg0.(type) {
			case env.Native:
				pin, ok := p.Value.(*simPin)
				if !ok {
					ps.FailureFlag = true
					return evaldo.MakeBuiltinError(ps, "Native not simPin", "gpio-pin//Pull-off")
				}
				pin.PullOff()
				return arg0
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "gpio-pin//Pull-off")
			}
		},
	},

	// Args:
	// * pin: native sim-pin object
	// * frequency: integer in Hz
	// Returns:
	// * the pin object (for chaining)
	"gpio-pin//Freq": {
		Argsn: 2,
		Doc:   "[SIM] Sets the PWM frequency.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch p := arg0.(type) {
			case env.Native:
				pin, ok := p.Value.(*simPin)
				if !ok {
					ps.FailureFlag = true
					return evaldo.MakeBuiltinError(ps, "Native not simPin", "gpio-pin//Freq")
				}
				switch freq := arg1.(type) {
				case env.Integer:
					pin.SetFreq(int(freq.Value))
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
	// * pin: native sim-pin object
	// * duty: integer duty length
	// * cycleLen: integer cycle length
	// Returns:
	// * the pin object (for chaining)
	"gpio-pin//Duty-cycle": {
		Argsn: 3,
		Doc:   "[SIM] Sets the PWM duty cycle.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch p := arg0.(type) {
			case env.Native:
				pin, ok := p.Value.(*simPin)
				if !ok {
					ps.FailureFlag = true
					return evaldo.MakeBuiltinError(ps, "Native not simPin", "gpio-pin//Duty-cycle")
				}
				switch duty := arg1.(type) {
				case env.Integer:
					switch cycleLen := arg2.(type) {
					case env.Integer:
						pin.SetDutyCycle(uint32(duty.Value), uint32(cycleLen.Value))
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
	// * pin: native sim-pin object
	// * mode: word ('Pwm, 'Clock, 'Input, 'Output)
	// Returns:
	// * the pin object (for chaining)
	"gpio-pin//Mode": {
		Argsn: 2,
		Doc:   "[SIM] Sets the pin mode.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch p := arg0.(type) {
			case env.Native:
				pin, ok := p.Value.(*simPin)
				if !ok {
					ps.FailureFlag = true
					return evaldo.MakeBuiltinError(ps, "Native not simPin", "gpio-pin//Mode")
				}
				switch mode := arg1.(type) {
				case env.Word:
					modeName := ps.Idx.GetWord(mode.Index)
					switch modeName {
					case "Pwm", "Clock", "Input", "Output":
						pin.SetMode(modeName)
					default:
						ps.FailureFlag = true
						return evaldo.MakeBuiltinError(ps, fmt.Sprintf("Unknown pin mode: %s", modeName), "gpio-pin//Mode")
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
	// * pin: native sim-pin object
	// * edge: word ('No, 'Rising, 'Falling, 'Both)
	// Returns:
	// * the pin object (for chaining)
	"gpio-pin//Detect": {
		Argsn: 2,
		Doc:   "[SIM] Sets edge detection mode.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch p := arg0.(type) {
			case env.Native:
				pin, ok := p.Value.(*simPin)
				if !ok {
					ps.FailureFlag = true
					return evaldo.MakeBuiltinError(ps, "Native not simPin", "gpio-pin//Detect")
				}
				switch edge := arg1.(type) {
				case env.Word:
					edgeName := ps.Idx.GetWord(edge.Index)
					switch edgeName {
					case "No", "Rising", "Falling", "Both":
						pin.SetEdgeDetect(edgeName)
					default:
						ps.FailureFlag = true
						return evaldo.MakeBuiltinError(ps, fmt.Sprintf("Unknown edge mode: %s", edgeName), "gpio-pin//Detect")
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
	// * pin: native sim-pin object
	// Returns:
	// * true if an edge was detected
	"gpio-pin//Edge-detected?": {
		Argsn: 1,
		Doc:   "[SIM] Returns true if edge was detected since last call.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch p := arg0.(type) {
			case env.Native:
				pin, ok := p.Value.(*simPin)
				if !ok {
					ps.FailureFlag = true
					return evaldo.MakeBuiltinError(ps, "Native not simPin", "gpio-pin//Edge-detected?")
				}
				return *env.NewBoolean(pin.EdgeDetected())
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "gpio-pin//Edge-detected?")
			}
		},
	},

	// Args:
	// * pin: native sim-pin object
	// * state: integer 0 or 1
	// Returns:
	// * the pin object (for chaining)
	"gpio-pin//Write": {
		Argsn: 2,
		Doc:   "[SIM] Writes a state to the pin.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch p := arg0.(type) {
			case env.Native:
				pin, ok := p.Value.(*simPin)
				if !ok {
					ps.FailureFlag = true
					return evaldo.MakeBuiltinError(ps, "Native not simPin", "gpio-pin//Write")
				}
				switch state := arg1.(type) {
				case env.Integer:
					pin.Write(int(state.Value))
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
