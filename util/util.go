// util.go
package util

import (
	"fmt"
	"regexp"
	"rye/env"
	"strconv"
	"strings"
)

func PrintHeader() {
	fmt.Println("=-===============-===-===-=============-=")   // Output: -3
	fmt.Println(" _/|\\\\_-~*>%,_  Rejy ZERO  _,%<*~-_//|\\_") // Output: -3
	fmt.Println("=-===============-===-===-=============-=")   // Output: -3
}

func IndexOfAt(s, sep string, n int) int {
	idx := strings.Index(s[n:], sep)
	if idx > -1 {
		idx += n
	}
	return idx
}

func IsTruthy(o env.Object) bool {
	switch oo := o.(type) {
	case env.Integer:
		return oo.Value > 0
	case env.String:
		return len(oo.Value) > 0
	default:
		return false
	}
}

func Dict2Context(ps *env.ProgramState, s1 env.Dict) env.RyeCtx {
	ctx := env.NewEnv(ps.Ctx)
	for k, v := range s1.Data {
		word := ps.Idx.IndexWord(k)
		switch v1 := v.(type) {
		case env.Integer:
			ctx.Set(word, v1)
		case env.String:
			ctx.Set(word, v1)
		}
	}
	return *ctx
}

func StringToFieldsWithQuoted(str string, sepa string, quote string) env.Block {
	quoted := false
	spl := strings.FieldsFunc(str, func(r rune) bool {
		if string(r) == quote {
			quoted = !quoted
		}
		return !quoted && string(r) == sepa
	})
	lst := make([]env.Object, len(spl))
	for i := 0; i < len(spl); i++ {
		//fmt.Println(spl[i])
		// TODO -- detect numbers and turn them to integers or floats, later we can also detect other types of values
		// val, _ := loader.LoadString(spl[i], false)
		numeric, _ := regexp.MatchString("[0-9]+", spl[i])
		// fmt.Println(numeric)
		pass := false
		if numeric {
			num, err := strconv.Atoi(spl[i])
			if err == nil {
				lst[i] = env.Integer{int64(num)}
				pass = true
			} else {
				// fmt.Println(err.Error())
			}
		}
		if !pass {
			clean := regexp.MustCompile(`^"(.*)"$`).ReplaceAllString(spl[i], `$1`)
			val := env.String{clean}
			lst[i] = val
		}
	}
	return *env.NewBlock(*env.NewTSeries(lst))
}

func FormatJson(val env.Object, e env.Idxs) string {
	// TODO -- this is currently made just for block of strings and integers
	var r strings.Builder
	switch b := val.(type) {
	case env.Block:
		r.WriteString("[ ")
		for i := 0; i < b.Series.Len(); i += 1 {
			if b.Series.Get(i) != nil {
				if i > 0 {
					r.WriteString(", ")
				}
				r.WriteString(b.Series.Get(i).Probe(e))
			}
		}
		r.WriteString(" ]")

	}
	return r.String()
}

func FormatCsv(val env.Object, e env.Idxs) string {
	// TODO -- this is currently made just for block of strings and integers
	var r strings.Builder
	switch b := val.(type) {
	case env.Block:
		for i := 0; i < b.Series.Len(); i += 1 {
			if b.Series.Get(i) != nil {
				if i > 0 {
					r.WriteString(",")
				}
				r.WriteString(b.Series.Get(i).Probe(e))
			}
		}
	}
	return r.String()
}

func FormatSsv(val env.Object, e env.Idxs) string {
	// TODO -- this is currently made just for block of strings and integers
	var r strings.Builder
	switch b := val.(type) {
	case env.Block:
		for i := 0; i < b.Series.Len(); i += 1 {
			if b.Series.Get(i) != nil {
				if i > 0 {
					r.WriteString(" ")
				}
				r.WriteString(b.Series.Get(i).Probe(e))
			}
		}
	}
	return r.String()
}
