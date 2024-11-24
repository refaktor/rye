package evaldo

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"testing"

	"github.com/refaktor/rye/env"
	"github.com/refaktor/rye/loader"
)

func BenchmarkAutotype(b *testing.B) {
	tmpfile, err := os.CreateTemp("", "example.*.csv")
	if err != nil {
		b.Errorf("Error creating temporary file: %s", err)
	}
	defer tmpfile.Close()

	writer := bufio.NewWriter(tmpfile)

	rand.NewSource(12345)
	for i := 0; i < 1000000; i++ {
		//nolint:all
		line := fmt.Sprintf("%f,%f,%f,%f,%f\n", rand.Float64(), rand.Float64(), rand.Float64(), rand.Float64(), rand.Float64())
		_, err := writer.WriteString(line)
		if err != nil {
			b.Errorf("Error writing to temporary file: %s", err)
		}
	}
	writer.Flush()

	input := fmt.Sprintf("load\\csv file://%s |autotype 1.0", tmpfile.Name())
	block, genv := loader.LoadString(input, false)
	for i := 0; i < b.N; i++ {
		es := env.NewProgramState(block.(env.Block).Series, genv)
		RegisterBuiltins(es)
		res := EvalBlock(es)
		if res != nil && res.ErrorFlag {
			fmt.Println(res.Res.Print(*es.Idx))
		}
	}
}
