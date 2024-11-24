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

func TestSpreadsheetLoadSaveXlsx(t *testing.T) {
	tmpfile1, err := os.CreateTemp("", "example.*.xlsx")
	if err != nil {
		t.Errorf("Error creating temporary file: %s", err)
	}
	defer tmpfile1.Close()

	input := fmt.Sprintf(`
		a: spreadsheet { "a" "b" "c" } { 1 1.1 "a" 2 2.2 "b" 3 3.3 "c" } 
		a .save\xlsx file://%s 
		b: load\xlsx file://%s |autotype 1.0
		a = b`,
		tmpfile1.Name(), tmpfile1.Name(),
	)
	block, genv := loader.LoadString(input, false)
	es := env.NewProgramState(block.(env.Block).Series, genv)
	RegisterBuiltins(es)
	res := EvalBlock(es)
	if res != nil && res.ErrorFlag {
		t.Errorf("Error: %s", res.Res.Print(*es.Idx))
	}
	if res.Res.Type() != env.IntegerType {
		t.Errorf("Expected integer type, got %T", res.Res.Type())
	}
	resVal := res.Res.(env.Integer).Value
	if resVal != 1 {
		fmt.Println(res.Res.Print(*es.Idx))
		t.Errorf("Expected spreadsheets to be identical but are not")
	}
}

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
