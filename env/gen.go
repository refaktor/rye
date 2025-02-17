package env

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// Gen -- generic functions dictionary

type Gen struct {
	dict map[int]map[int]Object
}

func NewGen() *Gen {
	var e Gen
	e.dict = make(map[int]map[int]Object)
	return &e
}

func (e *Gen) Print(idxs Idxs) {
	fmt.Print("<Gen Dict: ")
	for k, v := range e.dict {
		fmt.Print(" <Kind: " + strconv.FormatInt(int64(k), 10))
		for k1, v1 := range v {
			fmt.Print(" " + strconv.FormatInt(int64(k1), 10) + ": " + v1.Inspect(idxs) + " ")
		}
		fmt.Print(" >")
	}
	fmt.Println(">")
}

func (e *Gen) Get(kind int, word int) (Object, bool) {
	obj, exists := e.dict[kind][word]
	// since here is no parent ... this lookup could be faster maybe ... that parent lookup was taking a lot of time
	// in Env
	return obj, exists
}

func (e *Gen) Set(kind int, word int, val Object) Object {
	if e.dict[kind] == nil {
		e.dict[kind] = make(map[int]Object)
	}
	e.dict[kind][word] = val
	return val
}

func (e *Gen) GetKinds() map[int]int {
	keys := make(map[int]int)
	for k, v := range e.dict {
		keys[k] = len(v)
	}
	return keys
}

func (e *Gen) GetMethods(kind int) []int {
	meths := make([]int, len(e.dict[kind]))
	i := 0
	for k := range e.dict[kind] {
		meths[i] = k
		i++
	}
	return meths
}

func (e Gen) PreviewKinds(idxs Idxs, filter string) string {
	var bu strings.Builder
	bu.WriteString("Kinds:")
	arr := make([]string, 0)
	for k, _ := range e.dict {
		str1 := idxs.GetWord(k)
		if strings.Contains(str1, filter) {
			color := color_word2
			arr = append(arr, reset+color+str1+reset) // idxs.GetWord(v.GetKind()
		}
	}
	sort.Strings(arr)
	for aa := range arr {
		line := arr[aa]
		//pars := strings.Split(line, "|||")
		bu.WriteString("\n\r " + line)
	}
	return bu.String()
}

// const color_comment = "\033[38;5;247m"

func (e Gen) PreviewMethods(idxs Idxs, kind int, filter string) string {
	var bu strings.Builder
	bu.WriteString("Methods (" + idxs.GetWord(kind) + "):")
	arr := make([]string, 0)
	for k, v := range e.dict[kind] {
		str1 := idxs.GetWord(k)
		if strings.Contains(str1, filter) {
			color := color_word2
			// arr = append(arr, reset+color+str1+reset)                                 // idxs.GetWord(v.GetKind()
			arr = append(arr, color+str1+" "+reset+color_comment+v.Inspect(idxs)+reset) // idxs.GetWord(v.GetKind()
		}
	}
	sort.Strings(arr)
	for aa := range arr {
		line := arr[aa]
		//pars := strings.Split(line, "|||")
		bu.WriteString("\n\r " + line)
	}
	return bu.String()
}
