// +build b_email

package evaldo

import (
	"rye/env"
	"strings"

	"github.com/go-gomail/gomail"
)

func __newMessage(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
	return *env.NewNative(env1.Idx, gomail.NewMessage(), "gomail-message")
}

func __setHeader(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
	switch mailobj := arg0.(type) {
	case env.Native:
		var fld string
		var val string
		switch value := arg2.(type) {
		case env.String:
			val = value.Value
		case env.Email:
			val = value.Address
		default:
			return MakeError(env1, "A3 should be string or email")
		}
		switch field := arg1.(type) {
		case env.String:
			fld = field.Value
		case env.Tagword:
			fld = env1.Idx.GetWord(field.Index)
		default:
			return MakeError(env1, "A2 should be string or tagword")
		}
		if fld != "" && val != "" {
			mailobj.Value.(*gomail.Message).SetHeader(fld, val)
			return arg0
		} else {
			return MakeError(env1, "Not both values were defined")
		}
	default:
		return MakeError(env1, "A1 should be native")
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
		switch file := arg1.(type) {
		case env.Uri:
			ath := strings.Split(file.Path, "://")
			mailobj.Value.(*gomail.Message).Attach(ath[1])
			return arg0
		default:
			env1.FailureFlag = true
			return env.NewError("arg 2 should be Uri")
		}
	default:
		env1.FailureFlag = true
		return env.NewError("arg 1 should be native")
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

	"new-email-message": {
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
	"new-email-dialer": {
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
