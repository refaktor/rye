// util.go
package util

import (
	"fmt"
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
