package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
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

	// Step 3: Add custom get-env builtin and any combinator
	ps.RegisterBuiltin("get-env", 1, "get-env key",
		func(ps *env.ProgramState, a0, a1, a2, a3, a4 env.Object) env.Object {
			if v := os.Getenv(a0.(env.String).Value); v != "" {
				return *env.NewString(v)
			}
			return *env.NewBoolean(false)
		})

	evaldo.RegisterBuiltinsFilter(ps, []string{"_*", "_+", "any"})

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

	// Demonstrate that computed values work
	if cacheAge, ok := ps.Ctx.GetInteger("cache-max-age", ps.Idx); ok {
		fmt.Printf("Cache max age: %d seconds\n", cacheAge.Value)
	}
	if maxBody, ok := ps.Ctx.GetInteger("max-body-kb", ps.Idx); ok {
		fmt.Printf("Max body size: %d KB\n", maxBody.Value)
	}

	tpl := template.Must(template.New("").Parse(
		`<html><body>{{.}}</body></html>`))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		slug := strings.TrimPrefix(r.URL.Path, "/")
		if slug == "" {
			slug = "index"
		}
		
		// Validate the slug to prevent path injection
		if strings.Contains(slug, "/") || strings.Contains(slug, "\\") || strings.Contains(slug, "..") {
			http.Error(w, "Invalid file name", http.StatusBadRequest)
			return
		}
		
		// Use filepath.Join for safe path construction
		filePath := filepath.Join(dir, slug+".md")
		
		// Verify the resolved path is still within the safe directory
		absPath, err := filepath.Abs(filePath)
		if err != nil {
			http.Error(w, "Invalid file path", http.StatusBadRequest)
			return
		}
		
		absDir, err := filepath.Abs(dir)
		if err != nil {
			http.Error(w, "Invalid directory path", http.StatusBadRequest)
			return
		}
		
		if !strings.HasPrefix(absPath, absDir+string(filepath.Separator)) {
			http.Error(w, "Invalid file path", http.StatusBadRequest)
			return
		}
		
		md, err := os.ReadFile(absPath)
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