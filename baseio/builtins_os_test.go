//go:build !no_os && !b_wasm

package baseio

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/refaktor/rye/env"
	"github.com/refaktor/rye/evaldo"
	"github.com/refaktor/rye/loader"
	find "github.com/refaktor/rye/util/finder"
)

func TestFinderPathModePerInstance(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "item.txt")
	if err := os.WriteFile(file, []byte("ok"), 0600); err != nil {
		t.Fatal(err)
	}
	ps := env.NewProgramStateOLD(*env.NewTSeries(nil), env.NewIdxs())
	newFinder := func() env.Object {
		f := Builtins_os["finder"].Fn(ps, *env.NewUri1(ps.Idx, "file://"+dir), nil, nil, nil, nil)
		return Builtins_os["finder//Name"].Fn(ps, f, *env.NewString("item.txt"), nil, nil, nil)
	}
	a := newFinder()
	b := newFinder()
	set := func(f env.Object, mode string) env.Object {
		t.Helper()
		result := Builtins_os["finder//Path-mode!"].Fn(ps, f, *env.NewString(mode), nil, nil, nil)
		if ps.ErrorFlag || ps.FailureFlag {
			t.Fatalf("setting %s mode failed: %#v", mode, result)
		}
		return result
	}
	get := func(f env.Object) env.Object {
		t.Helper()
		result := Builtins_os["finder//Eval"].Fn(ps, f, nil, nil, nil, nil)
		block, ok := result.(env.Block)
		if !ok || ps.ErrorFlag || ps.FailureFlag {
			t.Fatalf("unexpected finder result: %#v", result)
		}
		for _, item := range block.Series.S {
			switch v := item.(type) {
			case env.String:
				if v.Value == "item.txt" || v.Value == file {
					return item
				}
			case env.Uri:
				if v.GetPath() == file {
					return item
				}
			}
		}
		t.Fatalf("file not found in finder result: %#v", result)
		return nil
	}
	a = set(a, "filenames")
	b = set(b, "full")
	if got := get(a).(env.String).Value; got != "item.txt" {
		t.Fatalf("filenames mode: got %q", got)
	}
	if got := get(b).(env.String).Value; got != file {
		t.Fatalf("full mode: got %q, want %q", got, file)
	}
	a = set(a, "uris")
	if got := get(a).(env.Uri).GetPath(); got != file {
		t.Fatalf("uris mode: got %q, want %q", got, file)
	}
	if got := get(b).(env.String).Value; got != file {
		t.Fatalf("changing a modified b: got %q", got)
	}

	// Plain upstream finder values also retain mode on the returned Native copy.
	plain := *env.NewNative(ps.Idx, find.NewFind(dir), "finder")
	plain = set(plain, "filenames").(env.Native)
	if got := get(plain).(env.String).Value; got != "item.txt" {
		t.Fatalf("plain finder mode: got %q", got)
	}
}

func TestArchiveExtractionRejectsEscapes(t *testing.T) {
	for _, format := range []string{"zip", "tar.gz"} {
		t.Run(format, func(t *testing.T) {
			for _, tc := range []struct {
				name, member string
				makeSymlink  bool
				wantError    bool
			}{
				{"safe", "sub/file.txt", false, false},
				{"traversal", "../outside.txt", false, true},
				{"sibling-prefix", "../dest-other/outside.txt", false, true},
				{"absolute", "/tmp/archive-absolute-escape.txt", false, true},
				{"symlink-parent", "link/file.txt", true, true},
				{"symlink-target", "file.txt", true, true},
			} {
				t.Run(tc.name, func(t *testing.T) {
					parent := t.TempDir()
					dest := filepath.Join(parent, "dest")
					outside := filepath.Join(parent, "outside")
					if err := os.MkdirAll(dest, 0700); err != nil {
						t.Fatal(err)
					}
					if tc.makeSymlink {
						if err := os.MkdirAll(outside, 0700); err != nil {
							t.Fatal(err)
						}
						link := "link"
						if tc.name == "symlink-target" {
							link = "file.txt"
						}
						if err := os.Symlink(outside, filepath.Join(dest, link)); err != nil {
							t.Fatal(err)
						}
					}
					archive := filepath.Join(parent, "archive")
					if format == "zip" {
						out, err := os.Create(archive)
						if err != nil {
							t.Fatal(err)
						}
						zw := zip.NewWriter(out)
						w, err := zw.Create(tc.member)
						if err != nil {
							t.Fatal(err)
						}
						if _, err := w.Write([]byte("contents")); err != nil {
							t.Fatal(err)
						}
						if err := zw.Close(); err != nil {
							t.Fatal(err)
						}
						if err := out.Close(); err != nil {
							t.Fatal(err)
						}
					} else {
						out, err := os.Create(archive)
						if err != nil {
							t.Fatal(err)
						}
						gz := gzip.NewWriter(out)
						tw := tar.NewWriter(gz)
						if err := tw.WriteHeader(&tar.Header{Name: tc.member, Mode: 0600, Size: int64(len("contents")), Typeflag: tar.TypeReg}); err != nil {
							t.Fatal(err)
						}
						if _, err := tw.Write([]byte("contents")); err != nil {
							t.Fatal(err)
						}
						if err := tw.Close(); err != nil {
							t.Fatal(err)
						}
						if err := gz.Close(); err != nil {
							t.Fatal(err)
						}
						if err := out.Close(); err != nil {
							t.Fatal(err)
						}
					}
					var err error
					if format == "zip" {
						err = extractZip(archive, dest)
					} else {
						err = extractTarGz(archive, dest)
					}
					if (err != nil) != tc.wantError {
						t.Fatalf("extracting %q: got error %v, wantError %v", tc.member, err, tc.wantError)
					}
					if !tc.wantError {
						if data, err := os.ReadFile(filepath.Join(dest, tc.member)); err != nil || string(data) != "contents" {
							t.Fatalf("safe extraction: data %q, error %v", data, err)
						}
					}
				})
			}
		})
	}
}

func TestUnsetenvReturnsVoid(t *testing.T) {
	const key = "RYE_TEST_UNSETENV_RETURNS_VOID"
	t.Setenv(key, "present")
	ps := env.NewProgramStateOLD(*env.NewTSeries(nil), env.NewIdxs())
	result := Builtins_os["unsetenv!"].Fn(ps, *env.NewString(key), nil, nil, nil, nil)
	if result.Type() != env.VoidType || ps.FailureFlag || ps.ErrorFlag {
		t.Fatalf("expected Void success, got %#v", result)
	}
	if _, found := os.LookupEnv(key); found {
		t.Fatal("environment variable was not removed")
	}
}

func TestFinderAdditionalFilters(t *testing.T) {
	dir := t.TempDir()
	old := time.Now().Add(-72 * time.Hour)
	recent := time.Now().Add(-10 * time.Minute)
	// Evaluate the same duration expression used in Rye scripts.
	code, idx := loader.LoadStringNoPEG("2 .days", false)
	ps := env.NewProgramStateOLD(code.(env.Block).Series, idx)
	evaldo.RegisterBuiltins(ps)
	evaldo.EvalExpression_CollectArg(ps, false, false)
	twoDays, ok := ps.Res.(env.Integer)
	if !ok || ps.ErrorFlag || ps.FailureFlag || twoDays.Value != 172800000 {
		t.Fatalf("duration evaluation failed: %#v", ps.Res)
	}
	for name, content := range map[string]string{"recent.txt": "large content", "old.txt": "large content", "small.txt": "x", "skip.log": "large content"} {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte(content), 0600); err != nil {
			t.Fatal(err)
		}
		modified := recent
		if name == "old.txt" {
			modified = old
		}
		if err := os.Chtimes(path, modified, modified); err != nil {
			t.Fatal(err)
		}
	}
	type filter struct {
		name string
		arg  env.Object
	}
	for _, tc := range []struct {
		name    string
		filters []filter
		want    []string
	}{
		{"mtime-milliseconds", []filter{{"finder//Mtime-since!", env.Integer{Value: 3600000}}}, []string{"recent.txt", "skip.log", "small.txt"}},
		{"mtime-days", []filter{{"finder//Mtime-since!", twoDays}}, []string{"recent.txt", "skip.log", "small.txt"}},
		{"mtime-short", []filter{{"finder//Mtime-since!", env.Integer{Value: 3600}}}, []string{}},
		{"exclude-glob", []filter{{"finder//Exclude-path", *env.NewString("*.log")}}, []string{"old.txt", "recent.txt", "small.txt"}},
		{"exclude-substring", []filter{{"finder//Exclude-path", *env.NewString("recent")}}, []string{"old.txt", "skip.log", "small.txt"}},
		{"size", []filter{{"finder//Size>", env.Integer{Value: 1}}}, []string{"old.txt", "recent.txt", "skip.log"}},
		{"combined", []filter{{"finder//Size>", env.Integer{Value: 1}}, {"finder//Mtime-since!", env.Integer{Value: 3600000}}, {"finder//Exclude-path", *env.NewString("*.log")}}, []string{"recent.txt"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ps := env.NewProgramStateOLD(*env.NewTSeries(nil), env.NewIdxs())
			paths := *env.NewBlock(*env.NewTSeries([]env.Object{*env.NewString(dir)}))
			finder := Builtins_os["finder"].Fn(ps, paths, nil, nil, nil, nil)
			for _, f := range tc.filters {
				finder = Builtins_os[f.name].Fn(ps, finder, f.arg, nil, nil, nil)
				if ps.FailureFlag || ps.ErrorFlag {
					t.Fatalf("filter %s failed: %#v", f.name, finder)
				}
			}
			// Existing predicates must continue to work after attaching custom filters.
			finder = Builtins_os["finder//Type"].Fn(ps, finder, *env.NewString("f"), nil, nil, nil)
			for repeat := 0; repeat < 2; repeat++ {
				result := Builtins_os["finder//Eval"].Fn(ps, finder, nil, nil, nil, nil)
				block, ok := result.(env.Block)
				if !ok || ps.FailureFlag || ps.ErrorFlag {
					t.Fatalf("evaluation failed: %#v", result)
				}
				names := make([]string, block.Series.Len())
				for i, item := range block.Series.S {
					names[i] = filepath.Base(item.(env.Uri).GetPath())
				}
				if !reflect.DeepEqual(names, tc.want) {
					t.Fatalf("got %v, want %v", names, tc.want)
				}
			}
		})
	}
}
