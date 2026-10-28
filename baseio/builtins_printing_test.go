package baseio

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/refaktor/rye/env"
	"github.com/refaktor/rye/evaldo"
)

func TestDisplayAndExploreRegistration(t *testing.T) {
	for _, name := range []string{"display", "explore", "explore\\custom", "_.."} {
		if builtins_printing_extra[name] == nil {
			t.Fatalf("missing %q builtin", name)
		}
	}
	if builtins_printing_extra["display\\custom"] != nil {
		t.Fatal("old interactive name still registered")
	}
}

func TestDisplayPrintsAllAndReturnsInput(t *testing.T) {
	ps := env.NewProgramStateOLD(*env.NewTSeries(nil), env.NewIdxs())
	evaldo.RegisterBuiltins(ps)
	items := make([]env.Object, 30)
	for i := range items {
		items[i] = *env.NewInteger(int64(i))
	}
	original := env.NewBlock(*env.NewTSeries(items))
	read, write, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	old := os.Stdout
	os.Stdout = write
	defer func() { os.Stdout = old }()
	result := builtins_printing_extra["display"].Fn(ps, original, nil, nil, nil, nil)
	write.Close()
	output, err := io.ReadAll(read)
	read.Close()
	if err != nil {
		t.Fatal(err)
	}
	if result != original || ps.FailureFlag || ps.ErrorFlag {
		t.Fatalf("display changed its argument: %#v", result)
	}
	lines := strings.Split(strings.TrimSuffix(string(output), "\n"), "\n")
	if len(lines) != 30 || lines[0] != "0" || lines[29] != "29" {
		t.Fatalf("output was truncated: %q", output)
	}
	if strings.Contains(string(output), "\x1b[") || strings.Contains(string(output), "enter=done") {
		t.Fatalf("unexpected interactive output: %q", output)
	}
}
