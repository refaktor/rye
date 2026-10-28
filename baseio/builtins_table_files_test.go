package baseio

import (
	"testing"

	"github.com/refaktor/rye/env"
	"github.com/refaktor/rye/evaldo"
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
