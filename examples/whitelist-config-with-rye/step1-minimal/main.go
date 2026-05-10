package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/refaktor/rye/env"
	"github.com/refaktor/rye/evaldo"
	"github.com/refaktor/rye/loader"
	"github.com/yuin/goldmark"
)

func main() {
	raw, err := os.ReadFile("config.rye")
	if err != nil {
		log.Fatalf("failed to read config: %v", err)
	}

	ps := env.NewProgramState()
	blk := loader.LoadString(string(raw), false, ps)

	// Check for parse errors
	if errorObj, ok := blk.(env.Error); ok {
		log.Fatalf("parse error: %s", errorObj.Message)
	}

	evaldo.EvalBlock(ps, blk.(env.Block))

	// Check for runtime errors
	if ps.ErrorFlag {
		log.Fatalf("runtime error: %s", ps.Res.Print(*ps.Idx))
	}

	port := ps.Ctx.GetStringOr("port", ps.Idx, "8080")
	dir := ps.Ctx.GetStringOr("docs-dir", ps.Idx, "docs")
	tpl := template.Must(template.New("").Parse(
		`<html><body>{{.}}</body></html>`))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		slug := strings.TrimPrefix(r.URL.Path, "/")
		if slug == "" {
			slug = "index"
		}
		md, err := os.ReadFile(dir + "/" + slug + ".md")
		if err != nil {
			http.NotFound(w, r)
			return
		}
		var buf strings.Builder
		goldmark.Convert(md, &buf)
		tpl.Execute(w, template.HTML(buf.String()))
	})

	fmt.Printf("Serving on port %s\n", port)
	http.ListenAndServe(":"+port, nil)
}