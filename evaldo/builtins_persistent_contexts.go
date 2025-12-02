package evaldo

import (
	"fmt"

	"github.com/dgraph-io/badger/v4"
	"github.com/refaktor/rye/env"
	"github.com/refaktor/rye/loader"
)

func (pc PersistentCtx) Type() env.Type {
	return env.PersistentContextType
}

// PersistentCtx wraps a RyeCtx with a BadgerDB backend for persistence.
// It operates in database-first mode for ACID compliance.
type PersistentCtx struct {
	env.RyeCtx // Minimal in-memory context for interface compatibility only
	db         *badger.DB
	Idxs       *env.Idxs
}

// NewPersistentCtx creates a new persistent context or loads an existing one.
func NewPersistentCtx(dbPath string, ps *env.ProgramState) (*PersistentCtx, error) {
	opts := badger.DefaultOptions(dbPath)
	opts.Logger = nil // Disable logging for cleaner output
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}

	ctx := &PersistentCtx{
		RyeCtx: *env.NewEnv(nil), // Empty in-memory context - database is source of truth
		db:     db,
		Idxs:   ps.Idx,
	}

	// No loading to memory - database-first approach for ACID compliance
	return ctx, nil
}

// Set overrides the embedded Set method to add ACID persistence.
func (pc *PersistentCtx) Set(word int, val env.Object) env.Object {
	// Database-first approach: check existence and set atomically within transaction
	var result env.Object

	err := pc.db.Update(func(txn *badger.Txn) error {
		key := []byte(pc.Idxs.GetWord(word))

		// Check if key already exists (ACID consistency)
		_, err := txn.Get(key)
		if err == nil {
			// Key exists - return error
			result = *env.NewError("Can't set already set word, try using modword!")
			return fmt.Errorf("word already exists")
		}
		if err != badger.ErrKeyNotFound {
			// Real error occurred
			result = *env.NewError("Database error: " + err.Error())
			return err
		}

		// Key doesn't exist - safe to set
		value := []byte(val.Dump(*pc.Idxs))
		err = txn.Set(key, value)
		if err != nil {
			result = *env.NewError("Failed to persist value: " + err.Error())
			return err
		}

		result = val
		return nil
	})

	// If transaction failed but we haven't set result yet
	if err != nil && result == nil {
		result = *env.NewError("Transaction failed: " + err.Error())
	}

	return result
}

// Mod overrides the embedded Mod method to add ACID persistence.
func (pc *PersistentCtx) Mod(word int, val env.Object) bool {
	// Database-first approach: modify directly in database atomically
	err := pc.db.Update(func(txn *badger.Txn) error {
		key := []byte(pc.Idxs.GetWord(word))
		varKey := []byte("__var__" + pc.Idxs.GetWord(word))

		// Check if key exists
		_, err := txn.Get(key)
		if err == badger.ErrKeyNotFound {
			// Word doesn't exist, create it as a variable
			err = txn.Set(varKey, []byte("true"))
			if err != nil {
				return err
			}
		} else {
			// Key exists - check if it's a variable
			_, err := txn.Get(varKey)
			if err == badger.ErrKeyNotFound {
				// Not a variable - cannot modify constant
				return fmt.Errorf("cannot modify constant")
			}
		}

		// Set/update the value
		value := []byte(val.Dump(*pc.Idxs))
		return txn.Set(key, value)
	})

	return err == nil
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

// Unset overrides the embedded Unset method to add ACID persistence.
func (pc *PersistentCtx) Unset(word int, idxs *env.Idxs) env.Object {
	var result env.Object

	err := pc.db.Update(func(txn *badger.Txn) error {
		key := []byte(pc.Idxs.GetWord(word))
		varKey := []byte("__var__" + pc.Idxs.GetWord(word))

		// Check if key exists
		_, err := txn.Get(key)
		if err == badger.ErrKeyNotFound {
			result = *env.NewError("Can't unset non-existing word " + idxs.GetWord(word) + " in this context")
			return fmt.Errorf("word not found")
		}

		// Delete both the value and variable flag
		err = txn.Delete(key)
		if err != nil {
			result = *env.NewError("Failed to delete value: " + err.Error())
			return err
		}

		// Also delete variable flag if it exists (ignore errors)
		txn.Delete(varKey)

		result = *env.NewInteger(1)
		return nil
	})

	if err != nil && result == nil {
		result = *env.NewError("Transaction failed: " + err.Error())
	}

	return result
}

// SetNew overrides the embedded SetNew method to add ACID persistence.
func (pc *PersistentCtx) SetNew(word int, val env.Object, idxs *env.Idxs) bool {
	err := pc.db.Update(func(txn *badger.Txn) error {
		key := []byte(pc.Idxs.GetWord(word))

		// Check if key already exists
		_, err := txn.Get(key)
		if err == nil {
			// Key exists - cannot set new
			return fmt.Errorf("word already exists")
		}
		if err != badger.ErrKeyNotFound {
			// Real error occurred
			return err
		}

		// Key doesn't exist - safe to set
		value := []byte(val.Dump(*pc.Idxs))
		return txn.Set(key, value)
	})

	return err == nil
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

// Get overrides the embedded Get method to read from database.
func (pc *PersistentCtx) Get(word int) (env.Object, bool) {
	var result env.Object
	var found bool

	pc.db.View(func(txn *badger.Txn) error {
		key := []byte(pc.Idxs.GetWord(word))
		item, err := txn.Get(key)
		if err != nil {
			found = false
			return nil
		}

		val, err := item.ValueCopy(nil)
		if err != nil {
			found = false
			return nil
		}

		// Create a minimal program state for deserialization
		ps := &env.ProgramState{Idx: pc.Idxs}
		result = loader.LoadStringNEW(string(val), false, ps)
		found = true
		return nil
	})

	// Check parent if not found locally and parent exists
	if !found && pc.RyeCtx.Parent != nil {
		return pc.RyeCtx.Parent.Get(word)
	}

	return result, found
}

// MarkAsVariable marks a word as a variable in the database
func (pc *PersistentCtx) MarkAsVariable(word int) {
	pc.db.Update(func(txn *badger.Txn) error {
		varKey := []byte("__var__" + pc.Idxs.GetWord(word))
		return txn.Set(varKey, []byte("true"))
	})
}

// IsVariable checks if a word is a variable by reading from database
func (pc *PersistentCtx) IsVariable(word int) bool {
	var isVar bool
	pc.db.View(func(txn *badger.Txn) error {
		varKey := []byte("__var__" + pc.Idxs.GetWord(word))
		_, err := txn.Get(varKey)
		isVar = (err == nil)
		return nil
	})
	return isVar
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
			EvalExpression_DispatchType(pce.ps)
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
	EvalExpression_DispatchType(pce.ps)
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
