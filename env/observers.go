package env

// TriggerObserversInChain triggers all observers for a word change by searching up the context chain
// This function searches for observers in all contexts in the chain (starting from current context going up to parents)
func TriggerObserversInChain(ps *ProgramState, ctx *RyeCtx, wordIndex int, oldValue, newValue Object) {
	if ctx == nil {
		return
	}

	// Search up the context chain for observers
	currentCtx := ctx
	for currentCtx != nil {
		if currentCtx.HasObservers(wordIndex) {
			observers := currentCtx.GetObservers(wordIndex)
			// Execute observers in this context where they were registered
			for _, observer := range observers {
				executeObserver(ps, currentCtx, observer, oldValue, newValue)
			}
		}
		currentCtx = currentCtx.Parent
	}
}

// executeObserver runs a single observer block with proper context setup
func executeObserver(ps *ProgramState, observerCtx *RyeCtx, observer Block, oldValue, newValue Object) {
	// Save current evaluation state
	originalSer := ps.Ser
	originalRes := ps.Res
	originalCtx := ps.Ctx
	originalFailureFlag := ps.FailureFlag
	originalErrorFlag := ps.ErrorFlag

	// Set up for observer execution with the context where the observer was registered
	ps.Ser = observer.Series
	ps.Ctx = observerCtx // Execute in the context where the observer was registered
	ps.FailureFlag = false
	ps.ErrorFlag = false

	// Execute the observer block with old value injected
	// This makes the old value available as the injected value in the observer block
	// We need to use a callback to avoid circular imports
	if ObserverExecutor != nil {
		ObserverExecutor(ps, oldValue, true)
	}

	// Handle any errors in observer execution
	if ps.ErrorFlag || ps.FailureFlag {
		// Log observer error but don't propagate it to main execution
		// Could add more sophisticated error handling here
	}

	// Restore original state
	ps.Ser = originalSer
	ps.Res = originalRes
	ps.Ctx = originalCtx
	ps.FailureFlag = originalFailureFlag
	ps.ErrorFlag = originalErrorFlag
}

// ObserverExecutor is a callback function to execute observer blocks
// This avoids circular imports between env and evaldo packages
var ObserverExecutor func(*ProgramState, Object, bool)
