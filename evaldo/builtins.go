// builtins.go
package evaldo

import (
	// "io/ioutil"
	//	"os/exec"
	// "reflect"

	//	"bufio"
	"fmt"
	//	"os"
	"rye/env"

	// "rye/loader"
	//	"rye/term"
	"strings"
)


func ss() {
	fmt.Print(1)
}

func makeError(env1 *env.ProgramState, msg string) *env.Error {
	env1.FailureFlag = true
	return env.NewError(msg)
}

func equalValues(ps *env.ProgramState, arg0 env.Object, arg1 env.Object) bool {
	return arg0.GetKind() == arg1.GetKind() && arg0.Inspect(*ps.Idx) == arg1.Inspect(*ps.Idx)
}

func getFrom(ps *env.ProgramState, data interface{}, key interface{}, posMode bool) env.Object {
	switch s1 := data.(type) {
	case env.Dict:
		switch s2 := key.(type) {
		case env.String:
			v := s1.Data[s2.Value]
			switch v1 := v.(type) {
			case int, int64, float64, string, []interface{}, map[string]interface{}:
				return JsonToRye(v1)
			case env.Integer:
				return v1
			case env.String:
				return v1
			case env.Date:
				return v1
			case env.Block:
				return v1
			case env.Dict:
				return v1
			case env.List:
				return v1
			case nil:
				ps.FailureFlag = true
				return env.NewError("missing key")
			default:
				ps.FailureFlag = true
				return env.NewError("Value of type: x") // JM 202202 + reflect.TypeOf(v1).String())
			}
		}
	case env.RyeCtx:
		switch s2 := key.(type) {
		case env.Tagword:
			v, ok := s1.Get(s2.Index)
			if ok {
				return v
			} else {
				ps.FailureFlag = true
				return env.NewError1(5) // NOT_FOUND
			}
		}
	case env.List:
		switch s2 := key.(type) {
		case env.Integer:
			idx := s2.Value
			if posMode {
				idx--
			}
			v := s1.Data[idx]
			ok := true
			if ok {
				return JsonToRye(v)
			} else {
				ps.FailureFlag = true
				return env.NewError1(5) // NOT_FOUND
			}
		}
	case env.Block:
		switch s2 := key.(type) {
		case env.Integer:
			idx := s2.Value
			if posMode {
				idx--
			}
			v := s1.Series.Get(int(idx))
			ok := true
			if ok {
				return v
			} else {
				ps.FailureFlag = true
				return env.NewError1(5) // NOT_FOUND
			}
		}
	}
	return env.NewError("wrong types TODO")
}

/* Terminal functions .. move to it's own later */

/*
func isTruthy(arg env.Object) env.Object {
	switch cond := arg.(type) {
	case env.Integer:
		return cond.Value != 0
	case env.String:
		return cond.Value != ""
	default:
		// if it's neither we just return error for now
		ps.FailureFlag = true
		return env.NewError("Error determining if truty")
	}
}
*/

func RegisterBuiltins(ps *env.ProgramState) {
	// 	RegisterBuiltins2(builtins, ps)
}

func RegisterBuiltins2(builtins map[string]*env.Builtin, ps *env.ProgramState) {
	for k, v := range builtins {
		bu := env.NewBuiltin(v.Fn, v.Argsn, v.AcceptFailure, v.Pure, v.Doc)
		registerBuiltin(ps, k, *bu)
	}
}

func registerBuiltin(ps *env.ProgramState, word string, builtin env.Builtin) {
	// indexWord
	// TODO -- this with string separator is a temporary way of how we define generic builtins
	// in future a map will probably not be a map but an array and builtin will also support the Kind value

	idxk := 0
	if strings.Index(word, "//") > 0 {
		temp := strings.Split(word, "//")
		word = temp[1]
		idxk = ps.Idx.IndexWord(temp[0])
	}
	idxw := ps.Idx.IndexWord(word)
	// set global word with builtin
	if idxk == 0 {
		ps.Ctx.Set(idxw, builtin)
		if builtin.Pure {
			ps.PCtx.Set(idxw, builtin)
		}

	} else {
		ps.Gen.Set(idxk, idxw, builtin)
	}
}
