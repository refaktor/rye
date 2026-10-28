//go:build !no_markdown

package baseio

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/refaktor/rye/env"
	"github.com/refaktor/rye/evaldo"
)

func TestDisplayMarkdownWithoutBatteries(t *testing.T) {
	ps := env.NewProgramStateOLD(*env.NewTSeries(nil), env.NewIdxs())
	evaldo.RegisterBaseBuiltins(ps)
	markdown := env.NewMarkdown("# Title\n\nA paragraph.")
	read, write, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	old := os.Stdout
	os.Stdout = write
	defer func() { os.Stdout = old }()
	result := builtins_printing_extra["display"].Fn(ps, *markdown, nil, nil, nil, nil)
	write.Close()
	output, err := io.ReadAll(read)
	read.Close()
	if err != nil {
		t.Fatal(err)
	}
	if result != *markdown || ps.ErrorFlag || ps.FailureFlag {
		t.Fatalf("display changed Markdown: %#v", result)
	}
	if !strings.Contains(string(output), "Title") || !strings.Contains(string(output), "A paragraph.") {
		t.Fatalf("Markdown not rendered without batteries: %q", output)
	}
}
