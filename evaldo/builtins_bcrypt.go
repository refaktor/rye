//go:build !no_bcrypt
// +build !no_bcrypt

package evaldo

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/refaktor/rye/env"

	"golang.org/x/crypto/bcrypt"
)

func __bcrypt_hash(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
	switch str := arg0.(type) {
	case env.String:
		bytes, err := bcrypt.GenerateFromPassword([]byte(str.Value), bcrypt.DefaultCost)
		if err != nil {
			ps.FailureFlag = true
			return MakeBuiltinError(ps, "Problem in hashing.", "__bcrypt_hash")
		}
		return env.String{string(bytes)}
	default:
		ps.FailureFlag = true
		return MakeArgError(ps, 1, []env.Type{env.StringType}, "__bcrypt_hash")
	}
}

func __bcrypt_check(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
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
			ps.FailureFlag = true
			return MakeArgError(ps, 1, []env.Type{env.StringType}, "__bcrypt_check")
		}
	default:
		ps.FailureFlag = true
		return MakeArgError(ps, 2, []env.Type{env.StringType}, "__bcrypt_check")
	}
}

func __generate_token(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
	switch n := arg0.(type) {
	case env.Integer:
		b := make([]byte, n.Value)
		if _, err := rand.Read(b); err != nil {
			return MakeBuiltinError(ps, "Problem reading random stream.", "__generate_token")
		}
		return env.String{hex.EncodeToString(b)}
	default:
		ps.FailureFlag = true
		return MakeArgError(ps, 1, []env.Type{env.IntegerType}, "__generate_token")
	}
}

var Builtins_bcrypt = map[string]*env.Builtin{

	"bcrypt-hash": {
		Argsn: 1,
		Doc:   "Generate hashing.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return __bcrypt_hash(ps, arg0, arg1, arg2, arg3, arg4)
		},
	},

	"bcrypt-check": {
		Argsn: 2,
		Doc:   "Compare hash and password.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return __bcrypt_check(ps, arg0, arg1, arg2, arg3, arg4)
		},
	},
	"generate-token": {
		Argsn: 1,
		Doc:   "Generate token for hashing.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return __generate_token(ps, arg0, arg1, arg2, arg3, arg4)
		},
	},
}
