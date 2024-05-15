// util.go
package util

import (
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/refaktor/rye/env"
)

func TermBold(s string) string {
	return "\033[1m" + s + "\033[22m"
}

func TermError(s string) string {
	return "\033[31m" + s + "\033[0m"
}

func IndexOfAt(s, sep string, n int) int {
	idx := strings.Index(s[n:], sep)
	if idx > -1 {
		idx += n
	}
	return idx
}

func IndexOfSlice(ps *env.ProgramState, slice []env.Object, value env.Object) int {
	for i, v := range slice {
		if v.Equal(value) {
			return i
		}
	}
	return -1 // not found
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
		case string:
			ctx.Set(word, *env.NewString(v1))
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

	re := regexp.MustCompile("[0-9]+")
	for i := 0; i < len(spl); i++ {
		//fmt.Println(spl[i])
		// TODO -- detect numbers and turn them to integers or floats, later we can also detect other types of values
		// val, _ := loader.LoadString(spl[i], false)
		// numeric, _ := regexp.MatchString("[0-9]+", spl[i])
		numeric := re.MatchString(spl[i])
		pass := false
		if numeric {
			num, err := strconv.Atoi(spl[i])
			if err == nil {
				lst[i] = *env.NewInteger(int64(num))
				pass = true
			}
		}
		if !pass {
			clean := regexp.MustCompile(`^"(.*)"$`).ReplaceAllString(spl[i], `$1`)
			val := *env.NewString(clean)
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
				r.WriteString(b.Series.Get(i).Print(e))
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
				r.WriteString(b.Series.Get(i).Print(e))
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
				o := b.Series.Get(i)
				switch ob := o.(type) {
				case env.String:
					r.WriteString(ob.Value)
				default:
					r.WriteString(ob.Print(e))
				}
			}
		}
	}
	return r.String()
}

func SplitEveryString(s string, chunkSize int) []string {
	if len(s) == 0 {
		return nil
	}
	if chunkSize >= len(s) {
		return []string{s}
	}
	var chunks []string = make([]string, 0, (len(s)-1)/chunkSize+1)
	currentLen := 0
	currentStart := 0
	for i := range s {
		if currentLen == chunkSize {
			chunks = append(chunks, s[currentStart:i])
			currentLen = 0
			currentStart = i
		}
		currentLen++
	}
	chunks = append(chunks, s[currentStart:])
	return chunks
}

func SplitEveryList(s []env.Object, chunkSize int) [][]env.Object {
	if len(s) == 0 {
		return nil
	}
	if chunkSize >= len(s) {
		return append(make([][]env.Object, 0), s)
	}
	var chunks [][]env.Object = make([][]env.Object, 0, (len(s)-1)/chunkSize+1)
	var chunk []env.Object = make([]env.Object, 0, chunkSize)
	currentLen := 0
	// currentStart := 0
	for i := range s {
		chunk = append(chunk, s[i])
		currentLen++
		if currentLen == chunkSize {
			chunks = append(chunks, chunk)
			currentLen = 0
			//		currentStart = i + 1
			chunk = make([]env.Object, 0, chunkSize)
		}
	}
	if len(chunk) > 0 {
		chunks = append(chunks, chunk)
	}
	return chunks
}

func IntersectStrings(a string, b string) string {
	set := make(map[rune]bool)
	var bu strings.Builder
	for _, ch := range a {
		if strings.ContainsRune(b, ch) && !set[ch] {
			bu.WriteRune(ch)
		}
	}
	return bu.String()
}

func IntersectBlocks(ps *env.ProgramState, a env.Block, b env.Block) []env.Object {
	set := make(map[env.Object]bool)
	res := make([]env.Object, 0)
	for _, v := range a.Series.S {
		if ContainsVal(ps, b.Series.S, v) && !set[v] {
			res = append(res, v)
		}
	}
	return res
}

func IntersectLists(ps *env.ProgramState, a env.List, b env.List) []any {
	set := make(map[any]bool)
	res := make([]any, 0)
	for _, v := range a.Data {
		if slices.Contains(b.Data, v) && !set[v] {
			res = append(res, v)
		}
	}
	return res
}

func UnionOfBlocks(ps *env.ProgramState, a env.Block, b env.Block) []env.Object {
	elementMap := make(map[env.Object]bool)
	for _, element := range a.Series.S {
		elementMap[element] = true
	}
	for _, element := range b.Series.S {
		elementMap[element] = true
	}
	mergedSlice := make([]env.Object, 0)
	for element := range elementMap {
		mergedSlice = append(mergedSlice, element)
	}
	return mergedSlice
}

func UnionOfLists(ps *env.ProgramState, a env.List, b env.List) []any {
	elementMap := make(map[any]bool)
	for _, element := range a.Data {
		elementMap[element] = true
	}
	for _, element := range b.Data {
		elementMap[element] = true
	}
	mergedSlice := make([]any, 0)
	for element := range elementMap {
		mergedSlice = append(mergedSlice, element)
	}
	return mergedSlice
}

func DiffStrings(a string, b string) string {
	set := make(map[rune]bool)
	var bu strings.Builder
	for _, ch := range a {
		if !strings.ContainsRune(b, ch) && !set[ch] {
			set[ch] = true
			bu.WriteRune(ch)
		}
	}
	return bu.String()
}

func DiffBlocks(ps *env.ProgramState, a env.Block, b env.Block) []env.Object {
	set := make(map[env.Object]bool)
	res := make([]env.Object, 0)
	for _, v := range a.Series.S {
		if !ContainsVal(ps, b.Series.S, v) && !set[v] {
			set[v] = true
			res = append(res, v)
		}
	}
	return res
}

func DiffLists(ps *env.ProgramState, a env.List, b env.List) []any {
	set := make(map[any]bool)
	res := make([]any, 0)
	for _, v := range a.Data {
		if !slices.Contains(b.Data, v) && !set[v] {
			set[v] = true
			res = append(res, v)
		}
	}
	return res
}

func SplitMulti(s string, seps string) []string {
	splitter := func(r rune) bool {
		return strings.ContainsRune(seps, r)
	}
	return strings.FieldsFunc(s, splitter)
}

func ContainsVal(ps *env.ProgramState, b []env.Object, val env.Object) bool {
	for _, a := range b {
		if a.Equal(val) {
			return true
		}
	}
	return false
}

func RemoveDuplicate(ps *env.ProgramState, slice []env.Object) []env.Object {
	allKeys := make(map[env.Object]bool)
	list := []env.Object{}
	for _, item := range slice {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
}

func TruncateString(s string, maxLen int) string {
	s = strings.ReplaceAll(s, "\n", "\\-")
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	if maxLen < 3 {
		maxLen = 3
	}
	return string(runes[0:maxLen-3]) + "..."
}

func ProcessFunctionSpec(args env.Block) (bool, string) {
	var doc string
	if args.Series.Len() > 0 {
		var hasDoc bool
		switch a := args.Series.S[len(args.Series.S)-1].(type) {
		case env.String:
			doc = a.Value
			hasDoc = true
			//fmt.Println("DOC DOC")
			// default:
			//return MakeBuiltinError(ps, "Series type should be string.", "fn")
		}
		for i, o := range args.Series.GetAll() {
			if i == len(args.Series.S)-1 && hasDoc {
				break
			}
			if o.Type() != env.WordType {
				return false, "Function arguments should be words"
			}
		}
	}
	return true, doc
}

// GetDimValue get max x-y or 0 value
func GetDimValue(x, y float64) float64 {
	difference := x - y
	if difference > 0 {
		return difference
	}
	return 0
}

/*
func RemoveDuplicate[T comparable](sliceList []T) []T {
	allKeys := make(map[T]bool)
	list := []T{}
	for _, item := range sliceList {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
}
*/

/* func Transpose(slice []env.Object) []env.Object {
	yl := len(slice)
	var xl int
	switch blk := slice[0].(type) {
	case env.Block:
		xl = len(blk.Series.S)
	}
	if xl == 0 {
		return nil
	}
	result := make([]env.Object, xl)
	//for i := range result { // TODOD .... finish this or comment it out next time ... left as it is
	// result[i] = env.NewBlock(env.NewTSeries()) // make([]env.Object, yl)
	//}
	for i := 0; i < xl; i++ {
		for j := 0; j < yl; j++ {
			result[i][j] = slice[j][i]
		}
	}
	return result
        }*/
