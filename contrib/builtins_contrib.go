//go:build b_contrib
// +build b_contrib

// NOTE: when compiling contrib modules, also add b_contrib
// for example: go build -tags "b_sqlite,b_telegram,b_contrib,b_ebitengine,b_surf"

package contrib

import (
	"strings"

	"github.com/refaktor/rye/contrib/aws"
	"github.com/refaktor/rye/contrib/bleve"
	"github.com/refaktor/rye/contrib/postmark"
	"github.com/refaktor/rye/contrib/ryeopenai"
	"github.com/refaktor/rye/env"
	"github.com/refaktor/rye/evaldo"
)

var Builtins_contrib = map[string]*env.Builtin{

	"contrib-one": {
		Argsn: 1,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return env.Integer{1}
		},
	},

	"contrib-do": {
		Argsn: 1,
		Doc:   "Takes a block of code and does (runs) it.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch bloc := arg0.(type) {
			case env.Block:
				ser := ps.Ser
				ps.Ser = bloc.Series
				evaldo.EvalBlock(ps)
				ps.Ser = ser
				return ps.Res
			}
			return nil
		},
	},
}

// Variable to hold conditional registration functions
var surfRegistrationFuncs []func(*env.ProgramState, *map[string]int)

func RegisterBuiltins(ps *env.ProgramState, builtinNames *map[string]int) {
	RegisterBuiltins2(Builtins_contrib, ps, "contrib", builtinNames)
	// RegisterBuiltins2(arkecs.Builtins_arkecs, ps, "arkecs", builtinNames)
	RegisterBuiltins2(aws.Builtins_aws, ps, "aws", builtinNames)
	RegisterBuiltins2(bleve.Builtins_bleve, ps, "bleve", builtinNames)
	RegisterBuiltins2(postmark.Builtins_postmark, ps, "postmark", builtinNames)
	RegisterBuiltins2(ryeopenai.Builtins_openai, ps, "openai", builtinNames)

	// Register surf if build tag is present
	for _, regFunc := range surfRegistrationFuncs {
		regFunc(ps, builtinNames)
	}
}

// var builtinNames map[string]int

func RegisterBuiltins2(builtins map[string]*env.Builtin, ps *env.ProgramState, name string, builtinNames *map[string]int) {
	bn := *builtinNames
	bn[name] = len(builtins)
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
