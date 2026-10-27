//go:build !no_sxml
// +build !no_sxml

package batteries

import (
	"encoding/xml"
	"strings"

	"github.com/refaktor/rye/env"
	"github.com/refaktor/rye/evaldo"

	"io"
)

// sxmlFrame is the output frame for one XML scope. A frame starts with an
// unknown shape and becomes a Dict or a Block depending on the first command
// used by a projection expression.
type sxmlFrameKind int

const (
	sxmlFrameUnknown sxmlFrameKind = iota
	sxmlFrameDict
	sxmlFrameList
)

type sxmlFrame struct {
	kind     sxmlFrameKind
	dict     map[string]any
	list     []env.Object
	tag      string
	attached bool
}

func newSxmlFrame(tag string) *sxmlFrame {
	return &sxmlFrame{tag: tag}
}

type sxmlSubBlock struct {
	depth int
	key   string
}

type sxmlCollector struct {
	frames    []*sxmlFrame
	subBlocks []sxmlSubBlock
	depth     int
	tag       string
}

func newSxmlCollector() *sxmlCollector {
	return &sxmlCollector{frames: []*sxmlFrame{newSxmlFrame("")}}
}

func (c *sxmlCollector) current() *sxmlFrame {
	return c.frames[len(c.frames)-1]
}

func (f *sxmlFrame) hasContent() bool {
	switch f.kind {
	case sxmlFrameDict:
		return len(f.dict) > 0
	case sxmlFrameList:
		return len(f.list) > 0
	default:
		return false
	}
}

func (f *sxmlFrame) object() env.Object {
	switch f.kind {
	case sxmlFrameDict:
		return *env.NewDict(f.dict)
	case sxmlFrameList:
		return *env.NewBlock(*env.NewTSeries(f.list))
	default:
		return *env.NewBlock(*env.NewTSeries(nil))
	}
}

// value returns the frame as a single value where possible. A one-item list
// frame is unwrapped so leaf text does not require |first |first after attach.
// Dict frames remain dicts; multi-item list frames remain blocks.
func (f *sxmlFrame) value() env.Object {
	if f.kind == sxmlFrameList && len(f.list) == 1 {
		return f.list[0]
	}
	return f.object()
}

func (f *sxmlFrame) ensureDict(ps *env.ProgramState) *env.Error {
	if f.kind == sxmlFrameUnknown {
		f.kind = sxmlFrameDict
		f.dict = make(map[string]any)
	}
	if f.kind != sxmlFrameDict {
		return evaldo.MakeBuiltinError(ps, "output frame is a list; keyed command requires a dict", "reader//Do-sxml")
	}
	return nil
}

func (f *sxmlFrame) ensureList(ps *env.ProgramState) *env.Error {
	if f.kind == sxmlFrameUnknown {
		f.kind = sxmlFrameList
		f.list = make([]env.Object, 0)
	}
	if f.kind != sxmlFrameList {
		return evaldo.MakeBuiltinError(ps, "output frame is a dict; list command requires a list", "reader//Do-sxml")
	}
	return nil
}

func (f *sxmlFrame) put(ps *env.ProgramState, value env.Object) *env.Error {
	if err := f.ensureList(ps); err != nil {
		return err
	}
	f.list = append(f.list, value)
	return nil
}

func (f *sxmlFrame) putKey(ps *env.ProgramState, key string, value env.Object) *env.Error {
	if err := f.ensureDict(ps); err != nil {
		return err
	}
	f.dict[key] = value
	return nil
}

func appendCollectionValue(obj env.Object, value env.Object) (env.Object, bool) {
	switch c := obj.(type) {
	case env.Block:
		items := append([]env.Object{}, c.Series.S...)
		items = append(items, value)
		return *env.NewBlock(*env.NewTSeries(items)), true
	case *env.Block:
		items := append([]env.Object{}, c.Series.S...)
		items = append(items, value)
		return *env.NewBlock(*env.NewTSeries(items)), true
	case env.List:
		items := append([]any{}, c.Data...)
		items = append(items, value)
		return *env.NewList(items), true
	case *env.List:
		items := append([]any{}, c.Data...)
		items = append(items, value)
		return *env.NewList(items), true
	default:
		return nil, false
	}
}

func collectionValues(obj env.Object) ([]env.Object, bool) {
	switch c := obj.(type) {
	case env.Block:
		return append([]env.Object{}, c.Series.S...), true
	case *env.Block:
		return append([]env.Object{}, c.Series.S...), true
	case env.List:
		items := make([]env.Object, 0, len(c.Data))
		for _, item := range c.Data {
			items = append(items, env.ToRyeValue(item))
		}
		return items, true
	case *env.List:
		items := make([]env.Object, 0, len(c.Data))
		for _, item := range c.Data {
			items = append(items, env.ToRyeValue(item))
		}
		return items, true
	default:
		return nil, false
	}
}

func dictData(obj env.Object) (map[string]any, bool) {
	switch d := obj.(type) {
	case env.Dict:
		return d.Data, true
	case *env.Dict:
		return d.Data, true
	default:
		return nil, false
	}
}

func (f *sxmlFrame) appendKey(ps *env.ProgramState, key string, value env.Object) *env.Error {
	if err := f.ensureDict(ps); err != nil {
		return err
	}
	existing, ok := f.dict[key]
	if !ok {
		f.dict[key] = *env.NewBlock(*env.NewTSeries([]env.Object{value}))
		return nil
	}
	if appended, ok := appendCollectionValue(existing.(env.Object), value); ok {
		f.dict[key] = appended
		return nil
	}
	// A scalar already occupies the key. Converting to a collection is more
	// useful than losing either value, and keeps attach/append predictable.
	f.dict[key] = *env.NewBlock(*env.NewTSeries([]env.Object{existing.(env.Object), value}))
	return nil
}

func (f *sxmlFrame) mergeKey(ps *env.ProgramState, key string, value env.Object) *env.Error {
	if err := f.ensureDict(ps); err != nil {
		return err
	}
	existing, ok := f.dict[key]
	if !ok {
		f.dict[key] = value
		return nil
	}
	existingObj, ok := existing.(env.Object)
	if !ok {
		return evaldo.MakeBuiltinError(ps, "cannot merge into a non-Rye value", "reader//Do-sxml")
	}
	if incoming, ok := dictData(value); ok {
		base, ok := dictData(existingObj)
		if !ok {
			return evaldo.MakeBuiltinError(ps, "merge-key expected a dict at key '"+key+"'", "reader//Do-sxml")
		}
		merged := make(map[string]any, len(base)+len(incoming))
		for k, v := range base {
			merged[k] = v
		}
		for k, v := range incoming {
			merged[k] = v
		}
		f.dict[key] = *env.NewDict(merged)
		return nil
	}
	if incoming, ok := collectionValues(value); ok {
		base, ok := collectionValues(existingObj)
		if !ok {
			return evaldo.MakeBuiltinError(ps, "merge-key expected a block or list at key '"+key+"'", "reader//Do-sxml")
		}
		base = append(base, incoming...)
		f.dict[key] = *env.NewBlock(*env.NewTSeries(base))
		return nil
	}
	return evaldo.MakeBuiltinError(ps, "merge-key value must be a dict, block, or list", "reader//Do-sxml")
}

func attachToFrame(ps *env.ProgramState, parent *sxmlFrame, child *sxmlFrame, key string, collection bool) *env.Error {
	if key == "" {
		key = child.tag
	}
	if key == "" {
		return evaldo.MakeBuiltinError(ps, "attach needs a key for the root frame", "reader//Do-sxml")
	}
	value := child.value()
	if parent.kind == sxmlFrameList {
		parent.list = append(parent.list, value)
		return nil
	}
	if parent.kind == sxmlFrameUnknown {
		parent.kind = sxmlFrameDict
		parent.dict = make(map[string]any)
	}
	if parent.kind != sxmlFrameDict {
		return evaldo.MakeBuiltinError(ps, "cannot attach a child to a list frame without a key", "reader//Do-sxml")
	}
	if collection {
		return parent.appendKey(ps, key, value)
	}

	// Direct attach: first occurrence sets the value itself. Repeats convert
	// the key to a collection, matching the natural cardinality of the XML.
	existing, ok := parent.dict[key]
	if !ok {
		parent.dict[key] = value
		return nil
	}
	existingObj, ok := existing.(env.Object)
	if !ok {
		return evaldo.MakeBuiltinError(ps, "cannot attach to a non-Rye value", "reader//Do-sxml")
	}
	if appended, ok := appendCollectionValue(existingObj, value); ok {
		parent.dict[key] = appended
		return nil
	}
	parent.dict[key] = *env.NewBlock(*env.NewTSeries([]env.Object{existingObj, value}))
	return nil
}

func mergeFrameIntoParent(ps *env.ProgramState, parent *sxmlFrame, child *sxmlFrame, key string) *env.Error {
	incoming, ok := dictData(child.object())
	if !ok {
		return evaldo.MakeBuiltinError(ps, "merge requires a dict-like child frame", "reader//Do-sxml")
	}
	if parent.kind == sxmlFrameList {
		return evaldo.MakeBuiltinError(ps, "cannot merge a dict into a list frame", "reader//Do-sxml")
	}
	if parent.kind == sxmlFrameUnknown {
		parent.kind = sxmlFrameDict
		parent.dict = make(map[string]any)
	}
	if parent.kind != sxmlFrameDict {
		return evaldo.MakeBuiltinError(ps, "merge requires a dict-like parent frame", "reader//Do-sxml")
	}
	if key == "" {
		for k, v := range incoming {
			parent.dict[k] = v
		}
		return nil
	}
	existing, ok := parent.dict[key]
	if !ok {
		// Copy the incoming dict so later changes to the child frame cannot
		// mutate the parent entry indirectly.
		copied := make(map[string]any, len(incoming))
		for k, v := range incoming {
			copied[k] = v
		}
		parent.dict[key] = *env.NewDict(copied)
		return nil
	}
	existingObj, ok := existing.(env.Object)
	if !ok {
		return evaldo.MakeBuiltinError(ps, "cannot merge into a non-Rye value", "reader//Do-sxml")
	}
	base, ok := dictData(existingObj)
	if !ok {
		return evaldo.MakeBuiltinError(ps, "merge key '"+key+"' does not contain a dict", "reader//Do-sxml")
	}
	merged := make(map[string]any, len(base)+len(incoming))
	for k, v := range base {
		merged[k] = v
	}
	for k, v := range incoming {
		merged[k] = v
	}
	parent.dict[key] = *env.NewDict(merged)
	return nil
}

func sxmlCommandName(ps *env.ProgramState, obj env.Object) (string, bool) {
	switch w := obj.(type) {
	case env.Tagword:
		return ps.Idx.GetWord(w.Index), true
	case env.Word:
		return ps.Idx.GetWord(w.Index), true
	default:
		return "", false
	}
}

func sxmlKeyString(ps *env.ProgramState, obj env.Object) (string, bool) {
	switch v := obj.(type) {
	case env.String:
		return v.Value, true
	case env.Word:
		return ps.Idx.GetWord(v.Index), true
	case env.Tagword:
		return ps.Idx.GetWord(v.Index), true
	default:
		return "", false
	}
}

func applySxmlCommand(ps *env.ProgramState, command string, args []env.Object, collector *sxmlCollector) env.Object {
	if len(collector.frames) == 0 {
		return evaldo.MakeBuiltinError(ps, "no output frame", "reader//Do-sxml")
	}
	current := collector.current()

	switch command {
	case "put":
		if len(args) != 1 {
			return evaldo.MakeBuiltinError(ps, "put expects exactly 1 value", "reader//Do-sxml")
		}
		if err := current.put(ps, args[0]); err != nil {
			return err
		}
	case "put-key":
		if len(args) != 2 {
			return evaldo.MakeBuiltinError(ps, "put-key expects a key and a value", "reader//Do-sxml")
		}
		key, ok := sxmlKeyString(ps, args[0])
		if !ok {
			return evaldo.MakeBuiltinError(ps, "put-key key must be a string, word, or tagword", "reader//Do-sxml")
		}
		if err := current.putKey(ps, key, args[1]); err != nil {
			return err
		}
	case "sub-block":
		if collector.depth == 0 {
			return evaldo.MakeBuiltinError(ps, "sub-block requires an open XML element", "reader//Do-sxml")
		}
		key := collector.tag
		rest := args
		if len(args) > 0 {
			var ok bool
			key, ok = sxmlKeyString(ps, args[0])
			if !ok {
				return evaldo.MakeBuiltinError(ps, "sub-block key must be a string, word, or tagword", "reader//Do-sxml")
			}
			rest = args[1:]
		}
		if key == "" {
			return evaldo.MakeBuiltinError(ps, "sub-block needs a key or an open named element", "reader//Do-sxml")
		}
		if len(rest)%2 != 0 {
			return evaldo.MakeBuiltinError(ps, "sub-block initial values must be key/value pairs", "reader//Do-sxml")
		}
		child := newSxmlFrame(key)
		for i := 0; i < len(rest); i += 2 {
			field, ok := sxmlKeyString(ps, rest[i])
			if !ok {
				return evaldo.MakeBuiltinError(ps, "sub-block field names must be strings, words, or tagwords", "reader//Do-sxml")
			}
			if err := child.putKey(ps, field, rest[i+1]); err != nil {
				return err
			}
		}
		collector.frames = append(collector.frames, child)
		collector.subBlocks = append(collector.subBlocks, sxmlSubBlock{depth: collector.depth, key: key})
	case "append-key":
		if len(args) != 2 {
			return evaldo.MakeBuiltinError(ps, "append-key expects a key and a value", "reader//Do-sxml")
		}
		key, ok := sxmlKeyString(ps, args[0])
		if !ok {
			return evaldo.MakeBuiltinError(ps, "append-key key must be a string, word, or tagword", "reader//Do-sxml")
		}
		if err := current.appendKey(ps, key, args[1]); err != nil {
			return err
		}
	case "merge-key":
		if len(args) != 2 {
			return evaldo.MakeBuiltinError(ps, "merge-key expects a key and a dict/block/list", "reader//Do-sxml")
		}
		key, ok := sxmlKeyString(ps, args[0])
		if !ok {
			return evaldo.MakeBuiltinError(ps, "merge-key key must be a string, word, or tagword", "reader//Do-sxml")
		}
		if err := current.mergeKey(ps, key, args[1]); err != nil {
			return err
		}
	case "attach", "append", "merge":
		if len(args) > 1 {
			return evaldo.MakeBuiltinError(ps, command+" expects zero or one key", "reader//Do-sxml")
		}
		if len(collector.frames) < 2 {
			return evaldo.MakeBuiltinError(ps, command+" requires an open sub-block", "reader//Do-sxml")
		}
		key := ""
		if len(args) == 1 {
			var ok bool
			key, ok = sxmlKeyString(ps, args[0])
			if !ok {
				return evaldo.MakeBuiltinError(ps, command+" key must be a string, word, or tagword", "reader//Do-sxml")
			}
		}
		parent := collector.frames[len(collector.frames)-2]
		if command == "merge" {
			if err := mergeFrameIntoParent(ps, parent, current, key); err != nil {
				return err
			}
		} else if err := attachToFrame(ps, parent, current, key, command == "append"); err != nil {
			return err
		}
		current.attached = true
	default:
		return evaldo.MakeBuiltinError(ps, "unknown projection command '"+command+"'", "reader//Do-sxml")
	}
	return nil
}

// projectionUsesInjection reports whether the first token of a projection
// argument should receive the current XML event as an injected left value.
// Literals such as strings and numbers must not receive it: otherwise the
// evaluator emits the injected value itself as an extra result.
func projectionUsesInjection(obj env.Object) bool {
	switch obj.(type) {
	case env.Dotword, env.Pipeword, env.Opword, env.LModword:
		return true
	default:
		return false
	}
}

// runSxmlProjection evaluates a ( ... ) projection block. The first element is
// a literal command and is never evaluated. The remaining comma-separated
// expressions are evaluated with the current XML event value injected when the
// expression starts with an operation that expects a left value.
func runSxmlProjection(ps *env.ProgramState, block env.Block, injected env.Object, collector *sxmlCollector) env.Object {
	if block.Series.Len() == 0 {
		return evaldo.MakeBuiltinError(ps, "empty projection block", "reader//Do-sxml")
	}
	command, ok := sxmlCommandName(ps, block.Series.Get(0))
	if !ok {
		return evaldo.MakeBuiltinError(ps, "projection must start with a literal command such as 'put-key", "reader//Do-sxml")
	}

	ser := ps.Ser
	defer func() { ps.Ser = ser }()

	args := make([]env.Object, 0, 2)
	segment := make([]env.Object, 0, 4)

	evalSegment := func() env.Object {
		if len(segment) == 0 {
			return nil
		}
		ps.Ser = *env.NewTSeries(segment)
		if projectionUsesInjection(segment[0]) {
			evaldo.EvalExpressionInj(ps, injected, true)
		} else {
			evaldo.EvalExpression_CollectArg(ps, false, false)
		}
		if ps.ErrorFlag || ps.FailureFlag || ps.ReturnFlag {
			if ps.Res != nil {
				return ps.Res
			}
			return evaldo.AsError(ps)
		}
		args = append(args, ps.Res)
		return nil
	}

	for i := 1; i < block.Series.Len(); i++ {
		obj := block.Series.Get(i)
		if _, ok := obj.(env.Comma); ok {
			if errObj := evalSegment(); errObj != nil {
				return errObj
			}
			segment = segment[:0]
			continue
		}
		segment = append(segment, obj)
	}
	if errObj := evalSegment(); errObj != nil {
		return errObj
	}
	return applySxmlCommand(ps, command, args, collector)
}

func runSxmlHandler(ps *env.ProgramState, handler env.Object, injected env.Object, collector *sxmlCollector) env.Object {
	switch obj := handler.(type) {
	case env.Block:
		switch obj.Mode {
		case 1: // executable [ ] code
			ser := ps.Ser
			defer func() { ps.Ser = ser }()
			ps.Ser = obj.Series
			evaldo.EvalBlockInj(ps, injected, true)
			evaldo.MaybeDisplayFailureOrError(ps, ps.Idx, "do-sxml")
			if ps.ErrorFlag || ps.FailureFlag || ps.ReturnFlag {
				if ps.Res != nil {
					return ps.Res
				}
				return evaldo.AsError(ps)
			}
			return nil
		case 2: // projection ( ... )
			return runSxmlProjection(ps, obj, injected, collector)
		default:
			return evaldo.MakeBuiltinError(ps, "SXML handler must be a code block [ ] or a projection block ( )", "reader//Do-sxml")
		}
	default:
		return evaldo.MakeBuiltinError(ps, "SXML handler must be a block", "reader//Do-sxml")
	}
}

func load_saxml_Dict(ps *env.ProgramState, block env.Block) (env.Dict, *env.Error) {
	var keys []string

	data := make(map[string]any)
	rmap := *env.NewDict(data)

	for block.Series.Pos() < block.Series.Len() {
		obj := block.Series.Peek()
		switch obj1 := obj.(type) {
		case env.Opword:
			keys = append(keys, ps.Idx.GetWord(obj1.Index))
			block.Series.Next()
			continue
		case env.Tagword:
			name := ps.Idx.GetWord(obj1.Index)
			switch name {
			case "start", "end", "any":
				keys = append(keys, "-"+name+"-")
			default:
				return EmptyRM(), evaldo.MakeBuiltinError(ps, "Unsupported SXML event '"+name+"'. Use 'start, 'end, or 'any.", "reader//Do-sxml")
			}
			block.Series.Next()
			continue
		case env.Void:
			keys = append(keys, "")
			block.Series.Next()
			continue
		case env.Block:
			block.Series.Next()
			switch obj1.Mode {
			case 1, 2:
				// [ ] code or ( ) projection. With no selector, both act as
				// the scope's start handler.
				if len(keys) > 0 {
					for _, k := range keys {
						rmap.Data[k] = obj1
					}
					keys = []string{}
				} else {
					rmap.Data["-start-"] = obj1
				}
			case 0:
				rm, err := load_saxml_Dict(ps, obj1)
				if err != nil {
					return EmptyRM(), err
				}
				if len(keys) > 0 {
					for _, k := range keys {
						rmap.Data[k] = rm
					}
					keys = []string{}
				} else {
					return EmptyRM(), evaldo.MakeBuiltinError(ps, "No selectors before tag map.", "reader//Do-sxml")
				}
			default:
				return EmptyRM(), evaldo.MakeBuiltinError(ps, "SXML handler must use [ ] for code, ( ) for projection, or { } for nested rules", "reader//Do-sxml")
			}
		default:
			return EmptyRM(), evaldo.MakeBuiltinError(ps, "Unknown type in block parsing: "+obj.Inspect(*ps.Idx), "reader//Do-sxml")
		}
	}
	return rmap, nil
}

type sxmlHandlerScope struct {
	depth        int
	tag          string
	prevRules    env.Dict
	restoreRules bool
}

func do_sxml(ps *env.ProgramState, reader io.Reader, rmap env.Dict) env.Object {
	var elementStack []string
	var handlerScopes []sxmlHandlerScope
	collector := newSxmlCollector()
	decoder := xml.NewDecoder(reader)

	for {
		t, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return evaldo.MakeBuiltinError(ps, "XML parse error: "+err.Error(), "reader//Do-sxml")
		}
		switch se := t.(type) {
		case xml.StartElement:
			tag := se.Name.Local
			elementStack = append(elementStack, tag)
			depth := len(elementStack)
			collector.depth = depth
			collector.tag = tag

			handler, ok := rmap.Data[tag]
			if !ok {
				handler, ok = rmap.Data["-any-"]
			}
			if !ok {
				continue
			}

			startValue := *env.NewNative(ps.Idx, se, "rye-sxml-start")
			switch obj := handler.(type) {
			case env.Dict:
				handlerScopes = append(handlerScopes, sxmlHandlerScope{
					depth:        depth,
					tag:          tag,
					prevRules:    rmap,
					restoreRules: true,
				})
				rmap = obj
				if startHandler, exists := rmap.Data["-start-"]; exists {
					startObj, ok := startHandler.(env.Object)
					if !ok {
						return evaldo.MakeBuiltinError(ps, "invalid SXML start handler", "reader//Do-sxml")
					}
					if errObj := runSxmlHandler(ps, startObj, startValue, collector); errObj != nil {
						return errObj
					}
				}
			case env.Block:
				if errObj := runSxmlHandler(ps, obj, startValue, collector); errObj != nil {
					return errObj
				}
				handlerScopes = append(handlerScopes, sxmlHandlerScope{
					depth:        depth,
					tag:          tag,
					prevRules:    rmap,
					restoreRules: false,
				})
			default:
				return evaldo.MakeBuiltinError(ps, "SXML handler for <"+tag+"> must be a block or nested rules map", "reader//Do-sxml")
			}

		case xml.CharData:
			if handler, ok := rmap.Data[""]; ok {
				handlerObj, ok := handler.(env.Object)
				if !ok {
					return evaldo.MakeBuiltinError(ps, "invalid SXML text handler", "reader//Do-sxml")
				}
				if errObj := runSxmlHandler(ps, handlerObj, *env.NewString(string(se.Copy())), collector); errObj != nil {
					return errObj
				}
			}

		case xml.EndElement:
			if len(elementStack) == 0 {
				continue
			}
			depth := len(elementStack)
			if se.Name.Local != elementStack[depth-1] {
				continue
			}
			collector.depth = depth
			collector.tag = se.Name.Local
			if len(handlerScopes) > 0 && handlerScopes[len(handlerScopes)-1].depth == depth {
				scope := handlerScopes[len(handlerScopes)-1]

				if endHandler, ok := rmap.Data["-end-"]; ok {
					endObj, ok := endHandler.(env.Object)
					if !ok {
						return evaldo.MakeBuiltinError(ps, "invalid SXML end handler", "reader//Do-sxml")
					}
					endValue := *env.NewNative(ps.Idx, se, "rye-sxml-end")
					if errObj := runSxmlHandler(ps, endObj, endValue, collector); errObj != nil {
						return errObj
					}
				}

				// Close sub-blocks opened while this element was active.
				// Unattached sub-blocks attach to their parent under the key
				// given to sub-block; attached ones have already been placed.
				for len(collector.subBlocks) > 0 && collector.subBlocks[len(collector.subBlocks)-1].depth == depth {
					sub := collector.subBlocks[len(collector.subBlocks)-1]
					collector.subBlocks = collector.subBlocks[:len(collector.subBlocks)-1]
					child := collector.current()
					collector.frames = collector.frames[:len(collector.frames)-1]
					if !child.attached {
						parent := collector.current()
						if errObj := attachToFrame(ps, parent, child, sub.key, false); errObj != nil {
							return errObj
						}
					}
				}
				if scope.restoreRules {
					rmap = scope.prevRules
				}
				handlerScopes = handlerScopes[:len(handlerScopes)-1]
			}
			elementStack = elementStack[:depth-1]

		default:
			// Comments, processing instructions, directives, and other XML
			// tokens are currently not projected.
		}
	}

	root := collector.frames[0]
	if root.hasContent() {
		return root.object()
	}
	return *env.NewBoolean(true)
}

var Builtins_sxml = map[string]*env.Builtin{

	//
	// ##### SXML ##### "Streaming, SAX-like XML processing"
	//
	// { <person> [ .print ] }
	// { <person> { _ [ .print ] <name> <surname> <age> { _ [ .print2 ";" ] } }
	// { <person> ( 'put-key , "id" , .Attr? "id" ) }
	// { <person> ( 'sub-block , "people" , "id" , .Attr? "id" ) }
	//
	// Tests:
	// stdout {
	//   "<scene><bot>C3PO</bot><bot>R2D2</bot><jedi>Luke</jedi></scene>" |reader
	//   |Do-sxml { _ [ .prns ] }
	// } "C3PO R2D2 Luke "
	// stdout {
	//   "<scene><bot>C3PO</bot><bot>R2D2</bot><jedi>Luke</jedi></scene>" |reader
	//   |Do-sxml { <bot> { _ [ .prns ] } }
	// } "C3PO R2D2 "
	// stdout {
	//   "<scene><ship>XWing</ship><bot>R2D2</bot><jedi>Luke</jedi></scene>" |reader
	//   |Do-sxml { <bot> <jedi> { _ [ .prns ] } }
	// } "R2D2 Luke "
	// stdout {
	//   "<scene><xwing><bot>R2D2</bot><person>Luke</person></xwing><destroyer><person>Vader</person></destroyer></scene>" |reader
	//   |Do-sxml { <xwing> { <person> { _ [ .prns ] } } }
	// } "Luke "
	// equal {
	//   r: `<r><i id="1"/></r>` |reader |Do-sxml {
	//     <i> { 'start ( 'put-key , "id" , .Attr? "id" ) }
	//   }
	//   r -> "id"
	// } "1"
	// equal {
	//   r: `<r><i>A</i><i>B</i></r>` |reader |Do-sxml { _ ( 'put , .trim ) }
	//   r |length?
	// } 2
	// equal {
	//   r: `<r><i>A</i><i>B</i></r>` |reader |Do-sxml { _ ( 'append-key , "texts" , .trim ) }
	//   r -> "texts" |length?
	// } 2
	// equal {
	//   r: `<r><x>A</x></r>` |reader |Do-sxml {
	//     <x> { 'end ( 'put-key , "name" , .Name? ) }
	//   }
	//   r -> "name"
	// } "x"
	// equal {
	//   r: `<r><book id="1"><title>Dune</title></book></r>` |reader |Do-sxml {
	//     <book> {
	//       'start ( 'sub-block , "book" , "id" , .Attr? "id" )
	//       <title> { _ ( 'put-key , "title" , .trim ) }
	//     }
	//   }
	//   r -> "book" -> "title"
	// } "Dune"
	// equal {
	//   r: `<r><i id="1"/><i id="2"/></r>` |reader |Do-sxml {
	//     <i> { 'start ( 'sub-block , "items" , "id" , .Attr? "id" ) }
	//   }
	//   r -> "items" |length?
	// } 2
	// equal {
	//   r: `<r><i>A</i></r>` |reader |Do-sxml {
	//     <i> {
	//       'start ( 'sub-block , "tmp" )
	//       _ ( 'put , .trim )
	//       'end ( 'append , "items" )
	//     }
	//   }
	//   r -> "items" |length?
	// } 1
	// equal {
	//   r: `<r><head><x>1</x></head></r>` |reader |Do-sxml {
	//     <head> {
	//       'start ( 'sub-block , "head" )
	//       <x> { _ ( 'put-key , "x" , .trim ) }
	//       'end ( 'merge )
	//     }
	//   }
	//   r -> "x"
	// } "1"
	// Example:
	// ; [ ] executes code at XML events
	// `<r><i>A</i></r>` |reader |Do-sxml { _ [ .prns ] }
	//
	// ; ( ) projects event values into a collected structure
	// `<r><i id="1">A</i></r>` |reader |Do-sxml {
	//   <i> {
	//     'start ( 'sub-block , "items" , "id" , .Attr? "id" )
	//     _ ( 'put-key , "text" , .trim )
	//   }
	// }
	// Args:
	// * reader: XML reader object
	// * block: SXML processing block with tag handlers and optional ( ) projections
	// Returns:
	// * collected structure when ( ) projections are used, otherwise true
	"reader//Do-sxml": {
		Argsn: 2,
		Doc:   "Processes XML with streaming tag handlers. [ ] runs code, { } nests rules, and ( 'command , ... ) projects values with put, put-key, append-key, merge-key, sub-block, attach, append, and merge.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			reader, ok := arg0.(env.Native)
			if !ok {
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "reader//Do-sxml")
			}
			ioReader, ok := reader.Value.(io.Reader)
			if !ok {
				return evaldo.MakeBuiltinError(ps, "reader native does not implement io.Reader", "reader//Do-sxml")
			}
			block, ok := arg1.(env.Block)
			if !ok {
				return evaldo.MakeArgError(ps, 2, []env.Type{env.BlockType}, "reader//Do-sxml")
			}
			rm, err := load_saxml_Dict(ps, block)
			if err != nil {
				ps.FailureFlag = true
				return err
			}
			return do_sxml(ps, ioReader, rm)
		},
	},

	// TODO ... make a do-sxml function that accepts a string

	// Tests:
	// ; TODO: method resolution for rye-sxml-start//Attr? not working - needs investigation
	// ; stdout {
	// ;   `<scene><ship type="xwing"><person age="25">Luke</person></ship><ship type="destroyer"><person age="55">Vader</person></ship></scene>` |reader
	// ;   .do-sxml { <ship> [ .Attr? 0 |last |prns ] }
	// ; } "xwing destroyer "
	// ; stdout {
	// ;   `<scene><ship type="xwing"><person age="25">Luke</person></ship><ship type="destroyer"><person age="55">Vader</person></ship></scene>` |reader
	// ;   .do-sxml { <person> [ .Attr? 0 |last |prns ] }
	// ; } "25 55 "
	// Example:
	// ; inside an executable [ ] handler
	// `<r><x id="a"/></r>` |reader |Do-sxml { <x> [ .Attr? "id" |prns ] }
	//
	// ; inside a projection ( ) handler
	// `<r><x id="a"/></r>` |reader |Do-sxml {
	//   <x> { 'start ( 'put-key , "id" , .Attr? "id" ) }
	// }
	// Args:
	// * element: XML start element
	// * index: Integer index of the Attribute to retrieve
	// Returns:
	// * list [ namespace tag value ] of the attribute or void if not found
	"rye-sxml-start//Attr?": {
		Argsn: 2,
		Doc:   "Retrieves an attribute by index or name from an XML start element.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch obj := arg0.(type) {
			case env.Native:
				switch obj1 := obj.Value.(type) {
				case xml.StartElement:
					switch n := arg1.(type) {
					case env.Integer:
						if n.Value >= 0 && int(n.Value) < len(obj1.Attr) {
							attr := obj1.Attr[int(n.Value)]
							return *env.NewList([]any{
								*env.NewString(attr.Name.Space),
								*env.NewString(attr.Name.Local),
								*env.NewString(attr.Value),
							})
						} else {
							return evaldo.MakeBuiltinError(ps, "Attribute index out of bounds.", "rye-sxml-start//Attr?")
						}
					case env.Word:
						attrName := ps.Idx.GetWord(n.Index)
						for _, attr := range obj1.Attr {
							if attr.Name.Local == attrName {
								// If you provide the argument name or the "namespace:arg"
								// I think it makes more sense to just return value, simpler code
								// to use the value forward?
								return *env.NewString(attr.Value)
							}
						}
						return evaldo.MakeBuiltinError(ps, "Attribute '"+attrName+"' not found.", "rye-sxml-start//Attr?")
					case env.String:
						attrName := n.Value
						// Check if string contains namespace (format: "namespace:attrname")
						namespaceAndName := strings.Split(attrName, ":")
						if len(namespaceAndName) == 2 {
							// Has namespace, match both namespace and local name
							targetNS := namespaceAndName[0]
							targetLocal := namespaceAndName[1]
							for _, attr := range obj1.Attr {
								if attr.Name.Space == targetNS && attr.Name.Local == targetLocal {
									return *env.NewString(attr.Value)
								}
							}
						} else {
							// No namespace, match only local name
							for _, attr := range obj1.Attr {
								if attr.Name.Local == attrName {
									return *env.NewString(attr.Value)
								}
							}
						}
						return evaldo.MakeBuiltinError(ps, "Attribute '"+attrName+"' not found.", "rye-sxml-start//Attr?")
					default:
						return evaldo.MakeArgError(ps, 2, []env.Type{env.IntegerType, env.WordType, env.StringType}, "rye-sxml-start//Attr?")
					}
				default:
					return evaldo.MakeBuiltinError(ps, "Not xml-start element.", "rye-sxml-start//Attr?")
				}
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "rye-sxml-start//Attr?")
			}
		},
	},

	// Example:
	// ; inside an executable [ ] handler
	// `<r><x id="a"/></r>` |reader |Do-sxml { <x> [ .Attrs? |print ] }
	//
	// ; inside a projection ( ) handler
	// `<r><x id="a"/></r>` |reader |Do-sxml {
	//   <x> { 'start ( 'put-key , "attrs" , .Attrs? ) }
	// }
	// Args:
	// * element: XML start element
	// Returns:
	// * dict with all attributes where keys are "attrname" or "namespace:attrname"
	"rye-sxml-start//Attrs?": {
		Argsn: 1,
		Doc:   "Returns a dict of all attributes from an XML start element.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch obj := arg0.(type) {
			case env.Native:
				switch obj1 := obj.Value.(type) {
				case xml.StartElement:
					data := make(map[string]any)
					for _, attr := range obj1.Attr {
						var key string
						if attr.Name.Space != "" {
							key = attr.Name.Space + ":" + attr.Name.Local
						} else {
							key = attr.Name.Local
						}
						data[key] = *env.NewString(attr.Value)
					}
					return *env.NewDict(data)
				default:
					return evaldo.MakeBuiltinError(ps, "Not xml-start element.", "rye-sxml-start//Attrs?")
				}
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "rye-sxml-start//Attrs?")
			}
		},
	},

	// Tests:
	// ; TODO: These SXML tests need investigation - actual output differs from expected
	// ; stdout {
	// ;   "<scene><xwing><bot>R2D2</bot><person><name>Luke</name></person></xwing><destroyer><person>Vader</person></destroyer></scene>" |reader
	// ;   .do-sxml { <xwing> { 'start [ prns "YYY" ] <bot> [ print "***" ] 'any [ .Name? .probe ] 'end [ print "xx" ] } }
	// ; } "bot R2D2 \nperson name Luke \n"
	// ; stdout {
	// ;   "<root><a>1</a><b>2</b></root>" |reader
	// ;   .do-sxml { 'any [ .Name? .prn ":" ] 'end [ print "---" ] }
	// ; } "root:\na:\n---\nb:\n---\n---\n"
	// Example:
	// ; inside an executable [ ] handler
	// `<r><x/></r>` |reader |Do-sxml { <x> { 'end [ .Name? |prns ] } }
	//
	// ; inside a projection ( ) handler
	// `<r><x/></r>` |reader |Do-sxml {
	//   <x> { 'start ( 'put-key , "name" , .Name? ) }
	// }
	// Args:
	// * element: XML start element
	// Returns:
	// * string name of the XML element
	"rye-sxml-start//Name?": {
		Argsn: 1,
		Doc:   "Returns the local name of an XML start or end element.",
		Fn:    sxmlElementName,
	},
	// Example:
	// ; inside an executable [ ] handler
	// `<r><x/></r>` |reader |Do-sxml { <x> { 'end [ .Name? |prns ] } }
	//
	// ; inside a projection ( ) handler
	// `<r><x>A</x></r>` |reader |Do-sxml {
	//   <x> { 'end ( 'put-key , "name" , .Name? ) }
	// }
	// Args:
	// * element: XML end element.
	// Returns:
	// * String local name of the element.
	"rye-sxml-end//Name?": {
		Argsn: 1,
		Doc:   "Returns the local name of an XML end element.",
		Fn:    sxmlElementName,
	},
}

func sxmlElementName(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
	switch obj := arg0.(type) {
	case env.Native:
		switch obj1 := obj.Value.(type) {
		case xml.StartElement:
			return *env.NewString(obj1.Name.Local)
		case xml.EndElement:
			return *env.NewString(obj1.Name.Local)
		default:
			return evaldo.MakeBuiltinError(ps, "Not an XML start or end element.", "rye-sxml-element//Name?")
		}
	default:
		return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "rye-sxml-element//Name?")
	}
}
