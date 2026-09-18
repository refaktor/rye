package loader

import (
	"testing"
)

// Sample Rye code for benchmarking
const benchmarkCode = `{
	; This is a comment
	word                ; A simple word
	.op-word            ; An op-word
	|pipe-word          ; A pipe-word
	|one-char-pipe      ; A pipe-word with one char
	123                 ; An integer
	-456                ; A negative integer
	3.14                ; A decimal
	-2.718              ; A negative decimal
	"string"            ; A string
	'tag-word           ; A tag-word
	?get-word           ; A get-word
	set-word:           ; A set-word
	:lset-word          ; A left set-word
	mod-word::          ; A mod-word
	::lmod-word         ; A left mod-word
	<xword>             ; An x-word (now op-word)
	</exword>           ; An ex-word
	http://example.com  ; A URI
	user@example.com    ; An email
	%file/path          ; A file path
	context/path        ; A context path
	.op/context/path    ; An op context path
	|pipe/context/path  ; A pipe context path
	{ nested block }    ; A nested block
	[ bracket block ]   ; A bracket block
	( group )           ; A group
	,                   ; A comma
	_                   ; Void
}`

// BenchmarkNoPEGParser benchmarks the no_peg parser
func BenchmarkNoPEGParser(b *testing.B) {
	for i := 0; i < b.N; i++ {
		LoadStringNoPEG(benchmarkCode, false)
	}
}

// BenchmarkStateMachineParser is only available with the stm_loader build tag.
// To benchmark the state machine parser, build with: go test -tags stm_loader -bench BenchmarkStateMachineParser
