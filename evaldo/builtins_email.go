// +build !b_tiny

package evaldo

import (
	"rye/env"

	"github.com/go-gomail/gomail"
)

func __newMessage(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
	return *env.NewNative(env1.Idx, gomail.NewMessage(), "gomail-message")
}

func __setHeader(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
	switch mailobj := arg0.(type) {
	case env.Native:
		switch field := arg1.(type) {
		case env.String:
			switch value := arg2.(type) {
			case env.String:
				mailobj.Value.(*gomail.Message).SetHeader(field.Value, value.Value)
				return arg0
			default:
				env1.FailureFlag = true
				return env.NewError("arg 3 should be string")
			}
		default:
			env1.FailureFlag = true
			return env.NewError("arg 2 should be string")
		}
	default:
		env1.FailureFlag = true
		return env.NewError("arg 1 should be native")
	}
}

func __setAddressHeader(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
	switch mailobj := arg0.(type) {
	case env.Native:
		switch field := arg1.(type) {
		case env.String:
			switch value := arg2.(type) {
			case env.String:
				switch name := arg3.(type) {
				case env.String:
					mailobj.Value.(*gomail.Message).SetAddressHeader(field.Value, value.Value, name.Value)
					return arg0
				default:
					env1.FailureFlag = true
					return env.NewError("arg 1 should be string")
				}
			default:
				env1.FailureFlag = true
				return env.NewError("arg 1 should be string")
			}
		default:
			env1.FailureFlag = true
			return env.NewError("arg 2 should be string")
		}
	default:
		env1.FailureFlag = true
		return env.NewError("arg 2 should be string")
	}
}

func __setBody(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
	switch mailobj := arg0.(type) {
	case env.Native:
		switch encoding := arg1.(type) {
		case env.String:
			switch value := arg2.(type) {
			case env.String:
				mailobj.Value.(*gomail.Message).SetBody(encoding.Value, value.Value)
				return arg0
			default:
				env1.FailureFlag = true
				return env.NewError("arg 1 should be string")
			}
		default:
			env1.FailureFlag = true
			return env.NewError("arg 2 should be string")
		}
	default:
		env1.FailureFlag = true
		return env.NewError("arg 2 should be string")
	}
}

func __attach(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
	switch mailobj := arg0.(type) {
	case env.Native:
		switch filepath := arg1.(type) {
		case env.String:
			mailobj.Value.(*gomail.Message).Attach(filepath.Value)
			return arg0
		default:
			env1.FailureFlag = true
			return env.NewError("arg 2 should be string")
		}
	default:
		env1.FailureFlag = true
		return env.NewError("arg 2 should be string")
	}

}

func __addAlternative(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
	switch mailobj := arg0.(type) {
	case env.Native:
		switch encoding := arg1.(type) {
		case env.String:
			switch value := arg2.(type) {
			case env.String:
				mailobj.Value.(*gomail.Message).AddAlternative(encoding.Value, value.Value)
				return arg0
			default:
				env1.FailureFlag = true
				return env.NewError("arg 1 should be string")
			}
		default:
			env1.FailureFlag = true
			return env.NewError("arg 2 should be string")
		}
	default:
		env1.FailureFlag = true
		return env.NewError("arg 2 should be string")
	}
}

func __newDialer(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
	switch url := arg0.(type) {
	case env.String:
		switch port := arg1.(type) {
		case env.Integer:
			switch username := arg2.(type) {
			case env.String:
				switch pwd := arg3.(type) {
				case env.String:
					return *env.NewNative(env1.Idx, gomail.NewDialer(url.Value, int(port.Value), username.Value, pwd.Value), "gomail-dialer")
				default:
					env1.FailureFlag = true
					return env.NewError("arg 4 should be string")
				}
			default:
				env1.FailureFlag = true
				return env.NewError("arg 3 should be string")
			}
		default:
			env1.FailureFlag = true
			return env.NewError("arg 2 should be int")
		}
	default:
		env1.FailureFlag = true
		return env.NewError("arg 1 should be string")
	}
}

func __dialAndSend(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
	switch dialer := arg0.(type) {
	case env.Native:
		switch message := arg1.(type) {
		case env.Native:
			if err := dialer.Value.(*gomail.Dialer).DialAndSend(message.Value.(*gomail.Message)); err != nil {
				env1.FailureFlag = true
				return env.NewError(err.Error())
			}
			return arg0
		default:
			env1.FailureFlag = true
			return env.NewError("arg 2 should be native")
		}
	default:
		env1.FailureFlag = true
		return env.NewError("arg 1 should be native")
	}
}

var Builtins_email = map[string]*env.Builtin{

	"new-gomail-message": {
		Argsn: 0,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return __newMessage(env1, arg0, arg1, arg2, arg3, arg4)
		},
	},

	"gomail-message//set-header": {
		Argsn: 3,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return __setHeader(env1, arg0, arg1, arg2, arg3, arg4)
		},
	},

	"gomail-message//set-address-header": {
		Argsn: 4,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return __setAddressHeader(env1, arg0, arg1, arg2, arg3, arg4)
		},
	},

	"gomail-message//set-body": {
		Argsn: 3,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return __setBody(env1, arg0, arg1, arg2, arg3, arg4)
		},
	},

	"gomail-message//attach": {
		Argsn: 2,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return __attach(env1, arg0, arg1, arg2, arg3, arg4)
		},
	},

	"gomail-message//add-alternative": {
		Argsn: 3,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return __addAlternative(env1, arg0, arg1, arg2, arg3, arg4)
		},
	},
	"new-gomail-dialer": {
		Argsn: 4,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return __newDialer(env1, arg0, arg1, arg2, arg3, arg4)
		},
	},
	"gomail-dialer//dial-and-send": {
		Argsn: 2,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return __dialAndSend(env1, arg0, arg1, arg2, arg3, arg4)
		},
	},
}
