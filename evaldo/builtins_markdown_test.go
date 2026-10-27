//go:build !no_markdown

package evaldo

import (
	"strings"
	"testing"

	"github.com/refaktor/rye/env"
	"github.com/refaktor/rye/loader"
)

func TestMarkdownAvailableWithoutBatteries(t *testing.T) {
	code, idx := loader.LoadStringNoPEG(`markdown "# Title" |headings`, false)
	ps := env.NewProgramStateOLD(code.(env.Block).Series, idx)
	RegisterBaseBuiltins(ps)
	EvalExpression_CollectArg(ps, false, false)
	if ps.ErrorFlag || ps.FailureFlag {
		t.Fatalf("markdown builtin failed: %#v", ps.Res)
	}
	list, ok := ps.Res.(env.List)
	if !ok || len(list.Data) != 1 {
		t.Fatalf("markdown headings unavailable in base evaldo: %#v", ps.Res)
	}
	items := MarkdownDisplayItems("# Title\n\nA paragraph.")
	if len(items) < 2 {
		t.Fatalf("missing Markdown display sections: %#v", items)
	}
	first, ok := items[0].(map[string]interface{})
	if !ok || !strings.Contains(strings.Join(first["DisplayLines"].([]string), "\n"), "Title") {
		t.Fatalf("invalid Markdown display section: %#v", items[0])
	}
}
