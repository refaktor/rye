// +build !b_tiny

package evaldo

import (
	"crypto/rand"
	"encoding/hex"
	"rye/env"

	"golang.org/x/crypto/bcrypt"
)

func __bcrypt_hash(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
	switch str := arg0.(type) {
	case env.String:
		bytes, err := bcrypt.GenerateFromPassword([]byte(str.Value), bcrypt.DefaultCost)
		if err != nil {
			env1.FailureFlag = true
			return env.NewError("problem hashing")
		}
		return env.String{string(bytes)}
	default:
		env1.FailureFlag = true
		return env.NewError("arg 1 should be string")
	}
}

func __bcrypt_check(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
	switch password := arg1.(type) {
	case env.String:
		switch hash := arg0.(type) {
		case env.String:
			err := bcrypt.CompareHashAndPassword([]byte(hash.Value), []byte(password.Value))
			if err == nil {
				return env.Integer{1}
			} else {
				return env.Integer{0}
			}
		default:
			env1.FailureFlag = true
			return env.NewError("arg 1 should be string")
		}
	default:
		env1.FailureFlag = true
		return env.NewError("arg 1 should be string")
	}
}

func __generate_token(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
	switch n := arg0.(type) {
	case env.Integer:
		b := make([]byte, n.Value)
		if _, err := rand.Read(b); err != nil {
			return env.NewError("problem reading random stream")
		}
		return env.String{hex.EncodeToString(b)}
	default:
		env1.FailureFlag = true
		return env.NewError("arg 1 should be string")
	}
}

var Builtins_bcrypt = map[string]*env.Builtin{

	"bcrypt-hash": {
		Argsn: 1,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return __bcrypt_hash(env1, arg0, arg1, arg2, arg3, arg4)
		},
	},

	"bcrypt-check": {
		Argsn: 2,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return __bcrypt_check(env1, arg0, arg1, arg2, arg3, arg4)
		},
	},
	"generate-token": {
		Argsn: 1,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return __generate_token(env1, arg0, arg1, arg2, arg3, arg4)
		},
	},
}
