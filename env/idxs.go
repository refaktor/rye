package env

import (
	"fmt"
	"strconv"
)

type Idxs struct {
	words1 [1000]string
	words2 map[string]int
	wordsn int
}

func (e *Idxs) IndexWord(w string) int {
	idx, ok := e.words2[w]
	if ok {
		return idx
	} else {
		e.words1[e.wordsn] = w
		e.words2[w] = e.wordsn
		e.wordsn += 1
		return e.wordsn - 1
	}
}

func (e *Idxs) GetIndex(w string) (int, bool) {
	idx, ok := e.words2[w]
	if ok {
		return idx, true
	}
	return 0, false
}

func (e Idxs) GetWord(i int) string {
	return e.words1[i]
}

func (e Idxs) Probe() {
	fmt.Print("<IDXS: ")
	for i := 0; i < e.wordsn; i++ {
		fmt.Print(strconv.FormatInt(int64(i), 10) + ": " + e.words1[i] + " ")
	}
	fmt.Println(">")
}

func (e Idxs) GetWordCount() int {
	return e.wordsn
}

func NewIdxs() *Idxs {
	var e Idxs
	e.words2 = make(map[string]int)
	e.wordsn = 0
	return &e
}
