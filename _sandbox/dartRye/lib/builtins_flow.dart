// builtins_flow.dart - Flow control builtins (do, defer, return, etc.)

import 'env.dart' show ProgramState, RyeCtx; // Import specific env types
// Import specific types needed from types.dart
import 'types.dart' show RyeObject, Block, Error, Void, Builtin, TSeries; // Added TSeries here
import 'evaldo.dart' show evalBlockInj; // Import the correct evaluator function

// Implements the "do" builtin function
RyeObject doBuiltin(ProgramState ps, RyeObject? arg0, RyeObject? arg1, RyeObject? arg2, RyeObject? arg3, RyeObject? arg4) {
  if (arg0 is Block) {
    // Store current series
    TSeries ser = ps.ser;
    
    // Set series to the block's series
    ps.ser = arg0.series;
    
    // Evaluate the block
    evalBlockInj(ps, null, false); // Use evalBlockInj
    
    // Restore original series
    ps.ser = ser;
    
    // Return the result of the evaluation
    return ps.res!;
  }
  ps.failureFlag = true;
  return Error("do expects a block argument");
}

// Implements the "defer" builtin
RyeObject deferBuiltin(ProgramState ps, RyeObject? arg0, RyeObject? arg1, RyeObject? arg2, RyeObject? arg3, RyeObject? arg4) {
  if (arg0 is Block) {
    // Add the block to the list of deferred blocks for the current scope
    ps.deferBlocks.add(arg0);
    return Void(); // Defer itself returns nothing
  } else {
    ps.failureFlag = true;
    return Error("defer expects a block argument");
  }
}

// Implements the "do\in" builtin
RyeObject doInBuiltin(ProgramState ps, RyeObject? arg0, RyeObject? arg1, RyeObject? arg2, RyeObject? arg3, RyeObject? arg4) {
  if (arg0 is RyeCtx && arg1 is Block) {
    // Store current series and context
    TSeries ser = ps.ser;
    RyeCtx currentCtx = ps.ctx;
    
    // Set series to the block's series and context to the provided context
    ps.ser = arg1.series;
    
    // We need to use the provided context directly
    RyeCtx providedCtx = arg0 as RyeCtx; // Explicit cast to satisfy analyzer
    ps.ctx = providedCtx;
    
    // Evaluate the block in the provided context
    evalBlockInj(ps, null, false); // Use evalBlockInj
    
    // Restore original series and context
    ps.ser = ser;
    ps.ctx = currentCtx;
    
    // Return the result of the evaluation
    return ps.res!;
  } else {
    ps.errorFlag = true;
    if (arg0 is! RyeCtx) {
      return Error("do\\in expects a context as its first argument");
    } else {
      return Error("do\\in expects a block as its second argument");
    }
  }
}

// Implements the "do\par" builtin
RyeObject doParBuiltin(ProgramState ps, RyeObject? arg0, RyeObject? arg1, RyeObject? arg2, RyeObject? arg3, RyeObject? arg4) {
  if (arg0 is RyeCtx && arg1 is Block) {
    // Store current series and parent context
    TSeries ser = ps.ser;
    RyeCtx? originalParent = ps.ctx.parent;
    
    // Set the provided context as the parent of the current context
    // We need to use the provided context directly
    RyeCtx providedCtx = arg0 as RyeCtx; // Explicit cast to satisfy analyzer
    ps.ctx.parent = providedCtx;
    
    // Set series to the block's series
    ps.ser = arg1.series;
    
    // Evaluate the block with the modified parent context
    evalBlockInj(ps, null, false); // Use evalBlockInj
    
    // Restore original series and parent context
    ps.ser = ser;
    ps.ctx.parent = originalParent;
    
    // Return the result of the evaluation
    return ps.res!;
  } else {
    ps.errorFlag = true;
    if (arg0 is! RyeCtx) {
      return Error("do\\par expects a context as its first argument");
    } else {
      return Error("do\\par expects a block as its second argument");
    }
  }
}

// TODO: Implement 'return' builtin if needed (might just use ps.returnFlag)

// --- Registration ---

void registerFlowBuiltins(ProgramState ps) {
  // Register do builtin
  int doIdx = ps.idx.indexWord("do");
  ps.ctx.set(doIdx,
    Builtin(doBuiltin, 1, false, true, "Takes a block of code and does (runs) it"));
    
  // Register defer builtin
  int deferIdx = ps.idx.indexWord("defer");
  ps.ctx.set(deferIdx, 
    Builtin(deferBuiltin, 1, false, false, "Defers execution of a block until the current scope exits"));
    
  // Register do\in builtin
  int doInIdx = ps.idx.indexWord("do\\in");
  ps.ctx.set(doInIdx,
    Builtin(doInBuiltin, 2, false, false, "Takes a Context and a Block. It Does a block inside a given Context."));
    
  // Register do\par builtin
  int doParIdx = ps.idx.indexWord("do\\par");
  ps.ctx.set(doParIdx,
    Builtin(doParBuiltin, 2, false, false, "Takes a Context and a Block. It Does a block in current context but with parent a given Context."));
    
  // TODO: Register 'return' if implemented as a builtin
}
