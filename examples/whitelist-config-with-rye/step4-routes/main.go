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

	// Step 4: Add route registration
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

	// Add if and = for conditional logic
	evaldo.RegisterBuiltinsFilter(ps, []string{"_*", "_+", "any", "if", "_="})

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

	// Wire up collected routes
	for prefix, routeDir := range routes {
		routeDir := routeDir // capture loop variable
		fmt.Printf("Registering route: %s -> %s\n", prefix, routeDir)
		http.Handle(prefix+"/", http.StripPrefix(prefix,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
				filePath := filepath.Join(routeDir, slug+".md")
				
				// Verify the resolved path is still within the safe directory
				absPath, err := filepath.Abs(filePath)
				if err != nil {
					http.Error(w, "Invalid file path", http.StatusBadRequest)
					return
				}
				
				absRouteDir, err := filepath.Abs(routeDir)
				if err != nil {
					http.Error(w, "Invalid directory path", http.StatusBadRequest)
					return
				}
				
				if !strings.HasPrefix(absPath, absRouteDir+string(filepath.Separator)) {
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
			})))
	}

	// Default handler for root
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		slug := "index"
		
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