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

	// Step 6: Add debugging support with probe and enter-console
	routes := map[string]string{}
	ps.RegisterBuiltin("route", 2, "Defines a route",
		func(ps *env.ProgramState, a0, a1, a2, a3, a4 env.Object) env.Object {
			routes[a0.(env.String).Value] = a1.(env.String).Value
			return env.Void{}
		})

	// Custom get-env builtin
	ps.RegisterBuiltin("get-env", 1, "get-env key",
		func(ps *env.ProgramState, a0, a1, a2, a3, a4 env.Object) env.Object {
			if v := os.Getenv(a0.(env.String).Value); v != "" {
				return *env.NewString(v)
			}
			return *env.NewBoolean(false)
		})

	// Add probe and enter-console for debugging
	evaldo.RegisterBuiltinsFilter(ps, []string{"_*", "_+", "any", "if", "_=", "fn", "replace", "capitalize", "str", "probe", "enter-console"})

	// Execution limits for safety
	ps.MaxCallDepth = 50
	ps.MaxOps = 100_000

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

	fmt.Printf("Configuration loaded successfully!\n")
	fmt.Printf("Port: %s\n", port)
	fmt.Printf("Docs dir: %s\n", dir)

	tpl := template.Must(template.New("").Parse(
		`<html><head><title>{{.Title}}</title></head><body><h1>{{.Title}}</h1>{{.Content}}</body></html>`))

	// Wire up collected routes
	for prefix, routeDir := range routes {
		routeDir := routeDir // capture loop variable
		fmt.Printf("Registering route: %s -> %s\n", prefix, routeDir)
		http.Handle(prefix+"/", http.StripPrefix(prefix,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				slug := strings.TrimPrefix(r.URL.Path, "/")

				// Use page-title function if defined
				title := slug
				if fn, ok := ps.Ctx.GetFunction("page-title", ps.Idx); ok {
					fmt.Printf("DEBUG: Found page-title function, calling with slug: %s\n", slug)
					evaldo.CallFunctionArgsN(fn, ps, ps.Ctx, *env.NewString(slug))
					fmt.Printf("DEBUG: Function result: %v (type: %T)\n", ps.Res, ps.Res)
					if s, ok := ps.Res.(env.String); ok {
						title = s.Value
						fmt.Printf("DEBUG: Using title: %s\n", title)
					} else {
						fmt.Printf("DEBUG: Result is not a string: %T\n", ps.Res)
					}
				} else {
					fmt.Printf("DEBUG: page-title function not found\n")
				}

				path, err := safeMarkdownPath(routeDir, slug)
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

				data := struct {
					Title   string
					Content template.HTML
				}{
					Title:   title,
					Content: template.HTML(buf.String()),
				}
				tpl.Execute(w, data)
			})))
	}

	// Default handler for root
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		slug := "index"

		// Use page-title function if defined
		title := slug
		if fn, ok := ps.Ctx.GetFunction("page-title", ps.Idx); ok {
			fmt.Printf("DEBUG: Found page-title function, calling with slug: %s\n", slug)
			evaldo.CallFunctionArgsN(fn, ps, ps.Ctx, *env.NewString(slug))
			fmt.Printf("DEBUG: Function result: %v (type: %T)\n", ps.Res, ps.Res)
			if s, ok := ps.Res.(env.String); ok {
				title = s.Value
				fmt.Printf("DEBUG: Using title: %s\n", title)
			} else {
				fmt.Printf("DEBUG: Result is not a string: %T\n", ps.Res)
			}
		} else {
			fmt.Printf("DEBUG: page-title function not found\n")
		}

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

		data := struct {
			Title   string
			Content template.HTML
		}{
			Title:   title,
			Content: template.HTML(buf.String()),
		}
		tpl.Execute(w, data)
	})

	fmt.Printf("Serving on port %s\n", port)
	http.ListenAndServe(":"+port, nil)
}