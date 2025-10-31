package evaldo

import (
	"fmt"
	"strconv"

	"github.com/dgraph-io/badger/v4"
	"github.com/refaktor/rye/env"
	"github.com/refaktor/rye/loader"
)

func (pc PersistentCtx) Type() env.Type {
	return env.PersistentContextType
}

// PersistentCtx wraps a RyeCtx with a BadgerDB backend for persistence.
type PersistentCtx struct {
	env.RyeCtx
	db   *badger.DB
	Idxs *env.Idxs
}

// NewPersistentCtx creates a new persistent context or loads an existing one.
func NewPersistentCtx(dbPath string, ps *env.ProgramState) (*PersistentCtx, error) {
	opts := badger.DefaultOptions(dbPath)
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}

	ctx := &PersistentCtx{
		RyeCtx: *env.NewEnv(ps.Ctx),
		db:     db,
		Idxs:   ps.Idx,
	}

	err = ctx.load(ps)
	if err != nil {
		return nil, err
	}

	return ctx, nil
}

// load reads all key-value pairs from the database and populates the in-memory context.
func (pc *PersistentCtx) load(ps *env.ProgramState) error {
	return pc.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		defer it.Close()

		fmt.Println("LOADING**")

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			key := item.Key()
			val, err := item.ValueCopy(nil)
			if err != nil {
				return err
			}

			// Parse the key to extract the word name
			keyParts := string(key)
			if keyParts == "" {
				continue
			}

			// The key format should be "wordname:index" or just the word name
			var wordName string
			if idx := string(key); len(idx) > 0 {
				// Try to convert to integer first (old format)
				if oldWordIndex, err := strconv.Atoi(idx); err == nil {
					// Check if this word index is valid in current session
					if oldWordIndex >= 0 && oldWordIndex < ps.Idx.GetWordCount() {
						wordName = ps.Idx.GetWord(oldWordIndex)
					} else {
						continue // Skip invalid indices
					}
				} else {
					// Assume it's a word name directly
					wordName = idx
				}
			}

			if wordName == "" {
				continue
			}

			// Get or create the word index in the current session
			wordIndex := ps.Idx.IndexWord(wordName)

			// Use the loader to deserialize the value.
			loadedVal := loader.LoadStringNEW(string(val), false, ps)
			pc.RyeCtx.Set(wordIndex, loadedVal)
		}
		return nil
	})
}

// Set overrides the embedded Set method to add persistence.
func (pc *PersistentCtx) Set(word int, val env.Object) env.Object {
	// Set the value in the in-memory context first.
	res := pc.RyeCtx.Set(word, val)

	// Then, persist the change to the database.
	err := pc.db.Update(func(txn *badger.Txn) error {
		key := []byte(pc.Idxs.GetWord(word)) // Use word name instead of index
		value := []byte(val.Dump(*pc.Idxs))
		return txn.Set(key, value)
	})

	if err != nil {
		// Handle the error appropriately.
		fmt.Println("Error persisting value:", err)
	}

	return res
}

// Mod overrides the embedded Mod method to add persistence.
func (pc *PersistentCtx) Mod(word int, val env.Object) bool {
	// First try to modify in the in-memory context
	ok := pc.RyeCtx.Mod(word, val)
	if !ok {
		return false
	}

	// Then, persist the change to the database.
	err := pc.db.Update(func(txn *badger.Txn) error {
		key := []byte(pc.Idxs.GetWord(word)) // Use word name instead of index
		value := []byte(val.Dump(*pc.Idxs))
		return txn.Set(key, value)
	})

	if err != nil {
		// Handle the error appropriately.
		fmt.Println("Error persisting value:", err)
		return false
	}

	return true
}

// Copy creates a copy of the persistent context (note: copy won't be persistent)
func (pc *PersistentCtx) Copy() env.Context {
	// Create a copy of the embedded RyeCtx
	ryeCtxCopy := pc.RyeCtx.Copy()
	if ryeCtx, ok := ryeCtxCopy.(*env.RyeCtx); ok {
		return ryeCtx
	}
	// Fallback - this shouldn't happen
	return env.NewEnv(nil)
}

// GetParent returns the parent context
func (pc *PersistentCtx) GetParent() env.Context {
	return pc.RyeCtx.GetParent()
}

// SetParent sets the parent context
func (pc *PersistentCtx) SetParent(parent env.Context) {
	pc.RyeCtx.SetParent(parent)
}

// Unset overrides the embedded Unset method to add persistence.
func (pc *PersistentCtx) Unset(word int, idxs *env.Idxs) env.Object {
	// First unset in the in-memory context
	res := pc.RyeCtx.Unset(word, idxs)
	if res.Type() == env.ErrorType {
		return res
	}

	// Then, remove from the database.
	err := pc.db.Update(func(txn *badger.Txn) error {
		key := []byte(pc.Idxs.GetWord(word)) // Use word name instead of index
		return txn.Delete(key)
	})

	if err != nil {
		// Handle the error appropriately.
		fmt.Println("Error removing persisted value:", err)
	}

	return res
}

// SetNew overrides the embedded SetNew method to add persistence.
func (pc *PersistentCtx) SetNew(word int, val env.Object, idxs *env.Idxs) bool {
	// First set in the in-memory context
	ok := pc.RyeCtx.SetNew(word, val, idxs)
	if !ok {
		return false
	}

	// Then, persist the change to the database.
	err := pc.db.Update(func(txn *badger.Txn) error {
		key := []byte(pc.Idxs.GetWord(word)) // Use word name instead of index
		value := []byte(val.Dump(*pc.Idxs))
		return txn.Set(key, value)
	})

	if err != nil {
		// Handle the error appropriately.
		fmt.Println("Error persisting new value:", err)
		return false
	}

	return true
}

// GetDoc returns the documentation string
func (pc *PersistentCtx) GetDoc() string {
	return pc.RyeCtx.GetDoc()
}

// SetDoc sets the documentation string
func (pc *PersistentCtx) SetDoc(doc string) {
	pc.RyeCtx.SetDoc(doc)
}

// GetKindWord returns the Kind as a Word
func (pc *PersistentCtx) GetKindWord() env.Word {
	return pc.RyeCtx.GetKindWord()
}

// SetKindWord sets the Kind
func (pc *PersistentCtx) SetKindWord(kind env.Word) {
	pc.RyeCtx.SetKindWord(kind)
}

// AsRyeCtx returns the underlying RyeCtx for backward compatibility
func (pc *PersistentCtx) AsRyeCtx() *env.RyeCtx {
	return &pc.RyeCtx
}

// CreatePersistentRyeCtx creates a RyeCtx that has persistent behavior by injecting persistence functions
func CreatePersistentRyeCtx(pctx *PersistentCtx) *PersistentRyeCtx {
	return &PersistentRyeCtx{
		RyeCtx:        pctx.RyeCtx, // Copy the underlying RyeCtx
		persistentCtx: pctx,
	}
}

// PersistentRyeCtx is a RyeCtx with persistence methods injected
type PersistentRyeCtx struct {
	env.RyeCtx
	persistentCtx *PersistentCtx
}

// Override Set to use PersistentCtx persistence
func (w *PersistentRyeCtx) Set(word int, val env.Object) env.Object {
	return w.persistentCtx.Set(word, val)
}

// Override Mod to use PersistentCtx persistence
func (w *PersistentRyeCtx) Mod(word int, val env.Object) bool {
	return w.persistentCtx.Mod(word, val)
}

// Override Unset to use PersistentCtx persistence
func (w *PersistentRyeCtx) Unset(word int, idxs *env.Idxs) env.Object {
	return w.persistentCtx.Unset(word, idxs)
}

// Override SetNew to use PersistentCtx persistence
func (w *PersistentRyeCtx) SetNew(word int, val env.Object, idxs *env.Idxs) bool {
	return w.persistentCtx.SetNew(word, val, idxs)
}

// EvalBlockInPersistentCtx evaluates a block within a PersistentCtx, ensuring persistence
func EvalBlockInPersistentCtx(ps *env.ProgramState, pctx *PersistentCtx) {
	// Save the original context
	originalCtx := ps.Ctx

	// Temporarily replace the context with the persistent context's RyeCtx
	// but override the methods that need persistence
	ps.Ctx = &pctx.RyeCtx

	// Create a special program state interceptor that routes persistence operations
	// through the PersistentCtx instead of the raw RyeCtx
	interceptor := &PersistentCtxEvaluator{
		ps:               ps,
		pctx:             pctx,
		originalEvalStep: ps, // Store original for delegation
	}

	// Execute the block with persistence intercepts
	interceptor.EvalBlock()

	// Restore the original context
	ps.Ctx = originalCtx
}

// PersistentCtxEvaluator intercepts evaluation to ensure persistence operations go through PersistentCtx
type PersistentCtxEvaluator struct {
	ps               *env.ProgramState
	pctx             *PersistentCtx
	originalEvalStep *env.ProgramState
}

// EvalBlock evaluates the current series while intercepting persistence operations
func (pce *PersistentCtxEvaluator) EvalBlock() {
	// Use a custom evaluation loop that handles setwords and modwords specially
	for pce.ps.Ser.Pos() < pce.ps.Ser.Len() && !pce.ps.ErrorFlag && !pce.ps.ReturnFlag && !pce.ps.FailureFlag {
		obj := pce.ps.Ser.Pop()

		// Handle setwords and modwords specially to ensure persistence
		switch val := obj.(type) {
		case env.Setword:
			// For setwords, we need to use PersistentCtx.SetNew instead of RyeCtx.SetNew
			pce.handleSetword(val)
		case env.Modword:
			// For modwords, we need to use PersistentCtx.Mod instead of RyeCtx.Mod
			pce.handleModword(val)
		default:
			// For all other operations, put the object back and use normal evaluation
			pce.ps.Ser.SetPos(pce.ps.Ser.Pos() - 1) // Put object back
			EvalExpressionConcrete(pce.ps)
		}

		if pce.ps.SkipFlag {
			pce.ps.SkipFlag = false
			break
		}
	}
}

// handleSetword processes setwords to ensure persistence
func (pce *PersistentCtxEvaluator) handleSetword(sw env.Setword) {
	// First evaluate the next expression to get the value to set
	EvalExpressionConcrete(pce.ps)
	if pce.ps.ErrorFlag || pce.ps.FailureFlag {
		return
	}

	val := pce.ps.Res

	// Use PersistentCtx's SetNew method for persistence
	success := pce.pctx.SetNew(sw.Index, val, pce.ps.Idx)
	if !success {
		pce.ps.FailureFlag = true
		pce.ps.Res = env.NewError("Can't set already set word " + pce.ps.Idx.GetWord(sw.Index))
	} else {
		pce.ps.Res = val
	}
}

// handleModword processes modwords to ensure persistence
func (pce *PersistentCtxEvaluator) handleModword(mw env.Modword) {
	// Get the value to set (should be the current result)
	val := pce.ps.Res

	// Use PersistentCtx's Mod method for persistence
	success := pce.pctx.Mod(mw.Index, val)
	if !success {
		pce.ps.FailureFlag = true
		pce.ps.Res = env.NewError("Cannot modify constant '" + pce.ps.Idx.GetWord(mw.Index) + "', use 'var' to declare it as a variable")
	} else {
		pce.ps.Res = val
	}
}

var builtins_persistent_contexts = map[string]*env.Builtin{
	"persistent-context": {
		Argsn: 1,
		Doc:   "Opens or creates a persistent context at the given path.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch path := arg0.(type) {
			case env.String:
				pctx, err := NewPersistentCtx(path.Value, ps)
				if err != nil {
					return env.NewError("Failed to open persistent context: " + err.Error())
				}
				return *pctx
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "persistent-context")
			}
		},
	},
}
