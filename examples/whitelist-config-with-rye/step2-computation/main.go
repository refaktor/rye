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

func safeMarkdownPath(baseDir, slug string) (string, error) {
	if slug == "" {
		slug = "index"
	}

	// Reject path traversal and nested paths
	if strings.Contains(slug, "/") ||
		strings.Contains(slug, "\\") ||
		strings.Contains(slug, "..") {
		return "", fmt.Errorf("invalid file name")
	}

	filePath := filepath.Join(baseDir, slug+".md")

	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return "", err
	}

	absDir, err := filepath.Abs(baseDir)
	if err != nil {
		return "", err
	}

	if !strings.HasPrefix(absPath, absDir+string(filepath.Separator)) {
		return "", fmt.Errorf("invalid file path")
	}

	return absPath, nil
}

func main() {
	raw, err := os.ReadFile("config.rye")
	if err != nil {
		log.Fatalf("failed to read config: %v", err)
	}

	ps := env.NewProgramState()
	
	// Step 2: Add basic computation
	evaldo.RegisterBuiltinsFilter(ps, []string{"_*", "_+"})
	
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

		path, err := safeMarkdownPath(dir, slug)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		md, err := os.ReadFile(path)
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