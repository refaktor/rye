// loader.go
package loader

import (
	"strings"
	"sync"

	"github.com/refaktor/rye/env"
)

func trace(x any) {
	//fmt.Print("\x1b[56m")
	//fmt.Print(x)
	//fmt.Println("\x1b[0m")
}

var wordIndex *env.Idxs
var wordIndexMutex sync.Mutex

func InitIndex() {
	if wordIndex == nil {
		wordIndex = env.NewIdxs()
	}
}

func GetIdxs() *env.Idxs {
	if wordIndex == nil {
		wordIndex = env.NewIdxs()
	}
	return wordIndex
}

func removeBangLine(content string) string {
	if strings.Index(content, "#!") == 0 {
		content = content[strings.Index(content, "\n")+1:]
	}
	return content
}
