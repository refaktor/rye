package baseio

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/refaktor/rye/env"
	"github.com/refaktor/rye/evaldo"
	"github.com/xuri/excelize/v2"
)

func TestTableFileBuiltinsOnlyInBaseIO(t *testing.T) {
	ps := env.NewProgramState()
	evaldo.RegisterBuiltins(ps)

	kind := ps.Idx.IndexWord("file-uri")
	for _, method := range []string{"Load\\csv__OLD", "Load\\csv", "Load\\csv\\ignore-bad", "Load\\tsv", "Save\\csv", "Save\\tsv", "Load\\xlsx", "Save\\xlsx"} {
		word := ps.Idx.IndexWord(method)
		if _, found := ps.Gen.Get(kind, word); found {
			t.Errorf("%s registered without baseio", method)
		}
	}

	Register(ps)
	for _, method := range []string{"Load\\csv__OLD", "Load\\csv", "Load\\csv\\ignore-bad", "Load\\tsv", "Save\\csv", "Save\\tsv", "Load\\xlsx", "Save\\xlsx"} {
		word := ps.Idx.IndexWord(method)
		if _, found := ps.Gen.Get(kind, word); !found {
			t.Errorf("%s not registered by baseio", method)
		}
	}

	for _, name := range []string{"parse\\csv", "parse\\tsv", "parse\\xlsx", "format\\csv", "format\\tsv", "format\\xlsx", "persistent-table"} {
		index, ok := ps.Idx.GetIndex(name)
		if !ok {
			t.Errorf("base table builtin %s not indexed", name)
			continue
		}
		if _, found := ps.Ctx.Get(index); !found {
			t.Errorf("base table builtin %s not registered", name)
		}
	}
}

func TestLoadXlsxRowsShorterThanHeader(t *testing.T) {
	ps := env.NewProgramState()
	file := excelize.NewFile()
	defer file.Close()
	header := []string{"a", "b", "c", "d"}
	short := []string{"only-a"}
	if err := file.SetSheetRow("Sheet1", "A1", &header); err != nil {
		t.Fatal(err)
	}
	if err := file.SetSheetRow("Sheet1", "A2", &short); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "short.xlsx")
	if err := file.SaveAs(path); err != nil {
		t.Fatal(err)
	}
	result := Builtins_table_files["file-uri//Load\\xlsx"].Fn(ps, *env.NewUri1(ps.Idx, "file://"+path), nil, nil, nil, nil)
	if ps.ErrorFlag || ps.FailureFlag {
		t.Fatalf("load failed: %v", result)
	}
	table, ok := result.(env.Table)
	if !ok {
		t.Fatalf("expected table, got %T", result)
	}
	if len(table.Rows) != 1 || len(table.Rows[0].Values) != 4 {
		t.Fatalf("row not padded to header: %#v", table.Rows)
	}
	for i := 1; i < 4; i++ {
		if v, ok := table.Rows[0].Values[i].(env.String); !ok || v.Value != "" {
			t.Fatalf("cell %d: %#v", i, table.Rows[0].Values[i])
		}
	}
}

func TestLoadXlsxSkipsEmptyRowsButPreservesGaps(t *testing.T) {
	ps := env.NewProgramState()
	file := excelize.NewFile()
	defer file.Close()
	header := []string{"a", "b", "c", "d"}
	if err := file.SetSheetRow("Sheet1", "A1", &header); err != nil {
		t.Fatal(err)
	}
	if err := file.SetCellStr("Sheet1", "A2", ""); err != nil {
		t.Fatal(err)
	}
	if err := file.SetCellStr("Sheet1", "C3", "value"); err != nil {
		t.Fatal(err)
	}
	if err := file.SetCellStr("Sheet1", "A1000", ""); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "empty-rows.xlsx")
	if err := file.SaveAs(path); err != nil {
		t.Fatal(err)
	}
	result := Builtins_table_files["file-uri//Load\\xlsx"].Fn(ps, *env.NewUri1(ps.Idx, "file://"+path), nil, nil, nil, nil)
	if ps.ErrorFlag || ps.FailureFlag {
		t.Fatalf("load failed: %v", result)
	}
	table, ok := result.(env.Table)
	if !ok {
		t.Fatalf("expected table, got %T", result)
	}
	if len(table.Rows) != 1 || len(table.Rows[0].Values) != 4 {
		t.Fatalf("expected one padded row, got %#v", table.Rows)
	}
	if v, ok := table.Rows[0].Values[2].(env.String); !ok || v.Value != "value" {
		t.Fatalf("missing value after empty cells: %#v", table.Rows[0].Values[2])
	}
}

func TestLoadXlsxRowWiderThanHeaderReturnsError(t *testing.T) {
	ps := env.NewProgramState()
	file := excelize.NewFile()
	defer file.Close()
	header := []string{"a"}
	wide := []string{"a", "b"}
	if err := file.SetSheetRow("Sheet1", "A1", &header); err != nil {
		t.Fatal(err)
	}
	if err := file.SetSheetRow("Sheet1", "A2", &wide); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "wide.xlsx")
	if err := file.SaveAs(path); err != nil {
		t.Fatal(err)
	}
	result := Builtins_table_files["file-uri//Load\\xlsx"].Fn(ps, *env.NewUri1(ps.Idx, "file://"+path), nil, nil, nil, nil)
	if !ps.FailureFlag {
		t.Fatalf("expected failure for extra cells, got %v", result)
	}
	if err, ok := result.(*env.Error); !ok || !strings.Contains(err.Message, "header has 1 columns") {
		t.Fatalf("unexpected error: %v", result)
	}
}
