//go:build !no_markdown
// +build !no_markdown

package evaldo

import (
	"bufio"
	"io"
	"strings"

	"github.com/refaktor/rye/env"
)

func load_markdown_Dict(ps *env.ProgramState, block env.Block) (env.Dict, *env.Error) {
	data := make(map[string]any)
	rmap := *env.NewDict(data)

	for block.Series.Pos() < block.Series.Len() {
		obj := block.Series.Peek()
		switch obj1 := obj.(type) {
		case env.Xword:
			key := ps.Idx.GetWord(obj1.Index)
			block.Series.Next()
			if nextObj, ok := block.Series.Peek().(env.Block); ok {
				block.Series.Next()
				rmap.Data[key] = nextObj
			} else {
				return _emptyRM(), MakeBuiltinError(ps, "Expected block after markdown section specifier.", "reader//do-markdown")
			}
		default:
			return _emptyRM(), MakeBuiltinError(ps, "Expected word specifying markdown section.", "reader//do-markdown")
		}
	}
	return rmap, nil
}

func do_markdown(ps *env.ProgramState, reader env.Object, rmap env.Dict) env.Object {
	file, ok := reader.(env.Native)
	if !ok {
		return MakeBuiltinError(ps, "Reader must be a file object.", "reader//do-markdown")
	}

	//fileObj, ok := file.Value.(*os.File)
	//if !ok {
	//	return MakeBuiltinError(ps, "Reader must be a file object.", "reader//do-markdown")
	// }
	fileObj, ok := file.Value.(io.Reader)
	if !ok {
		return MakeBuiltinError(ps, "Reader must be a reader object.", "reader//do-markdown")
	}

	scanner := bufio.NewScanner(fileObj)

	inCodeBlock := false
	codeLanguage := ""
	codeLines := []string{}

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "### ") {
			if block, ok := rmap.Data["H3"].(env.Block); ok {
				ps.Ser = block.Series
				EvalBlockInj(ps, *env.NewString(strings.TrimPrefix(line, "### ")), true)
			}
		} else if strings.HasPrefix(line, "## ") {
			if block, ok := rmap.Data["H2"].(env.Block); ok {
				ps.Ser = block.Series
				EvalBlockInj(ps, *env.NewString(strings.TrimPrefix(line, "## ")), true)
			}
		} else if strings.HasPrefix(line, "# ") {
			if block, ok := rmap.Data["H1"].(env.Block); ok {
				ps.Ser = block.Series
				EvalBlockInj(ps, *env.NewString(strings.TrimPrefix(line, "# ")), true)
			}
		} else if strings.HasPrefix(line, "```") {
			if inCodeBlock {
				if block, ok := rmap.Data["code"].(env.Block); ok {
					ps.Ser = block.Series
					EvalBlockInj(ps, *env.NewString(strings.Join(codeLines, "\n")), true)
					if codeLanguage != "" {
						EvalBlockInj(ps, *env.NewString(codeLanguage), true)
					}

				}
				inCodeBlock = false
				codeLines = []string{}
				codeLanguage = ""
			} else {
				inCodeBlock = true
				codeLanguage = strings.TrimPrefix(line, "```")
			}
		} else if inCodeBlock {
			codeLines = append(codeLines, line)
		} else if strings.HasPrefix(line, "* ") || strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "+ ") {
			if block, ok := rmap.Data["line-item"].(env.Block); ok {
				ps.Ser = block.Series
				EvalBlockInj(ps, *env.NewString(strings.TrimPrefix(line, "* ")), true)
			}
		} else if strings.TrimSpace(line) != "" {
			if block, ok := rmap.Data["paragraph"].(env.Block); ok {
				ps.Ser = block.Series
				EvalBlockInj(ps, *env.NewString(line), true)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return MakeBuiltinError(ps, err.Error(), "reader//do-markdown")
	}

	return env.Void{}
}

var Builtins_markdown = map[string]*env.Builtin{
	"reader//do-markdown": {
		Argsn: 2,
		Doc:   "Processes Markdown using a streaming approach with section handlers.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			rm, err := load_markdown_Dict(ps, arg1.(env.Block))
			if err != nil {
				ps.FailureFlag = true
				return err
			}
			return do_markdown(ps, arg0, rm)
		},
	},
}
