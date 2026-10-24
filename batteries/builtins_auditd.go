//go:build !b_wasm

package batteries

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	libauparse "github.com/elastic/go-libaudit/auparse"

	"github.com/refaktor/rye/env"
	"github.com/refaktor/rye/evaldo"
)

// Auditd parsing builtins backed by github.com/elastic/go-libaudit/auparse
// Namespace: auditd
//
// - parse-line: String -> Dict
// - parse-file: file-uri -> Block of Dicts (one per line)
// - tail-parse: file-uri, Integer(max-lines) -> Block of Dicts (last N lines)

// export builtins (namespace: auditd)
var Builtins_auditd = map[string]*env.Builtin{
	"parse-line": {
		Argsn: 1,
		Doc:   "Parses a single auditd log line into a dict using go-libaudit auparse.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, _ env.Object, _ env.Object, _ env.Object, _ env.Object) env.Object {
			switch s := arg0.(type) {
			case env.String:
				m := parseLineWithAuparse(s.Value)
				if m == nil {
					ps.FailureFlag = true
					return evaldo.MakeBuiltinError(ps, "malformed auditd line", "auditd-parse-line")
				}
				return *env.NewDict(m)
			default:
				ps.FailureFlag = true
				return evaldo.MakeArgError(ps, 1, []env.Type{env.StringType}, "auditd-parse-line")
			}
		},
	},
	"parse-file": {
		Argsn: 1,
		Doc:   "Parses all lines of an auditd log file into a block of dicts using go-libaudit auparse.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, _ env.Object, _ env.Object, _ env.Object, _ env.Object) env.Object {
			switch u := arg0.(type) {
			case env.Uri:
				path := resolveAuditPath(ps.WorkingPath, u.GetPath())
				f, err := os.Open(path)
				if err != nil {
					return evaldo.MakeBuiltinError(ps, err.Error(), "auditd-parse-file")
				}
				defer f.Close()
				scanner := bufio.NewScanner(f)
				var items []env.Object
				for scanner.Scan() {
					if m := parseLineWithAuparse(scanner.Text()); m != nil {
						items = append(items, *env.NewDict(m))
					}
				}
				if err := scanner.Err(); err != nil {
					return evaldo.MakeBuiltinError(ps, err.Error(), "auditd-parse-file")
				}
				return *env.NewBlock(*env.NewTSeries(items))
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.UriType}, "auditd-parse-file")
			}
		},
	},
	"tail-parse": {
		Argsn: 2,
		Doc:   "Parses the last N lines of an auditd log file into a block of dicts using go-libaudit auparse.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, _ env.Object, _ env.Object, _ env.Object) env.Object {
			switch u := arg0.(type) {
			case env.Uri:
				path := resolveAuditPath(ps.WorkingPath, u.GetPath())
				lim, ok := arg1.(env.Integer)
				if !ok {
					return evaldo.MakeArgError(ps, 2, []env.Type{env.IntegerType}, "auditd-tail-parse")
				}
				lines, err := tailLastNLines(path, int(lim.Value))
				if err != nil {
					return evaldo.MakeBuiltinError(ps, err.Error(), "auditd-tail-parse")
				}
				items := make([]env.Object, 0, len(lines))
				for _, line := range lines {
					if m := parseLineWithAuparse(line); m != nil {
						items = append(items, *env.NewDict(m))
					}
				}
				return *env.NewBlock(*env.NewTSeries(items))
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.UriType}, "auditd-tail-parse")
			}
		},
	},
}

func resolveAuditPath(workingPath, path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(workingPath, path)
}

func parseLineWithAuparse(line string) map[string]any {
	line = strings.TrimSpace(line)
	if line == "" {
		return nil
	}
	rec, err := libauparse.ParseLogLine(line)
	if err != nil || rec == nil {
		return nil
	}
	m := make(map[string]any, 16)
	m["record_type"] = *env.NewString(rec.RecordType.String())
	m["record_type_num"] = *env.NewInteger(int64(rec.RecordType))
	m["timestamp"] = *env.NewString(rec.Timestamp.UTC().String())
	m["sequence"] = *env.NewInteger(int64(rec.Sequence))
	if data, err := rec.Data(); err == nil && data != nil {
		for k, v := range data {
			m[k] = *env.NewString(v)
		}
	}
	return m
}

// Simple last-N-lines reader without extra deps. Efficient for small N.
func tailLastNLines(path string, n int) ([]string, error) {
	if n <= 0 {
		return []string{}, nil
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	// Read entire file when reasonably small; otherwise fallback to scan
	info, err := f.Stat()
	if err == nil && info.Size() <= 4*1024*1024 { // 4MB
		b, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		lines := strings.Split(strings.TrimRight(string(b), "\n"), "\n")
		if len(lines) <= n {
			return lines, nil
		}
		return lines[len(lines)-n:], nil
	}
	// Fallback: read all lines streaming
	scanner := bufio.NewScanner(f)
	ring := make([]string, n)
	count := 0
	for scanner.Scan() {
		ring[count%n] = scanner.Text()
		count++
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if count <= n {
		return ring[:count], nil
	}
	out := make([]string, n)
	start := count % n
	copy(out, ring[start:])
	copy(out[count%n:], ring[:start])
	return out, nil
}
