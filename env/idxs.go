package env

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
)

// Idxs is a bidirectional mapping between words (strings) and their integer indices.
// This structure is thread-safe and can be shared across multiple ProgramState instances
// for multi-session scenarios (e.g., HTTP REPL with sessions).
type Idxs struct {
	words1 []string
	words2 map[string]int
	mu     sync.RWMutex // Protects concurrent access for multi-session use
}

var NativeTypes = [...]string{ // Todo change to BuiltinTypes
	"Block",
	"Integer",
	"Word",
	"Setword",
	"Opword",
	"Pipeword",
	"Builtin",
	"Function",
	"Error",
	"Comma",
	"Void",
	"String",
	"Tagword",
	"Genword",
	"Getword",
	"Argword",
	"Native",
	"Uri",
	"LSetword",
	"Context",
	"Dict",
	"List",
	"Date",
	"CPath",
	"Xword",
	"EXword",
	"Table",
	"Email",
	"Kind",
	"Kindword",
	"Converter",
	"Time",
	"TableRowType",
	"Decimal",
	"Vector",
	"OpCPath",
	"PipeCPath",
	"Modword",
	"LModword",
	"Boolean",
	"VarBuiltin",
	"CurriedCaller",
	"Complex",
	"Markdown",
	"PersistentCtx",
	"LocationNode",
	"Flagword",
	"Secret",
	"LazyValue",
	"PersistentTable",
}

// IndexWord returns the index of a word, creating a new index if it doesn't exist.
// This method is thread-safe for concurrent use across multiple sessions.
func (e *Idxs) IndexWord(w string) int {
	// First try with read lock (fast path for existing words)
	e.mu.RLock()
	idx, ok := e.words2[w]
	e.mu.RUnlock()
	if ok {
		return idx
	}

	// Word not found, need write lock to add it
	e.mu.Lock()
	defer e.mu.Unlock()

	// Double-check after acquiring write lock (another goroutine might have added it)
	idx, ok = e.words2[w]
	if ok {
		return idx
	}

	e.words1 = append(e.words1, w)
	e.words2[w] = len(e.words1) - 1
	return len(e.words1) - 1
}

// GetIndex returns the index of a word if it exists.
// This method is thread-safe for concurrent use across multiple sessions.
func (e *Idxs) GetIndex(w string) (int, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	idx, ok := e.words2[w]
	if ok {
		return idx, true
	}
	return 0, false
}

// GetWord returns the word at the given index.
// This method is thread-safe for concurrent use across multiple sessions.
func (e *Idxs) GetWord(i int) string {
	if i < 0 {
		return "isolate!" // TODO -- behaviour aroung isolates ... so that it prevents generic function lookup
	}
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.words1[i]
}

// Print outputs all words in the index.
func (e *Idxs) Print() {
	e.mu.RLock()
	defer e.mu.RUnlock()
	fmt.Print("<IDXS: ")
	for i := range e.words1 {
		fmt.Print(strconv.FormatInt(int64(i), 10) + ": " + e.words1[i] + " ")
	}
	fmt.Println(">")
}

// GetWordCount returns the number of words in the index.
// This method is thread-safe for concurrent use across multiple sessions.
func (e *Idxs) GetWordCount() int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return len(e.words1)
}

func NewIdxs() *Idxs {
	var e Idxs
	e.words1 = []string{""}
	e.words2 = make(map[string]int)

	/*
		BlockType    Type = 1
		IntegerType  Type = 2
		WordType     Type = 3
		SetwordType  Type = 4
		OpwordType   Type = 5
		PipewordType Type = 6
		BuiltinType  Type = 7
		FunctionType Type = 8
		ErrorType    Type = 9
		CommaType    Type = 10
		VoidType     Type = 11
		StringType   Type = 12
		TagwordType  Type = 13
	*/

	// register words for builtin kinds, which the value objects should return on GetKind()

	for _, value := range NativeTypes {
		e.IndexWord(strings.ToLower(value))
	}
	return &e
}
