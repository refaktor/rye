// evaldo.dart - Core evaluator logic for Dart Rye (Mirroring Go's evaldo.go)

import 'dart:io'; // For display functions

// Ensure Comma is imported from types.dart
import 'types.dart' show RyeObject, RyeType, Word, Error, Void, TSeries, Block, Builtin, RyeFunction, Tagword, Comma; 
import 'env.dart'; // Import env for ProgramState, RyeCtx
import 'idxs.dart'; // Import idxs for Idxs
// Remove Comma from loader import if it was there, it's in types.dart
import 'loader.dart' show Setword, OpWord, PipeWord, LSetword, LModword, CPath, LitWord, Getword, Genword, Modword; 
// builtins.dart is likely not needed directly here anymore, callBuiltin/callFunction are defined below
// import 'builtins.dart'; 

// --- Core Evaluation Functions ---

// Entry point for evaluating a block
void evalBlock(ProgramState ps) {
  evalBlockInj(ps, null, false);
}

// Evaluates a block with optional injection
void evalBlockInj(ProgramState ps, RyeObject? inj, bool injnow) {
  // Repeats until at the end of the block
  while (ps.ser.pos < ps.ser.len()) {
    injnow = evalExpression(ps, inj, injnow, false); // Evaluate expression, not limited

    // Handle potential failure after expression evaluation
    if (tryHandleFailure(ps)) {
      // If tryHandleFailure returns true, it means the failure was unhandled and promoted to error
      return; // Propagate error
    }

    // If return flag was raised or an error occurred (either originally or from handler)
    if (ps.returnFlag || ps.errorFlag) {
      // Execute deferred blocks *before* returning from the block evaluation
      if (ps.deferBlocks.isNotEmpty) {
         executeDeferredBlocks(ps);
      }
      return; // Propagate return/error
    }

    // Check for comma and potentially reset injection flag for the *next* expression
    injnow = maybeAcceptComma(ps, inj, injnow); 
    // If comma was accepted and there was an initial injection, 
    // the *original* injected value 'inj' might be used again if 'injnow' becomes true.
    // However, the current design uses the *result* of the previous expression for op/pipe words.
    // Resetting injnow based on comma seems correct for block-level injection.
  }
}

// Handles optional comma separator between expressions in a block.
// Returns true if the next injected value should be used (injnow).
bool maybeAcceptComma(ProgramState ps, RyeObject? inj, bool injnow) {
  RyeObject? obj = ps.ser.peek();
  if (obj is Comma) {
    ps.ser.next(); // Consume the comma
    if (inj != null) {
      // If there was an initial injection, make it available again after comma
      return true; 
    }
  }
  // If no comma, or comma but no initial injection, the injection state for the *next*
  // expression depends on whether the *current* expression used an injection.
  // The evalExpression function returns the updated 'injnow' state reflecting this.
  return injnow; 
}


// Consolidated evaluation function (mirrors Go's EvalExpression)
// Returns the updated 'injnow' status for the *next* expression in the block.
bool evalExpression(ProgramState ps, RyeObject? inj, bool injnow, bool limited) {
  RyeObject? originalRes = ps.res; // Store result before potential opword evaluation

  if (inj == null || !injnow) {
    evalExpressionConcrete(ps, limited: limited); // Pass limited flag
    if (ps.returnFlag || ps.errorFlag) { // Check flags *after* concrete evaluation
      return false; // Injection not used for the next expression if error/return
    }
  } else {
    ps.res = inj;
    injnow = false; // Injection consumed for this expression
    if (ps.returnFlag) { // Check return flag after setting injected value
      return false; 
    }
  }

  // Store the result *before* checking for opwords on the right
  RyeObject? valueBeforeOpword = ps.res; 

  // Check for opwords/pipewords to the right
  maybeEvalOpwordOnRight(ps, limited: limited); 

  // If maybeEvalOpwordOnRight caused an error or return, propagate that state
  if (ps.returnFlag || ps.errorFlag) {
     return false; 
  }

  // Restore original result if maybeEvalOpwordOnRight modified ps.res unexpectedly?
  // Go version doesn't seem to do this explicitly, relies on call stack.
  // Let's assume ps.res holds the final result after opword chain.

  // The 'injnow' for the *next* expression is false because the injection (if any) was used.
  return false; 
}


// Evaluates a concrete Rye object (mirrors Go's EvalExpressionConcrete)
void evalExpressionConcrete(ProgramState ps, {bool limited = false}) { // Added limited, though might not be used directly here
  RyeObject? object = ps.ser.pop();
  if (object == null) {
    ps.setError00("Expected Rye value but got to the end of the block"); // Use ps.setError00
    return;
  }

  // Handle simple literal types directly
  RyeType objType = object.type();
  if (objType == RyeType.integerType ||
      objType == RyeType.stringType || // Assuming String exists
      objType == RyeType.decimalType || // Assuming Decimal exists
      objType == RyeType.voidType ||
      objType == RyeType.uriType ||    // Assuming Uri exists
      objType == RyeType.emailType ||  // Assuming Email exists
      objType == RyeType.booleanType || // Assuming Boolean exists
      objType == RyeType.dateType ||    // Assuming Date exists
      objType == RyeType.timeType ||    // Assuming Time exists
      objType == RyeType.listType ||    // Assuming List exists and evaluates to self
      objType == RyeType.dictType) {    // Assuming Dict exists and evaluates to self
    ps.res = object;
    return;
  }

  // Handle more complex types
  switch (object.type()) {
    case RyeType.blockType:
      Block block = object as Block;
      if (block.mode == 0) { // Mode 0 []: Literal block
        ps.res = object;
      } else if (block.mode == 1) { // Mode 1 []: Eval items, return new block
        TSeries ser = ps.ser; // Store current series
        ps.ser = block.series;
        ps.ser.reset();
        List<RyeObject?> results = [];
        while (ps.ser.pos < ps.ser.len()) {
          // Evaluate expression within the block, NOT limited
          evalExpression(ps, null, false, false); 
          if (ps.errorFlag || ps.returnFlag) {
            ps.ser = ser; // Restore series on error/return
            return; 
          }
          results.add(ps.res);
        }
        ps.ser = ser; // Restore original series
        ps.res = Block(TSeries(results)); // Return new block with results
      } else if (block.mode == 2) { // Mode 2 {}: Evaluate block in place
         TSeries ser = ps.ser; // Store current series
         ps.ser = block.series;
         ps.ser.reset();
         evalBlockInj(ps, null, false); // Evaluate block content, no injection
         ps.ser = ser; // Restore original series
         // ps.res holds the result of the block evaluation
      } else {
         ps.setError00("Unsupported block mode: ${block.mode}"); // Use ps.setError00
      }
      break;
    case RyeType.wordType: // This now covers Word, OpWord, PipeWord, CPath, etc.
       if (object is LitWord) { // Handle 'LitWord specifically
         ps.res = Word(object.index); // Evaluate to the underlying Word
       } else if (object is Tagword) { // Handle 'Tagword
         ps.res = Word(object.index); // Evaluate to the underlying Word
       } else if (object is Getword) {
         evalGetword(ps, object);
       } else if (object is Genword) {
         evalGenword(ps, object);
       } else if (object is Setword) { // Should have been handled earlier, but as fallback
         evalSetword(ps, object);
       } else if (object is Modword) {
         evalModword(ps, object);
       } else { // Includes Word, OpWord, PipeWord, CPath, LSetword, LModword
         // Normal word evaluation (lookup and potentially execute)
         evalWord(ps, object, null, false, false, null); // No leftVal initially
       }
       break;
    case RyeType.builtinType:
      callBuiltin(object as Builtin, ps, null, false, false, null); // No leftVal initially
      break;
    // VarBuiltinType might need its own case if different calling convention
    case RyeType.errorType:
      ps.setError00("Error object encountered directly in code block", object); // Use ps.setError00
      break;
    case RyeType.commaType: // Use the enum member
       ps.setError00("Expression guard (comma) found inside expression"); // Use ps.setError00
       break;
    default:
      // Assuming unknown types evaluate to themselves, or throw error? Go throws error.
      // stdout.writeln(object.inspect(ps.idx)); // Use inspect for debugging
      ps.setError00("Unknown or unevaluatable Rye type in block: ${object.runtimeType}"); // Use ps.setError00
      break;
  }
}

// Finds the value associated with a word/cpath (mirrors Go's findWordValue)
(bool, RyeObject?, RyeCtx?) findWordValue(ProgramState ps, RyeObject word1) {
  if (word1 is Word) { // Base case: simple word
    // Check current context and parents
    var (obj, found, ctx) = ps.ctx.get2(word1.index); // Destructure correctly
    return (found, obj, ctx); // Return in expected order
  } else if (word1 is CPath) {
    // Context Path logic
    RyeCtx? currentCtx = ps.ctx;
    RyeObject? object;
    bool found = false;
    RyeCtx? foundCtx = ps.ctx; // Track the context where the final part is found

    for (int i = 0; i < word1.path.length; i++) {
      Word segment = word1.path[i];
      var (obj, foundSeg, ctxSeg) = currentCtx!.get2(segment.index); // Use get2 to search parents, rename vars

      if (!foundSeg) { // Use renamed var
        return (false, null, null); // Segment not found
      }

      if (i == word1.path.length - 1) {
        // Last segment, this is our object
        object = obj;
        found = true;
        foundCtx = ctxSeg; // Record the context where the final segment was found
      } else {
        // Not the last segment, must resolve to a context
        if (obj is RyeCtx) { // Check if obj is a RyeCtx
          currentCtx = obj as RyeCtx; // Explicit cast to RyeCtx
        } else {
          // Intermediate path segment did not resolve to a context
          return (false, null, null); 
        }
      }
    }
    return (found, object, foundCtx);
  }
  // Should not happen for valid word types
  return (false, null, null); 
}


// Evaluates a word (Word, OpWord, PipeWord, CPath) (mirrors Go's EvalWord)
void evalWord(ProgramState ps, RyeObject word, RyeObject? leftVal, bool toLeft, bool pipeSecond, RyeObject? firstVal) {
  int originalPos = ps.ser.getPos(); // Remember position in case of failure/generic lookup

  // 1. Try normal context lookup first
  var (found, object, foundCtx) = findWordValue(ps, word); // Already corrected return order

  // 2. If not found, try generic word lookup (if applicable)
  if (!found) {
    RyeObject? dispatchVal = null;
    int kind = -1;

    if (pipeSecond && firstVal != null) {
      dispatchVal = firstVal;
      kind = dispatchVal.getKind();
    } else if (leftVal != null) {
      dispatchVal = leftVal;
      kind = dispatchVal.getKind();
    } else if (!pipeSecond && !toLeft) {
      // If it's a plain word with no preceding value, evaluate the *next* expression
      // to see if it can be used for generic dispatch.
      if (!ps.ser.atLast()) {
        int peekPos = ps.ser.getPos();
        evalExpression(ps, null, false, true); // Evaluate next expression, limited
        if (!ps.errorFlag && !ps.returnFlag && !ps.failureFlag) {
          dispatchVal = ps.res;
          kind = dispatchVal?.getKind() ?? -1;
          leftVal = dispatchVal; // Use this evaluated value as the effective leftVal
        } else {
          // If evaluation failed, reset position and ignore potential generic dispatch
          ps.ser.setPos(peekPos); 
          // Clear flags? Or let the outer loop handle them? Go seems to let outer loop handle.
        }
      }
    }

    // Perform generic lookup if we have a dispatch value and a plain Word index
    if (dispatchVal != null && kind != -1 && word is Word && word.index > 0) { // Ensure it's a valid word index
       // Check if context kind allows generic lookup (ps.Ctx.Kind.Index != -1 in Go)
       // Assuming kind index 0 is the default/global context where generics might not apply?
       // Let's assume generics apply unless explicitly disabled.
       bool allowGeneric = true; // Simplified for now
       if (allowGeneric) {
          var (genObj, genFound) = ps.getGeneric(kind, word.index);
          if (genFound) {
             // Generic word found, evaluate it
             evalObject(ps, genObj!, dispatchVal, true, null, false, null); // toLeft=true for generic
             return; // Exit after successful generic dispatch
          }
       }
    }
  }

  // 3. Evaluate the found object (or error if still not found)
  if (found) {
    evalObject(ps, object!, leftVal, toLeft, foundCtx, pipeSecond, firstVal);
  } else {
    // Word not found in normal or generic contexts
    ps.ser.setPos(originalPos); // Reset position to before word lookup
    ps.setError00("Word not found: ${word.print(ps.idx)}"); // Use ps.setError00
  }
}

// Evaluates a Genword (explicit generic word) (mirrors Go's EvalGenword)
void evalGenword(ProgramState ps, Genword word) {
  // Evaluate the expression to the right to determine the type for dispatch
  evalExpression(ps, null, false, true); // Evaluate argument, limited
  if (ps.errorFlag || ps.returnFlag || ps.failureFlag) {
     // Propagate error/return/failure
     if (!ps.errorFlag && !ps.returnFlag) ps.setError00("Failure evaluating generic word argument", ps.res); // Use ps.setError00
     return;
  }
  
  RyeObject? arg0 = ps.res;
  int kind = arg0?.getKind() ?? -1; // Use Void kind if null? Or error? Let's use -1.

  if (kind != -1) {
     var (genericObj, genericFound) = ps.getGeneric(kind, word.index);
     if (genericFound) {
         // Generic word found, evaluate it with the argument as leftVal
         evalObject(ps, genericObj!, arg0, true, null, false, null); // toLeft=true
     } else {
        ps.setError00("Generic word not found for type ${arg0?.runtimeType}: ${word.print(ps.idx)}"); // Use ps.setError00
     }
  } else {
     ps.setError00("Invalid type for generic word dispatch: ${arg0?.runtimeType}"); // Use ps.setError00
  }
}

// Evaluates a Getword (e.g., :name) (mirrors Go's EvalGetword)
void evalGetword(ProgramState ps, Getword word) {
  var (object, found, _) = ps.ctx.get2(word.index); // Search current and parent contexts
  if (found) {
    ps.res = object; // Return the found object without evaluation
  } else {
    ps.setError00("Word not found for get-word: ${word.print(ps.idx)}"); // Use ps.setError00
  }
}

// Evaluates a Setword (e.g., name:) (mirrors Go's EvalSetword)
// Note: Also handled specially in evalExpressionConcrete
void evalSetword(ProgramState ps, Setword word) {
  // Evaluate the expression to the right
  evalExpression(ps, null, false, false); // Not limited
  if (ps.errorFlag || ps.returnFlag) return;
  if (ps.failureFlag) { ps.setError00("Failure evaluating value for set-word", ps.res); return; } // Use ps.setError00

  int idx = word.index;
  // Go version checks ps.AllowMod, but Setword seems to bypass it, using SetNew directly.
  bool ok = ps.ctx.setNew(idx, ps.res!); // Use setNew - defines in current context only if new
  if (!ok) {
    ps.setError00("Can't set already set word '${ps.idx.getWord(idx)}', try using '::'"); // Use ps.setError00
    ps.failureFlag = true; // Match Go behavior
  }
  // ps.res already holds the value
}

// Evaluates a Modword (e.g., name::) (mirrors Go's EvalModword)
void evalModword(ProgramState ps, Modword word) {
  // Evaluate the expression to the right
  evalExpression(ps, null, false, false); // Not limited
  if (ps.errorFlag || ps.returnFlag) return;
  if (ps.failureFlag) { ps.setError00("Failure evaluating value for mod-word", ps.res); return; } // Use ps.setError00

  int idx = word.index;
  bool ok = ps.ctx.mod(idx, ps.res!); // Use mod - modifies existing variable in hierarchy
  if (!ok) {
    // Check if it exists at all before suggesting 'var'
    var (_, exists, _) = ps.ctx.get2(idx); // Destructuring order is correct here based on env.dart
    if (exists) {
       ps.setError00("Cannot modify constant '${ps.idx.getWord(idx)}', use 'var' to declare it as a variable"); // Use ps.setError00
    } else {
       ps.setError00("Word '${ps.idx.getWord(idx)}' not found for modification"); // Use ps.setError00
    }
    ps.failureFlag = true; // Match Go behavior
  }
  // ps.res already holds the value
}


// Evaluates a found object (Builtin or Function) (mirrors Go's EvalObject)
void evalObject(ProgramState ps, RyeObject object, RyeObject? leftVal, bool toLeft, RyeCtx? ctx, bool pipeSecond, RyeObject? firstVal) {
  switch (object.type()) {
    case RyeType.builtinType:
      Builtin bu = object as Builtin;
      // Failure check is now inside callBuiltin
      callBuiltin(bu, ps, leftVal, toLeft, pipeSecond, firstVal);
      break;
    case RyeType.wordType: // Go treats Function as Word type internally sometimes
      if (object is RyeFunction) {
        callFunction(object, ps, leftVal, toLeft, ctx, pipeSecond, firstVal);
      } else {
        // If it's just a plain word found (e.g. from a CPath resolving to a word), return it
        ps.res = object;
      }
      break;
    // Add VarBuiltinType if its calling convention differs
    default:
      // Non-evaluatable object found (e.g., Integer, String), just return it
      ps.res = object;
  }
}

// --- Function & Builtin Calling ---

// Calls a user-defined function (mirrors Go's CallFunction)
void callFunction(RyeFunction fn, ProgramState ps, RyeObject? arg0_, bool toLeft, RyeCtx? callCtx, bool pipeSecond, RyeObject? firstVal) {
  // Determine execution context
  RyeCtx fnCtx;
  if (fn.pure) {
    fnCtx = RyeCtx(ps.pCtx); // Pure functions inherit from pure context
  } else if (fn.inCtx && fn.ctx != null) {
    fnCtx = fn.ctx!; // Execute in definition context
  } else {
    // Inherit from definition context (closure) or calling context
    fnCtx = RyeCtx(fn.ctx ?? callCtx ?? ps.ctx); 
  }

  // Bind arguments
  int argIndex = 0;
  RyeObject? firstEvaluatedArgForInject = null;

  // Handle potential first argument based on call type
  if (pipeSecond) {
     // Pipe: firstVal is arg0, arg0_ (leftVal) is arg1
     if (fn.argsn > 0) {
        RyeObject? specArgObj = fn.spec.series.get(argIndex);
        if (specArgObj is Word) {
          if (!fnCtx.setNew(specArgObj.index, firstVal ?? Void())) { ps.setError00("Pipe arg1 set fail"); return; }
        } else { ps.setError00("Function spec must contain words"); return; }
        firstEvaluatedArgForInject = firstVal ?? Void(); argIndex++;
     }
     if (fn.argsn > 1) {
        RyeObject? specArgObj = fn.spec.series.get(argIndex);
        if (specArgObj is Word) {
          if (!fnCtx.setNew(specArgObj.index, arg0_ ?? Void())) { ps.setError00("Pipe arg2 set fail"); return; }
        } else { ps.setError00("Function spec must contain words"); return; }
        argIndex++;
     }
  } else {
     // Normal/Opword: arg0_ (leftVal) is arg0 if toLeft
     if (arg0_ != null && toLeft && fn.argsn > 0) {
        RyeObject? specArgObj = fn.spec.series.get(argIndex);
        if (specArgObj is Word) {
          if (!fnCtx.setNew(specArgObj.index, arg0_)) { ps.setError00("Opword arg set fail"); return; }
        } else { ps.setError00("Function spec must contain words"); return; }
        firstEvaluatedArgForInject = arg0_; argIndex++;
     }
  }

  // Evaluate and bind remaining arguments from the calling series
  while (argIndex < fn.argsn) {
     RyeObject? specArgObj = fn.spec.series.get(argIndex); // Can be null
     if (specArgObj is Word) { // Check if it's a Word
        // No need to cast again, specArgObj is already known to be Word here
        evalExpression(ps, null, false, true); // Evaluate LIMITED
        if (ps.errorFlag || ps.returnFlag) { ps.setError00("Arg ${argIndex + 1} eval error", ps.res); return; } 
        if (ps.failureFlag) { ps.setError00("Arg ${argIndex + 1} eval failure", ps.res); return; } 
        // Use specArgObj.index directly
        if (!fnCtx.setNew(specArgObj.index, ps.res!)) { ps.setError00("Arg ${argIndex + 1} set fail"); return; } 
        firstEvaluatedArgForInject ??= ps.res!; 
        argIndex++;
     } else { ps.setError00("Function spec must contain words (got ${specArgObj?.runtimeType})"); return; } 
  }

  // Store caller state
  TSeries callerSeries = ps.ser;
  RyeCtx callerCtx = ps.ctx;
  bool callerReturnFlag = ps.returnFlag;
  // Store and clear flags relevant to the function call
  bool callerErrorFlag = ps.errorFlag; ps.errorFlag = false;
  bool callerFailureFlag = ps.failureFlag; ps.failureFlag = false;


  // Set up state for function execution
  ps.ser = fn.body.series;
  ps.ser.reset();
  ps.ctx = fnCtx;
  ps.returnFlag = false; // Reset return flag for the function body

  // Evaluate the function body
  evalBlockInj(ps, firstEvaluatedArgForInject, firstEvaluatedArgForInject != null); 

  // Execute deferred blocks *before* restoring caller state if function returned/errored
  if ((ps.returnFlag || ps.errorFlag) && ps.deferBlocks.isNotEmpty) {
     executeDeferredBlocks(ps);
  }

  // Store result and flags from function execution
  RyeObject? funcRes = ps.res;
  bool funcReturnFlag = ps.returnFlag;
  bool funcErrorFlag = ps.errorFlag;
  bool funcFailureFlag = ps.failureFlag; // Though failure should ideally be handled or become error

  // Restore caller state
  ps.ser = callerSeries;
  ps.ctx = callerCtx;
  ps.returnFlag = callerReturnFlag || funcReturnFlag; // Propagate return if function returned
  ps.errorFlag = callerErrorFlag || funcErrorFlag; // Propagate error if function errored
  ps.failureFlag = callerFailureFlag || funcFailureFlag; // Propagate failure? Or should it be error?
  ps.res = funcRes; // Result is the function's result

  // Go version uses envPool.Put(fnCtx) - Dart GC handles this.
}

// Calls a builtin function (mirrors Go's CallBuiltin)
void callBuiltin(Builtin bi, ProgramState ps, RyeObject? arg0_, bool toLeft, bool pipeSecond, RyeObject? firstVal) {
  // Fast path for fully curried builtins (simplified check)
  if (bi.argsn == 0 || (bi.argsn == 1 && bi.cur0 != null) /* ... add more checks if needed */) {
     if (rye00_checkForFailureWithBuiltin(bi, ps, 999)) return; // Check failure even for curried
     ps.res = bi.fn(ps, bi.cur0, bi.cur1, bi.cur2, bi.cur3, bi.cur4);
     return;
  }

  // Initialize arguments with curried values
  RyeObject? arg0 = bi.cur0;
  RyeObject? arg1 = bi.cur1;
  RyeObject? arg2 = bi.cur2;
  RyeObject? arg3 = bi.cur3;
  RyeObject? arg4 = bi.cur4;
  bool curry = false; // Flag to indicate if we are currying

  // --- Argument Handling based on pipeSecond ---
  if (pipeSecond) {
    // Pipeword logic: firstVal (evaluated after pipe) becomes arg0, arg0_ (value before pipe) becomes arg1
    if (bi.argsn > 0 && arg0 == null) {
      arg0 = firstVal;
      if (arg0 is Void) curry = true; // Check for currying on each arg
    }
    if (bi.argsn > 1 && arg1 == null) {
      arg1 = arg0_; // Value before pipe becomes second argument
      if (arg1 is Void) curry = true;
    }
    // Evaluate remaining args normally (limited)
    if (bi.argsn > 2 && arg2 == null) {
      evalExpression(ps, null, false, true); // Use evalExpression
      if (rye00_checkForFailureWithBuiltin(bi, ps, 2)) return; // Check failure after eval
      if (ps.errorFlag || ps.returnFlag) { ps.setError00("Missing pipe arg 3 for ${bi.doc}", ps.res); return; } // Use ps.setError00
      arg2 = ps.res;
      if (arg2 is Void) curry = true;
    }
     if (bi.argsn > 3 && arg3 == null) {
      evalExpression(ps, null, false, true);
      if (rye00_checkForFailureWithBuiltin(bi, ps, 3)) return;
      if (ps.errorFlag || ps.returnFlag) { ps.setError00("Missing pipe arg 4 for ${bi.doc}", ps.res); return; } // Use ps.setError00
      arg3 = ps.res;
      if (arg3 is Void) curry = true;
    }
     if (bi.argsn > 4 && arg4 == null) {
      evalExpression(ps, null, false, true);
      if (rye00_checkForFailureWithBuiltin(bi, ps, 4)) return;
      if (ps.errorFlag || ps.returnFlag) { ps.setError00("Missing pipe arg 5 for ${bi.doc}", ps.res); return; } // Use ps.setError00
      arg4 = ps.res;
      if (arg4 is Void) curry = true;
    }
  } else {
    // Normal / Opword logic
    if (bi.argsn > 0 && arg0 == null) {
      if (arg0_ != null && toLeft) { // Use leftVal if opword
        arg0 = arg0_;
      } else { // Evaluate next expression
        evalExpression(ps, null, false, true); // LIMITED
        if (rye00_checkForFailureWithBuiltin(bi, ps, 0)) return;
        if (ps.errorFlag || ps.returnFlag) { ps.setError00("Missing arg 1 for ${bi.doc}", ps.res); return; } // Use ps.setError00
        arg0 = ps.res;
      }
      if (arg0 is Void) curry = true;
    }
    if (bi.argsn > 1 && arg1 == null) {
      evalExpression(ps, null, false, true); // LIMITED
      if (rye00_checkForFailureWithBuiltin(bi, ps, 1)) return;
      if (ps.errorFlag || ps.returnFlag) { ps.setError00("Missing arg 2 for ${bi.doc}", ps.res); return; } // Use ps.setError00
      arg1 = ps.res;
      if (arg1 is Void) curry = true;
    }
    if (bi.argsn > 2 && arg2 == null) {
      evalExpression(ps, null, false, true); // LIMITED
      if (rye00_checkForFailureWithBuiltin(bi, ps, 2)) return;
      if (ps.errorFlag || ps.returnFlag) { ps.setError00("Missing arg 3 for ${bi.doc}", ps.res); return; } // Use ps.setError00
      arg2 = ps.res;
      if (arg2 is Void) curry = true;
    }
    if (bi.argsn > 3 && arg3 == null) {
      evalExpression(ps, null, false, true); // LIMITED
      if (rye00_checkForFailureWithBuiltin(bi, ps, 3)) return;
      if (ps.errorFlag || ps.returnFlag) { ps.setError00("Missing arg 4 for ${bi.doc}", ps.res); return; } // Use ps.setError00
      arg3 = ps.res;
      if (arg3 is Void) curry = true;
    }
    if (bi.argsn > 4 && arg4 == null) {
      evalExpression(ps, null, false, true); // LIMITED
      if (rye00_checkForFailureWithBuiltin(bi, ps, 4)) return;
      if (ps.errorFlag || ps.returnFlag) { ps.setError00("Missing arg 5 for ${bi.doc}", ps.res); return; } // Use ps.setError00
      arg4 = ps.res;
      if (arg4 is Void) curry = true;
    }
  }
  // --- End Argument Handling ---

  // Final check for failure before calling the function
  if (rye00_checkForFailureWithBuiltin(bi, ps, 999)) return;

  if (curry) {
    // Return a new Builtin object with the collected arguments curried
    ps.res = Builtin(bi.fn, bi.argsn, bi.acceptFailure, bi.pure, bi.doc,
      cur0: arg0, cur1: arg1, cur2: arg2, cur3: arg3, cur4: arg4);
  } else {
    // Call the actual builtin function
    ps.res = bi.fn(ps, arg0, arg1, arg2, arg3, arg4);
  }
}

// TODO: Implement CallVarBuiltin if needed (Go has it for variadic builtins)


// --- Error and Failure Handling ---

// Checks failure flag before calling a builtin (mirrors Go's checkForFailureWithBuiltin)
bool rye00_checkForFailureWithBuiltin(Builtin bi, ProgramState ps, int n) { // Renamed to avoid conflict if needed
  if (ps.failureFlag) {
    if (bi.acceptFailure) {
      // Builtin accepts failure, proceed.
      return false; 
    } else {
      // Builtin does not accept failure, promote to error.
      ps.setError00("Unhandled failure before calling ${bi.doc}", ps.res); // Use ps.setError00
      return true; // Indicate error occurred
    }
  }
  return false; // No failure
}

// TODO: Implement checkForFailureWithVarBuiltin if VarBuiltin is added

// Attempts to handle a failure by looking for and executing an 'error-handler'.
// Returns true if the failure was *not* handled (and thus should propagate or become an error),
// Returns false if the failure *was* handled by an error-handler. (Mirrors Go's tryHandleFailure)
bool tryHandleFailure(ProgramState ps) {
  if (ps.failureFlag && !ps.returnFlag && !ps.inErrHandler) {
    
    // Ensure the failure object (ps.res) has context and position info
    if (ps.res is Error) {
      Error failureObj = ps.res as Error;
      failureObj.codeContext ??= ps.ctx; 
      failureObj.position ??= ps.ser.getPos() > 0 ? ps.ser.getPos() - 1 : 0;
       // TODO: Add source file path if available
    } else {
       // Wrap non-Error failure value
       ps.res = Error("Failure value was not an Error object: ${ps.res?.inspect(ps.idx) ?? 'null'}", 
                      codeContext: ps.ctx, 
                      position: ps.ser.getPos() > 0 ? ps.ser.getPos() - 1 : 0);
    }

    // Check for 'error-handler' word
    var (errHandlerIdx, wordExists) = ps.idx.getIndex("error-handler");
    if (!wordExists) {
      ps.errorFlag = true; // Promote unhandled failure to error
      return true; // Indicate failure was not handled
    }

    // Check if 'error-handler' is defined in the context hierarchy
    var (handlerObj, handlerExists, _) = ps.ctx.get2(errHandlerIdx); // Destructuring order is correct here
    if (!handlerExists || handlerObj is! Block) {
       ps.errorFlag = true; // Promote unhandled failure to error
       return true; // Indicate failure was not handled (handler missing or not a block)
    }

    // Execute the handler block
    Block handlerBlock = handlerObj;
    TSeries ser = ps.ser; // Store current series
    RyeCtx ctx = ps.ctx; // Store current context
    ps.ser = handlerBlock.series;
    ps.ser.reset();
    ps.inErrHandler = true; // Prevent recursive error handling

    // Evaluate the handler, injecting the failure object (current ps.res)
    evalBlockInj(ps, ps.res, true); 

    ps.inErrHandler = false; // Exit error handling mode
    ps.ser = ser; // Restore original series
    ps.ctx = ctx; // Restore original context

    // If the handler itself resulted in an error, keep the error flag.
    // Otherwise, the failure is considered handled (clear failure flag).
    if (!ps.errorFlag) {
       ps.failureFlag = false; 
    }
    return false; // Indicate failure was handled (or handler errored)
  }
  // No failure, or return flag set, or already in handler
  return false; 
}

// Displays failure or error information (mirrors Go's MaybeDisplayFailureOrError)
void maybeDisplayFailureOrError(ProgramState ps, String tag) {
  if (ps.failureFlag && !ps.inErrHandler) { // Don't display during handling
    stdout.write("\x1b[33mFailure: ${ps.res?.print(ps.idx) ?? ''}\x1b[0m\n"); 
    // Add position/context if available in ps.res (Error object)
    if (ps.res is Error) {
       Error err = ps.res as Error;
       if (err.position != null) stdout.write("  at position: ${err.position}\n");
       // Consider printing context info from err.codeContext if useful
    }
    // stdout.write(" ($tag)\n"); // Optional tag
  }
  if (ps.errorFlag) {
    stdout.write("\x1b[31mError: ${ps.res?.print(ps.idx) ?? 'Unknown Error'}\x1b[0m\n"); 
    if (ps.res is Error) {
       Error err = ps.res as Error;
       if (err.position != null) stdout.write("  at position: ${err.position}\n");
       // Consider printing context info from err.codeContext if useful
    }
     // stdout.write(" ($tag)\n"); // Optional tag
  }
}


// --- Opword/Pipeword Handling ---

// Checks the next token and evaluates if it's an opword/pipeword (mirrors Go's MaybeEvalOpwordOnRight)
void maybeEvalOpwordOnRight(ProgramState ps, {bool limited = false}) {
  if (ps.returnFlag || ps.errorFlag) {
    return; // Don't process if already returning or errored
  }

  RyeObject? nextObj = ps.ser.peek();
  if (nextObj == null) {
    return; // End of series
  }

  // Quick exit for common non-op types
  RyeType objType = nextObj.type();
   if (objType == RyeType.integerType ||
       objType == RyeType.stringType ||
       objType == RyeType.blockType || // Mode 0 blocks don't trigger opwords
       objType == RyeType.voidType ||
       objType == RyeType.commaType || // Commas handled by evalBlockInj
       (objType == RyeType.wordType && nextObj is! OpWord && nextObj is! PipeWord && nextObj is! LSetword && nextObj is! LModword && nextObj is! CPath) // Plain words don't trigger
       ) {
     return;
   }


  // Handle specific op/pipe types
  if (nextObj is OpWord) {
    ps.ser.next(); // Consume the opword
    RyeObject? valueBeforeOp = ps.res; // Value from the left
    // Evaluate the word associated with the opword, passing the left value
    evalWord(ps, nextObj, valueBeforeOp, true, false, null); // toLeft = true
    // Recursively check for more opwords after evaluation
    if (!ps.errorFlag && !ps.returnFlag) {
       maybeEvalOpwordOnRight(ps, limited: limited);
    }
  } else if (nextObj is PipeWord) {
    if (limited) return; // Don't evaluate pipewords when limited
    ps.ser.next(); // Consume the pipeword
    RyeObject? valueBeforePipe = ps.res; // Value from the left

    // Evaluate the expression *after* the pipeword first
    RyeObject? firstVal; // This will be arg0 for the piped function
    int peekPos = ps.ser.getPos();
    if (!ps.ser.atLast()) {
      evalExpression(ps, null, false, false); // Evaluate right side, NOT limited
      if (ps.errorFlag || ps.returnFlag) return;
      if (ps.failureFlag) { ps.setError00("Failure evaluating right side of pipe", ps.res); return; } // Use ps.setError00
      firstVal = ps.res;
    } else {
      ps.setError00("Pipeword requires an expression following it"); // Use ps.setError00
      return;
    }

    // Now, evaluate the pipeword itself, passing both values
    evalWord(ps, nextObj, valueBeforePipe, false, true, firstVal); // pipeSecond = true

    // Recursively check for more opwords after evaluation
    if (!ps.errorFlag && !ps.returnFlag) {
       maybeEvalOpwordOnRight(ps, limited: limited);
    }
  } else if (nextObj is LSetword) {
    if (limited) return; // Don't evaluate l-setwords when limited
    ps.ser.next(); // Consume
    LSetword word = nextObj;
    int idx = word.index;
    // Go version checks ps.AllowMod here, but uses SetNew regardless? Let's use SetNew.
    bool ok = ps.ctx.setNew(idx, ps.res!); 
    if (!ok) {
      ps.setError00("Can't set already set word '${ps.idx.getWord(idx)}', try using '::'"); // Use ps.setError00
      ps.failureFlag = true; return;
    }
    // Recursively check
    if (!ps.errorFlag && !ps.returnFlag) {
       maybeEvalOpwordOnRight(ps, limited: limited);
    }
  } else if (nextObj is LModword) {
     if (limited) return; // Don't evaluate l-modwords when limited
     ps.ser.next(); // Consume
     LModword word = nextObj;
     int idx = word.index;
     bool ok = ps.ctx.mod(idx, ps.res!); 
     if (!ok) {
       var (_, exists, _) = ps.ctx.get2(idx); // Destructuring order is correct here
       if (exists) {
          ps.setError00("Cannot modify constant '${ps.idx.getWord(idx)}', use 'var' to declare it"); // Use ps.setError00
       } else {
          ps.setError00("Word '${ps.idx.getWord(idx)}' not found for modification"); // Use ps.setError00
       }
       ps.failureFlag = true; return;
     }
     // Recursively check
     if (!ps.errorFlag && !ps.returnFlag) {
        maybeEvalOpwordOnRight(ps, limited: limited);
     }
  } else if (nextObj is CPath) {
    CPath path = nextObj;
    if (path.mode == 1) { // Opword CPath
      ps.ser.next(); // Consume
      RyeObject? valueBeforeOp = ps.res;
      evalWord(ps, path, valueBeforeOp, true, false, null); // toLeft = true
      if (!ps.errorFlag && !ps.returnFlag) {
         maybeEvalOpwordOnRight(ps, limited: limited);
      }
    } else if (path.mode == 2) { // Pipeword CPath
      if (limited) return;
      ps.ser.next(); // Consume
      RyeObject? valueBeforePipe = ps.res;
      RyeObject? firstVal;
      if (!ps.ser.atLast()) {
         evalExpression(ps, null, false, false); // Eval right side
         if (ps.errorFlag || ps.returnFlag) return;
         if (ps.failureFlag) { ps.setError00("Failure evaluating right side of CPath pipe", ps.res); return; } // Use ps.setError00
         firstVal = ps.res;
      } else { ps.setError00("CPath pipe requires an expression following it"); return; } // Use ps.setError00
      
      evalWord(ps, path, valueBeforePipe, false, true, firstVal); // pipeSecond = true
      if (!ps.errorFlag && !ps.returnFlag) {
         maybeEvalOpwordOnRight(ps, limited: limited);
      }
    }
    // Mode 0 CPaths are handled by evalExpressionConcrete calling evalWord
  }
}


// --- Deferred Execution ---

// Executes deferred blocks in LIFO order (mirrors Go's ExecuteDeferredBlocks)
void executeDeferredBlocks(ProgramState ps) {
  // Execute in reverse order (Last-In, First-Out)
  for (int i = ps.deferBlocks.length - 1; i >= 0; i--) {
    Block block = ps.deferBlocks[i];

    // Save crucial state parts that shouldn't be affected by defer
    TSeries oldSer = ps.ser;
    RyeCtx oldCtx = ps.ctx; // Save context too
    RyeObject? oldRes = ps.res; 
    bool oldFailureFlag = ps.failureFlag;
    bool oldErrorFlag = ps.errorFlag;
    bool oldReturnFlag = ps.returnFlag; // Defer runs *before* return propagates

    // Evaluate the deferred block in the current context
    ps.ser = block.series;
    ps.ser.reset();
    // Clear flags before evaluating defer block? Go version doesn't seem to.
    // ps.errorFlag = false; 
    // ps.failureFlag = false;
    // ps.returnFlag = false; // Defer block shouldn't cause outer function to return

    evalBlockInj(ps, null, false); // Evaluate the block without injection

    // Restore state, keeping any *new* error from the defer block
    ps.ser = oldSer;
    ps.ctx = oldCtx; // Restore context
    if (!ps.errorFlag) { // If defer didn't cause an error, restore original result/flags
       ps.res = oldRes;
       ps.failureFlag = oldFailureFlag;
       ps.returnFlag = oldReturnFlag;
    } else {
       // If defer *did* cause an error, keep the error state but potentially
       // restore the original return flag if the defer wasn't meant to stop it.
       ps.returnFlag = oldReturnFlag; 
       // Should the failure flag also be restored? Go seems to restore it.
       ps.failureFlag = oldFailureFlag; 
    }
  }
  // Clear the list after execution
  ps.deferBlocks.clear();
}
