//go:build b_email
// +build b_email

package evaldo

import (
	"strings"

	"github.com/refaktor/rye/env"

	"github.com/go-gomail/gomail"
)

func __newMessage(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
	return *env.NewNative(ps.Idx, gomail.NewMessage(), "gomail-message")
}

func __setHeader(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
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
			return MakeArgError(ps, 3, []env.Type{env.StringType, env.EmailType}, "gomail-message//set-header")
		}
		switch field := arg1.(type) {
		case env.String:
			fld = field.Value
		case env.Tagword:
			fld = ps.Idx.GetWord(field.Index)
		default:
			return MakeArgError(ps, 2, []env.Type{env.StringType, env.TagwordType}, "gomail-message//set-header")
		}
		if fld != "" && val != "" {
			mailobj.Value.(*gomail.Message).SetHeader(fld, val)
			return arg0
		} else {
			return MakeBuiltinError(ps, "Not both values were defined.", "gomail-message//set-header")
		}
	default:
		return MakeArgError(ps, 1, []env.Type{env.NativeType}, "gomail-message//set-header")
	}
}

func __setAddressHeader(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
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
					ps.FailureFlag = true
					return MakeArgError(ps, 4, []env.Type{env.StringType}, "gomail-message//set-address-header")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 3, []env.Type{env.StringType}, "gomail-message//set-address-header")
			}
		default:
			ps.FailureFlag = true
			return MakeArgError(ps, 2, []env.Type{env.StringType}, "gomail-message//set-address-header")
		}
	default:
		ps.FailureFlag = true
		return MakeArgError(ps, 1, []env.Type{env.NativeType}, "gomail-message//set-address-header")
	}
}

func __setBody(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
	switch mailobj := arg0.(type) {
	case env.Native:
		switch encoding := arg1.(type) {
		case env.String:
			switch value := arg2.(type) {
			case env.String:
				mailobj.Value.(*gomail.Message).SetBody(encoding.Value, value.Value)
				return arg0
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 3, []env.Type{env.StringType}, "gomail-message//set-body")
			}
		default:
			ps.FailureFlag = true
			return MakeArgError(ps, 2, []env.Type{env.StringType}, "gomail-message//set-body")
		}
	default:
		ps.FailureFlag = true
		return MakeArgError(ps, 1, []env.Type{env.NativeType}, "gomail-message//set-body")
	}
}

func __attach(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
	switch mailobj := arg0.(type) {
	case env.Native:
		switch file := arg1.(type) {
		case env.Uri:
			ath := strings.Split(file.Path, "://")
			mailobj.Value.(*gomail.Message).Attach(ath[1])
			return arg0
		default:
			ps.FailureFlag = true
			return MakeArgError(ps, 2, []env.Type{env.UriType}, "gomail-message//attach")
		}
	default:
		ps.FailureFlag = true
		return MakeArgError(ps, 1, []env.Type{env.NativeType}, "gomail-message//attach")
	}

}

func __addAlternative(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
	switch mailobj := arg0.(type) {
	case env.Native:
		switch encoding := arg1.(type) {
		case env.String:
			switch value := arg2.(type) {
			case env.String:
				mailobj.Value.(*gomail.Message).AddAlternative(encoding.Value, value.Value)
				return arg0
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 3, []env.Type{env.StringType}, "gomail-message//add-alternative")
			}
		default:
			ps.FailureFlag = true
			return MakeArgError(ps, 2, []env.Type{env.StringType}, "gomail-message//add-alternative")
		}
	default:
		ps.FailureFlag = true
		return MakeArgError(ps, 1, []env.Type{env.NativeType}, "gomail-message//add-alternative")
	}
}

func __newDialer(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
	switch url := arg0.(type) {
	case env.String:
		switch port := arg1.(type) {
		case env.Integer:
			switch username := arg2.(type) {
			case env.String:
				switch pwd := arg3.(type) {
				case env.String:
					return *env.NewNative(ps.Idx, gomail.NewDialer(url.Value, int(port.Value), username.Value, pwd.Value), "gomail-dialer")
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 4, []env.Type{env.StringType}, "new-email-dialer")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 3, []env.Type{env.StringType}, "new-email-dialer")
			}
		default:
			ps.FailureFlag = true
			return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "new-email-dialer")
		}
	default:
		ps.FailureFlag = true
		return MakeArgError(ps, 1, []env.Type{env.StringType}, "new-email-dialer")
	}
}

func __dialAndSend(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
	switch dialer := arg0.(type) {
	case env.Native:
		switch message := arg1.(type) {
		case env.Native:
			if err := dialer.Value.(*gomail.Dialer).DialAndSend(message.Value.(*gomail.Message)); err != nil {
				ps.FailureFlag = true
				return env.NewError(err.Error())
			}
			return arg0
		default:
			ps.FailureFlag = true
			return MakeArgError(ps, 2, []env.Type{env.NativeType}, "gomail-dialer//dial-and-send")
		}
	default:
		ps.FailureFlag = true
		return MakeArgError(ps, 1, []env.Type{env.NativeType}, "gomail-dialer//dial-and-send")
	}
}

var Builtins_email = map[string]*env.Builtin{

	"new-email-message": {
		Argsn: 0,
		Doc:   "Create new email message.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return __newMessage(ps, arg0, arg1, arg2, arg3, arg4)
		},
	},

	"gomail-message//set-header": {
		Argsn: 3,
		Doc:   "Set email header.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return __setHeader(ps, arg0, arg1, arg2, arg3, arg4)
		},
	},

	"gomail-message//set-address-header": {
		Argsn: 4,
		Doc:   "TODODOC.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return __setAddressHeader(ps, arg0, arg1, arg2, arg3, arg4)
		},
	},

	"gomail-message//set-body": {
		Argsn: 3,
		Doc:   "TODODOC.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return __setBody(ps, arg0, arg1, arg2, arg3, arg4)
		},
	},

	"gomail-message//attach": {
		Argsn: 2,
		Doc:   "TODODOC.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return __attach(ps, arg0, arg1, arg2, arg3, arg4)
		},
	},

	"gomail-message//add-alternative": {
		Argsn: 3,
		Doc:   "TODODOC.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return __addAlternative(ps, arg0, arg1, arg2, arg3, arg4)
		},
	},
	"new-email-dialer": {
		Argsn: 4,
		Doc:   "TODODOC.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return __newDialer(ps, arg0, arg1, arg2, arg3, arg4)
		},
	},
	"gomail-dialer//dial-and-send": {
		Argsn: 2,
		Doc:   "TODODOC.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return __dialAndSend(ps, arg0, arg1, arg2, arg3, arg4)
		},
	},
}
