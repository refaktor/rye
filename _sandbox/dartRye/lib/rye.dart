// rye.dart - A simplified Rye language evaluator in Dart

import 'dart:io';
import 'dart:async';
import 'flutter/flutter_app.dart';
import 'loader.dart' show Setword, OpWord, PipeWord, LSetword, LModword, CPath, Comma, LitWord; // Import LitWord
import 'builtins.dart';
import 'builtins_registration.dart';

// Type enum for Rye values
enum RyeType {
  blockType,
  integerType,
  wordType,
  builtinType,
  errorType,
  voidType,
  stringType,
  listType,
  dictType,
  contextType,
  booleanType,
  decimalType,
  timeType,
  dateType,
  uriType,
  emailType,
  vectorType,
}

// Base class for all Rye values
abstract class RyeObject {
  RyeType type();
  String print(Idxs idxs);
  String inspect(Idxs idxs);
  bool equal(RyeObject other);
  int getKind();
}

// Integer implementation
class Integer implements RyeObject {
  int value;

  Integer(this.value);

  @override
  RyeType type() => RyeType.integerType;

  @override
  String print(Idxs idxs) {
    return value.toString();
  }

  @override
  String inspect(Idxs idxs) {
    return '[Integer: ${print(idxs)}]';
  }

  @override
  bool equal(RyeObject other) {
    if (other.type() != RyeType.integerType) return false;
    return value == (other as Integer).value;
  }

  @override
  int getKind() => RyeType.integerType.index;
}

// Word implementation
class Word implements RyeObject {
  final int index;

  const Word(this.index);

  @override
  RyeType type() => RyeType.wordType;

  @override
  String print(Idxs idxs) {
    return idxs.getWord(index);
  }

  @override
  String inspect(Idxs idxs) {
    return '[Word: ${print(idxs)}]';
  }

  @override
  bool equal(RyeObject other) {
    if (other.type() != RyeType.wordType) return false;
    return index == (other as Word).index;
  }

  @override
  int getKind() => RyeType.wordType.index;
}

// Error implementation
class Error implements RyeObject {
  String message;
  int status;
  Error? parent;
  RyeCtx? codeContext; // Context where the error occurred
  int? position;      // Position in the series where the error occurred

  Error(this.message, {this.status = 0, this.parent, this.codeContext, this.position});

  @override
  RyeType type() => RyeType.errorType;

  @override
  String print(Idxs idxs) {
    String statusStr = status != 0 ? "($status)" : "";
    StringBuffer b = StringBuffer();
    b.write("Error$statusStr: $message ");
    
    if (parent != null) {
      b.write("\n  ${parent!.print(idxs)}");
    }
    
    // Optionally add position info if available
    if (position != null) {
       b.write(" (at pos $position)");
    }
    
    return b.toString();
  }

  @override
  String inspect(Idxs idxs) {
    return '[${print(idxs)}]';
  }

  @override
  bool equal(RyeObject other) {
    if (other.type() != RyeType.errorType) return false;
    
    Error otherError = other as Error;
    if (status != otherError.status) return false;
    if (message != otherError.message) return false;
    
    // Check if both have parent or both don't have parent
    if ((parent == null) != (otherError.parent == null)) return false;
    
    // If both have parent, check if parents are equal
    if (parent != null && otherError.parent != null) {
      if (!parent!.equal(otherError.parent!)) return false;
    }
    
    return true;
  }

  @override
  int getKind() => RyeType.errorType.index;
}

// Void implementation
class Void implements RyeObject {
  const Void();

  @override
  RyeType type() => RyeType.voidType;

  @override
  String print(Idxs idxs) {
    return '_';
  }

  @override
  String inspect(Idxs idxs) {
    return '[Void]';
  }

  @override
  bool equal(RyeObject other) {
    return other.type() == RyeType.voidType;
  }

  @override
  int getKind() => RyeType.voidType.index;
}

// TSeries represents a series of Objects with a position pointer.
class TSeries {
  List<RyeObject?> s; // The underlying list of Objects
  int pos = 0; // Current position in the series

  TSeries(this.s);

  bool ended() {
    return pos > s.length;
  }

  bool atLast() {
    return pos > s.length - 1;
  }

  int getPos() {
    return pos;
  }

  void next() {
    pos++;
  }

  RyeObject? pop() {
    if (pos >= s.length) {
      return null;
    }
    RyeObject? obj = s[pos];
    pos++;
    return obj;
  }

  bool put(RyeObject obj) {
    if (pos > 0 && pos <= s.length) {
      s[pos - 1] = obj;
      return true;
    }
    return false;
  }

  TSeries append(RyeObject obj) {
    s.add(obj);
    return this;
  }

  void reset() {
    pos = 0;
  }

  void setPos(int position) {
    pos = position;
  }

  List<RyeObject?> getAll() {
    return s;
  }

  RyeObject? peek() {
    if (s.length > pos) {
      return s[pos];
    }
    return null;
  }

  RyeObject? get(int n) {
    if (n >= 0 && n < s.length) {
      return s[n];
    }
    return null;
  }

  int len() {
    return s.length;
  }
}

// Block implementation
class Block implements RyeObject {
  TSeries series;
  int mode;

  Block(this.series, [this.mode = 0]);

  @override
  RyeType type() => RyeType.blockType;

  @override
  String print(Idxs idxs) {
    StringBuffer r = StringBuffer();
    for (int i = 0; i < series.len(); i++) {
      if (series.get(i) != null) {
        r.write(series.get(i)!.print(idxs));
        r.write(' ');
      } else {
        r.write('[NIL]');
      }
    }
    return r.toString();
  }

  @override
  String inspect(Idxs idxs) {
    StringBuffer r = StringBuffer();
    r.write('[Block: ');
    for (int i = 0; i < series.len(); i++) {
      if (series.get(i) != null) {
        if (series.getPos() == i) {
          r.write('^');
        }
        r.write(series.get(i)!.inspect(idxs));
        r.write(' ');
      }
    }
    r.write(']');
    return r.toString();
  }

  @override
  bool equal(RyeObject other) {
    if (other.type() != RyeType.blockType) return false;
    
    Block otherBlock = other as Block;
    if (series.len() != otherBlock.series.len()) return false;
    if (mode != otherBlock.mode) return false;
    
    for (int j = 0; j < series.len(); j++) {
      if (!series.get(j)!.equal(otherBlock.series.get(j)!)) {
        return false;
      }
    }
    return true;
  }

  @override
  int getKind() => RyeType.blockType.index;
}

// Idxs is a bidirectional mapping between words (strings) and their indices.
class Idxs {
  List<String> words1 = [""];
  Map<String, int> words2 = {};

  Idxs() {
    // Register some basic words
    indexWord("_+");
    indexWord("integer");
    indexWord("word");
    indexWord("block");
    indexWord("builtin");
    indexWord("error");
    indexWord("void");
  }

  int indexWord(String w) {
    int? idx = words2[w];
    if (idx != null) {
      return idx;
    } else {
      words1.add(w);
      words2[w] = words1.length - 1;
      return words1.length - 1;
    }
  }

  (int, bool) getIndex(String w) {
    int? idx = words2[w];
    if (idx != null) {
      return (idx, true);
    }
    return (0, false);
  }

  String getWord(int i) {
    if (i < 0) {
      return "isolate!";
    }
    return words1[i];
  }

  int getWordCount() {
    return words1.length;
  }
}

// RyeCtx represents a context in Rye, which is a mapping from words to values.
class RyeCtx {
  Map<int, RyeObject> state = {};
  Map<int, bool> varFlags = {};
  RyeCtx? parent;
  Word kind;

  RyeCtx(this.parent, [this.kind = const Word(0)]);

  (RyeObject?, bool) get(int word) {
    RyeObject? obj = state[word];
    bool exists = obj != null;
    
    if (!exists && parent != null) {
      var (obj1, exists1) = parent!.get(word);
      if (exists1) {
        obj = obj1;
        exists = exists1;
      }
    }
    
    return (obj, exists);
  }

  (RyeObject?, bool, RyeCtx) get2(int word) {
    RyeObject? obj = state[word];
    bool exists = obj != null;
    
    if (!exists && parent != null) {
      var (obj1, exists1, ctx) = parent!.get2(word);
      if (exists1) {
        obj = obj1;
        exists = exists1;
        return (obj, exists, ctx);
      }
    }
    
    return (obj, exists, this);
  }

  RyeObject set(int word, RyeObject val) {
    if (state.containsKey(word)) {
      return Error("Can't set already set word, try using modword! FIXME !");
    } else {
      state[word] = val;
      return val;
    }
  }

  bool setNew(int word, RyeObject val) {
    if (state.containsKey(word)) {
      return false;
    } else {
      state[word] = val;
      return true;
    }
  }

  // Checks if a word exists in this context (not parent contexts)
  bool has(int word) {
    return state.containsKey(word);
  }

  // Marks a word as a variable (can be modified)
  void markAsVariable(int word) {
    varFlags[word] = true;
  }

  // Checks if a word is marked as a variable
  bool isVariable(int word) {
    // Check current context first
    if (varFlags.containsKey(word)) {
      return varFlags[word] ?? false;
    }
    
    // Check parent contexts recursively
    if (parent != null && !state.containsKey(word)) {
      return parent!.isVariable(word);
    }
    
    // Default to false if not found
    return false;
  }

  // Removes a word from the context
  // Returns true if successful, false if the word doesn't exist
  bool unset(int word) {
    if (state.containsKey(word)) {
      state.remove(word);
      varFlags.remove(word);
      return true;
    }
    return false;
  }

  // Modifies an existing variable in the current context or parent contexts.
  // Returns true if successful, false if the variable doesn't exist or is not modifiable.
  bool mod(int word, RyeObject val) {
    // Check current context first
    if (state.containsKey(word)) {
      // Check if the word is marked as a variable
      if (isVariable(word)) {
        state[word] = val;
        return true;
      }
      return false; // Word exists but is not a variable
    }
    
    // Check parent contexts recursively
    if (parent != null) {
      return parent!.mod(word, val);
    }
    
    // Word not found in hierarchy
    return false;
  }
}

// Builtin function type definition
typedef BuiltinFunction = RyeObject Function(ProgramState ps, RyeObject? arg0, RyeObject? arg1, RyeObject? arg2, RyeObject? arg3, RyeObject? arg4);

// Builtin implementation
class Builtin implements RyeObject {
  BuiltinFunction fn;
  int argsn;
  RyeObject? cur0;
  RyeObject? cur1;
  RyeObject? cur2;
  RyeObject? cur3;
  RyeObject? cur4;
  bool acceptFailure;
  bool pure;
  String doc;

  Builtin(this.fn, this.argsn, this.acceptFailure, this.pure, this.doc, 
      {this.cur0, this.cur1, this.cur2, this.cur3, this.cur4});

  @override
  RyeType type() => RyeType.builtinType;

  @override
  String print(Idxs idxs) {
    String pureStr = pure ? 'Pure ' : '';
    return '${pureStr}BFunction($argsn): $doc';
  }

  @override
  String inspect(Idxs idxs) {
    return '[${print(idxs)}]';
  }

  @override
  bool equal(RyeObject other) {
    if (other.type() != RyeType.builtinType) return false;
    
    Builtin otherBuiltin = other as Builtin;
    if (argsn != otherBuiltin.argsn) return false;
    if (cur0 != otherBuiltin.cur0) return false;
    if (cur1 != otherBuiltin.cur1) return false;
    if (cur2 != otherBuiltin.cur2) return false;
    if (cur3 != otherBuiltin.cur3) return false;
    if (cur4 != otherBuiltin.cur4) return false;
    if (acceptFailure != otherBuiltin.acceptFailure) return false;
    if (pure != otherBuiltin.pure) return false;
    
    return true;
  }

  @override
  int getKind() => RyeType.builtinType.index;
}

// Function implementation (user-defined)
// NOTE: This requires parser support for function definition syntax (e.g., name: fn [spec] { body })
// or a 'make-function' builtin.
class RyeFunction implements RyeObject {
  Block spec; // Block containing argument words
  Block body; // Block containing the function code
  RyeCtx? ctx; // Context where the function was defined (for closure)
  bool pure;   // Is the function pure?
  bool inCtx;  // Should the function execute in its definition context?
  int argsn;   // Number of arguments expected

  RyeFunction(this.spec, this.body, this.ctx, {this.pure = false, this.inCtx = false})
    : argsn = spec.series.len(); // Calculate argsn based on spec length

  @override
  RyeType type() => RyeType.wordType; // Treat as word-like for now (like Go)

  @override
  String print(Idxs idxs) {
    // Basic representation, could be improved
    return "Function($argsn)"; 
  }

  @override
  String inspect(Idxs idxs) {
     return '[Function: Args ${spec.print(idxs)} Body ${body.print(idxs)}]';
  }

  @override
  bool equal(RyeObject other) {
     // Functions are generally compared by reference or definition location,
     // simple equality check might not be meaningful. For now, reference equality.
     return identical(this, other); 
  }

  @override
  int getKind() => type().index;
}


// ProgramState represents the state of a Rye program.
class ProgramState {
  TSeries ser; // current block of code
  RyeObject? res; // result of expression
  RyeCtx ctx; // Env object ()
  RyeCtx pCtx; // Env object () -- pure context
  Idxs idx; // Idx object (index of words)
  List<int> args; // names of current arguments (indexes of names)
  RyeObject? inj; // Injected first value in a block evaluation
  bool injnow = false;
  bool returnFlag = false;
  bool errorFlag = false;
  bool failureFlag = false;
  RyeObject? forcedResult;
  bool skipFlag = false;
  bool inErrHandler = false;
  List<Block> deferBlocks = []; // List to hold deferred blocks
  
  // Generic word registry: Map<TypeKind, Map<WordIndex, RyeObject>>
  Map<int, Map<int, RyeObject>> gen = {}; 

  ProgramState(this.ser, this.idx)
      : ctx = RyeCtx(null),
        pCtx = RyeCtx(null),
        args = List.filled(6, 0);

  // Helper to register a generic word/builtin
  void registerGeneric(int typeKind, int wordIdx, RyeObject value) {
    if (!gen.containsKey(typeKind)) {
      gen[typeKind] = {};
    }
    gen[typeKind]![wordIdx] = value;
  }

   // Helper to get a generic word/builtin
  (RyeObject?, bool) getGeneric(int typeKind, int wordIdx) {
    if (gen.containsKey(typeKind) && gen[typeKind]!.containsKey(wordIdx)) {
      return (gen[typeKind]![wordIdx], true);
    }
    return (null, false);
  }
}

// Pre-allocated common error messages to avoid allocations
final Error errMissingValue = Error("Expected Rye value but it's missing");
final Error errExpressionGuard = Error("Expression guard inside expression");
final Error errErrorObject = Error("Error object encountered");
final Error errUnsupportedType = Error("Unsupported type in simplified interpreter");
final Error errArg1Missing = Error("Argument 1 missing for builtin");
final Error errArg2Missing = Error("Argument 2 missing for builtin");
final Error errArg3Missing = Error("Argument 3 missing for builtin");

// Helper function to set error state - adds context and position
void setError00(ProgramState ps, String message, [RyeObject? cause]) {
  ps.errorFlag = true;
  int currentPos = ps.ser.getPos() > 0 ? ps.ser.getPos() -1 : 0; // Position of the item that caused the error

  // Create the new error with context and position
  Error newError = Error(message, codeContext: ps.ctx, position: currentPos);

  // Link cause if provided
  if (cause is Error) {
    newError.parent = cause;
  }

  // Use pre-allocated errors for common messages *if* no specific cause/context needed
  // Otherwise, use the newError with context. For simplicity now, always use newError.
  /* switch (message) {
      ps.res = newError; // Always use the error with context
      break;
  } */
  ps.res = newError; // Assign the new error with context
}

// Rye00_findWordValue returns the value associated with a word in the current context.
(bool, RyeObject?, RyeCtx?) rye00_findWordValue(ProgramState ps, RyeObject word1) {
  // Extract the word index
  int index;
  if (word1 is Word) {
    index = word1.index;
  } else {
    return (false, null, null);
  }

  // First try to get the value from the current context
  var (object, found) = ps.ctx.get(index);
  if (found && object != null) {
    // Enable word replacement optimization for builtins
    if (object.type() == RyeType.builtinType && ps.ser.getPos() > 0) {
      ps.ser.put(object);
    }
    return (found, object, null);
  }

  // If not found in the current context and there's no parent, return not found
  if (ps.ctx.parent == null) {
    return (false, null, null);
  }

  // Try to get the value from parent contexts
  var (object2, found2, foundCtx) = ps.ctx.get2(index);
  return (found2, object2, foundCtx);
}

// Rye00_EvalExpressionConcrete evaluates a concrete expression.
void rye00_evalExpressionConcrete(ProgramState ps) {
  RyeObject? object = ps.ser.pop();

  if (object == null) {
    ps.errorFlag = true;
    ps.res = errMissingValue;
    return;
  }

  if (object is Setword) {
    rye00_evalWord(ps, object, null, false, false);
  } else {
    switch (object.type()) {
      case RyeType.integerType:
      case RyeType.stringType:
      case RyeType.listType:
      case RyeType.dictType:
      case RyeType.contextType:
      case RyeType.voidType:
      case RyeType.booleanType:
      case RyeType.decimalType:
      case RyeType.timeType:
      case RyeType.dateType:
      case RyeType.uriType:
      case RyeType.emailType:
      case RyeType.vectorType:
      // Also add LitWord here, although its type() might still be wordType
      // A direct type check is safer.
        if (object is LitWord) {
           ps.res = object; // LitWords evaluate to themselves
        } else {
           ps.res = object;
        }
        break;
      case RyeType.blockType:
        Block block = object as Block;
        if (block.mode == 0) {
          // Mode 0: Return the block itself
          ps.res = object;
        } else if (block.mode == 1) {
          // Mode 1 []: Evaluate items, return new block with results
          TSeries ser = ps.ser; // Store current series
          ps.ser = block.series;
          ps.ser.reset();
          List<RyeObject?> results = [];
          while (ps.ser.getPos() < ps.ser.len()) {
            rye00_evalExpressionConcrete(ps); // Evaluate expression within the block
            if (ps.errorFlag || ps.returnFlag) {
              ps.ser = ser; // Restore series on error/return
              return; 
            }
            results.add(ps.res);
          }
          ps.ser = ser; // Restore original series
          ps.res = Block(TSeries(results)); // Return new block with results
        } else if (block.mode == 2) {
          // Mode 2 {}: Return the block itself (like Mode 0)
          ps.res = object; 
        } else {
           setError00(ps, "Unsupported block mode: ${block.mode}");
        }
        break;
      case RyeType.wordType:
        // Check if it's a CPath (even though type is wordType for now)
        if (object is CPath) {
           // Mode 0 CPath encountered directly
           rye00_evalWord(ps, object, null, false, false);
        } else if (object is LitWord) {
           // LitWord encountered directly (should evaluate to itself)
           ps.res = object;
        } else if (object is Word) {
           // Regular Word (includes OpWord, PipeWord etc. if not handled by maybeEvalOpwordOnRight)
           rye00_evalWord(ps, object, null, false, false);
        }
        // Note: OpWord, PipeWord, LSetword, LModword should ideally be handled by maybeEvalOpwordOnRight
        break;
      case RyeType.builtinType:
        rye00_callBuiltin(object as Builtin, ps, null, false, false, null);
        break;
      case RyeType.errorType:
        setError00(ps, "Error object encountered");
        break;
      default:
        setError00(ps, "Unsupported type in simplified interpreter: ${object.type().index}");
        break;
    }
  }
  
  // After evaluating the concrete object, check if an opword follows
  maybeEvalOpwordOnRight(ps);
}

// Rye00_EvalWord evaluates a word in the current context.
void rye00_evalWord(ProgramState ps, RyeObject word, RyeObject? leftVal, bool toLeft, bool pipeSecond) {
  // Handle Setword objects
  if (word is Setword) {
    // Get the next value
    rye00_evalExpressionConcrete(ps);
    
    if (ps.errorFlag || ps.failureFlag) {
      return;
    }
    
    // Set the value in the context
    ps.ctx.set(word.index, ps.res!);
    return;
  }
  // Determine the kind of the preceding value for potential generic dispatch
  int kind = -1;
  RyeObject? precedingVal = null;
  if (leftVal != null) {
    kind = leftVal.getKind();
    precedingVal = leftVal;
  /* else if (pipeSecond && firstVal != null) { // TODO: Fix firstVal/arg0_ handling for pipeSecond
    kind = firstVal.getKind();
    precedingVal = firstVal;
  } else if (pipeSecond && arg0_ != null) { // Use arg0_ if firstVal is null in pipeSecond scenario
     kind = arg0_.getKind();
     precedingVal = arg0_;
  } */
  }

  // Try Generic Dispatch FIRST if there's a preceding value and it's a standard Word type
  if (kind != -1 && word is Word && word is! OpWord && word is! PipeWord && word is! LSetword && word is! LModword && word is! CPath && word is! Setword && word is! LitWord) {
     var (genericObj, genericFound) = ps.getGeneric(kind, word.index);
     if (genericFound) {
       // Found in generic registry, evaluate it, passing the preceding value
       // For generic dispatch, leftVal (or precedingVal) is the object the word operates on.
       // TODO: Fix firstVal handling
       rye00_evalObject(ps, genericObj!, precedingVal, true, null, pipeSecond, null /* firstVal */);
       return; // Exit after successful generic dispatch
     }
  }

  // If not found generically (or no preceding value), try normal context lookup
  var (found, object, session) = rye00_findWordValue(ps, word);
  if (found) {
    // Found in normal context
    // TODO: Fix firstVal handling
    rye00_evalObject(ps, object!, leftVal, toLeft, session, pipeSecond, null /* firstVal */);
  } else {
    // Still not found after checking generic and normal contexts
    setError00(ps, "Word not found: ${word.print(ps.idx)}");
  }
}

// Rye00_EvalObject evaluates a Rye object.
void rye00_evalObject(ProgramState ps, RyeObject object, RyeObject? leftVal, bool toLeft, RyeCtx? ctx, bool pipeSecond, RyeObject? firstVal) {
  switch (object.type()) {
    case RyeType.builtinType:
      Builtin bu = object as Builtin;

      if (rye00_checkForFailureWithBuiltin(bu, ps, 333)) {
        return;
      }
      rye00_callBuiltin(bu, ps, leftVal, toLeft, pipeSecond, firstVal);
      return;
    case RyeType.wordType: // Check if it's actually a Function disguised as a Word
       if (object is RyeFunction) { // Use renamed RyeFunction
         // TODO: Handle leftVal, toLeft, pipeSecond correctly if needed for function calls
         rye00_callFunction(object, ps, leftVal, toLeft, ctx);
         return;
       }
       // Fallthrough if it's just a regular word (shouldn't happen if lookup was correct)
       ps.res = object;
       break; 
    default:
      ps.res = object;
  }
}

// Helper function to call a user-defined function
void rye00_callFunction(RyeFunction fn, ProgramState ps, RyeObject? arg0_, bool toLeft, RyeCtx? callCtx) { // Use renamed RyeFunction
  // Determine the context for the function execution
  RyeCtx fnCtx;
  RyeCtx? parentCtx;

  if (fn.pure) {
    parentCtx = ps.pCtx; // Pure functions inherit from the pure context
    fnCtx = RyeCtx(parentCtx); // Create the context for pure functions
  } else if (fn.inCtx && fn.ctx != null) {
    // Execute directly *within* the function's definition context (rare, usually for methods)
    fnCtx = fn.ctx!; // Assign fnCtx here
    parentCtx = null; // Explicitly null for clarity, though not strictly needed as fnCtx is used directly
  } else {
     // Normal function call: Inherit from definition context (closure) if it exists,
     // otherwise inherit from the calling context.
     parentCtx = fn.ctx ?? callCtx ?? ps.ctx;
     // Create the context for normal functions
     fnCtx = RyeCtx(parentCtx);
  }
  // Now fnCtx is guaranteed to be assigned in all branches.

  // Bind arguments from the spec to the new context
  int argIndex = 0;

  // Handle potential first argument from opword/pipeword call
  if (arg0_ != null && toLeft && fn.argsn > 0) {
     RyeObject specArg = fn.spec.series.get(argIndex)!;
     if (specArg is Word) {
       // Use setNew to define the argument in the new function context
       if (!fnCtx.setNew(specArg.index, arg0_)) {
          // This should ideally not happen if fnCtx is fresh, but handle defensively
          setError00(ps, "Failed to set argument ${specArg.print(ps.idx)} in function context");
          return;
       }
       argIndex++;
     } else {
       setError00(ps, "Function spec must contain words"); // Keep this check
       return;
     }
  }

  // Evaluate and bind remaining arguments from the calling series
  while (argIndex < fn.argsn) {
     RyeObject specArg = fn.spec.series.get(argIndex)!;
     if (specArg is Word) {
        // Evaluate the next expression in the *calling* series
        rye00_evalExpressionConcrete(ps); 
        if (ps.errorFlag || ps.failureFlag) { 
           // Pass the underlying error/failure as cause if available
           setError00(ps, "Error evaluating argument ${argIndex + 1} for function", ps.res); 
           return;
        }
        // Use setNew to define the argument in the new function context
        if (!fnCtx.setNew(specArg.index, ps.res!)) {
           setError00(ps, "Failed to set argument ${specArg.print(ps.idx)} in function context");
           return;
        }
        argIndex++;
     } else {
       setError00(ps, "Function spec must contain words"); // Keep this check
       return;
     }
  }

  // Store current state
  TSeries callerSeries = ps.ser;
  RyeCtx callerCtx = ps.ctx;
  bool callerReturnFlag = ps.returnFlag; // Store return flag state

  // Set up state for function execution
  ps.ser = fn.body.series;
  ps.ser.reset();
  ps.ctx = fnCtx;
  ps.returnFlag = false; // Reset return flag for the function body

  // Evaluate the function body
  // Use the first *evaluated* argument for potential injection if applicable, or null
  RyeObject? firstEvaluatedArg = null;
  if (fn.argsn > 0) {
     RyeObject? firstSpecArg = fn.spec.series.get(0);
     if (firstSpecArg is Word) {
       firstEvaluatedArg = fnCtx.state[firstSpecArg.index];
     }
  }
  
  rye00_evalBlockInj(ps, firstEvaluatedArg, firstEvaluatedArg != null); 

  // Execute deferred blocks *before* restoring caller state if function returned/errored
  if ((ps.returnFlag || ps.errorFlag) && ps.deferBlocks.isNotEmpty) {
     executeDeferredBlocks(ps);
  }

  // Restore caller state
  ps.ser = callerSeries;
  ps.ctx = callerCtx;
  
  // If the function body evaluation set the return flag, keep it set.
  // Otherwise, restore the caller's original return flag state.
  // This prevents a function's internal return from affecting the caller's return state
  // unless the function call itself was the last thing before a return.
  if (!ps.returnFlag) { 
     ps.returnFlag = callerReturnFlag;
  }
  // Note: ps.res holds the result from the function body evaluation.
  // If ps.returnFlag is true, this result should propagate up.
}


// Rye00_CallBuiltin calls a builtin function.
void rye00_callBuiltin(Builtin bi, ProgramState ps, RyeObject? arg0_, bool toLeft, bool pipeSecond, RyeObject? firstVal) {
  // Fast path: If all arguments are already available (curried), call directly
  if ((bi.argsn == 0) ||
      (bi.argsn == 1 && bi.cur0 != null) ||
      (bi.argsn == 2 && bi.cur0 != null && bi.cur1 != null)) {
    ps.res = bi.fn(ps, bi.cur0, bi.cur1, bi.cur2, bi.cur3, bi.cur4);
    return;
  }

  // Initialize arguments with curried values
  RyeObject? arg0 = bi.cur0;
  RyeObject? arg1 = bi.cur1;
  RyeObject? arg2 = bi.cur2;
  RyeObject? arg3 = bi.cur3;
  RyeObject? arg4 = bi.cur4;

  // Process first argument if needed
  if (bi.argsn > 0 && bi.cur0 == null) {
    // If arg0_ is provided (from opword evaluation), use it directly
    if (arg0_ != null && toLeft) { // 'toLeft' indicates it's from an opword
      arg0 = arg0_;
    } else {
      // Otherwise, evaluate the next expression from the series
      rye00_evalExpressionConcrete(ps);

      // Inline error checking for speed
      if (ps.failureFlag) {
        if (!bi.acceptFailure) {
          ps.errorFlag = true;
          return;
        }
      }

      if (ps.errorFlag || ps.returnFlag) {
        ps.res = errArg1Missing;
        return;
      }
      arg0 = ps.res;
    }
  }

  // Process second argument if needed
  // Removed incorrect duplicate block that assigned to arg0 instead of arg1
  if (bi.argsn > 1 && bi.cur1 == null) {
    rye00_evalExpressionConcrete(ps);

    // Inline error checking for speed
    if (ps.failureFlag) {
      if (!bi.acceptFailure) {
        ps.errorFlag = true;
        return;
      }
    }

    if (ps.errorFlag || ps.returnFlag) {
      ps.res = errArg2Missing;
      return;
    }

    arg1 = ps.res;
  }

  // Process third argument if needed
  if (bi.argsn > 2 && bi.cur2 == null) {
    rye00_evalExpressionConcrete(ps);

    // Inline error checking for speed
    if (ps.failureFlag) {
      if (!bi.acceptFailure) {
        ps.errorFlag = true;
        return;
      }
    }

    if (ps.errorFlag || ps.returnFlag) {
      ps.res = errArg3Missing;
      return;
    }

    arg2 = ps.res;
  }

  // Process remaining arguments with minimal error checking
  if (bi.argsn > 3 && bi.cur3 == null) {
    rye00_evalExpressionConcrete(ps);
    arg3 = ps.res;
  }

  if (bi.argsn > 4 && bi.cur4 == null) {
    rye00_evalExpressionConcrete(ps);
    arg4 = ps.res;
  }

  // Call the builtin function
  ps.res = bi.fn(ps, arg0, arg1, arg2, arg3, arg4);
}

// Rye00_checkForFailureWithBuiltin checks if there are failure flags and handles them appropriately.
bool rye00_checkForFailureWithBuiltin(Builtin bi, ProgramState ps, int n) {
  if (ps.failureFlag) {
    if (bi.acceptFailure) {
      // Accept failure
    } else {
      ps.errorFlag = true;
      return true;
    }
  }
  return false;
}

// Attempts to handle a failure by looking for and executing an 'error-handler'.
// Returns true if the failure was *not* handled (and thus should propagate or become an error),
// Returns false if the failure *was* handled by an error-handler.
bool tryHandleFailure(ProgramState ps) {
  if (ps.failureFlag && !ps.returnFlag && !ps.inErrHandler) {
    
    // Ensure the failure object (ps.res) has context and position info
    if (ps.res is Error) {
      Error failureObj = ps.res as Error;
      // Set context and position if they weren't set when the failure was initially created
      failureObj.codeContext ??= ps.ctx; 
      failureObj.position ??= ps.ser.getPos() > 0 ? ps.ser.getPos() - 1 : 0;
       // TODO: Add source file path if available (ps.ScriptPath in Go)
    } else {
       // If ps.res is not an Error object during failure, wrap it (though this shouldn't normally happen)
       ps.res = Error("Failure value was not an Error object", codeContext: ps.ctx, position: ps.ser.getPos() > 0 ? ps.ser.getPos() - 1 : 0);
    }

    // Now, check if 'error-handler' word exists
    var (errHandlerIdx, wordExists) = ps.idx.getIndex("error-handler");
    if (!wordExists) {
      ps.errorFlag = true; // Promote unhandled failure to error
      return true; // Indicate failure was not handled
    }

    // Check if 'error-handler' is defined in the context
    var (handlerObj, handlerExists) = ps.ctx.get(errHandlerIdx);
    if (!handlerExists || handlerObj is! Block) {
       ps.errorFlag = true; // Promote unhandled failure to error
       return true; // Indicate failure was not handled (handler missing or not a block)
    }

    // Execute the handler block
    Block handlerBlock = handlerObj;
    TSeries ser = ps.ser; // Store current series
    ps.ser = handlerBlock.series;
    ps.ser.reset();
    ps.inErrHandler = true; // Prevent recursive error handling

    // Evaluate the handler, injecting the failure object (current ps.res)
    rye00_evalBlockInj(ps, ps.res, true); 

    ps.inErrHandler = false; // Exit error handling mode
    ps.ser = ser; // Restore original series

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
  return injnow; // Otherwise, keep the current injnow state
}


// evalBlock evaluates a block without injection
void evalBlock(ProgramState ps) {
  rye00_evalBlockInj(ps, null, false);
}

// Rye00_EvalBlockInj evaluates a block with an optional injected value.
ProgramState rye00_evalBlockInj(ProgramState ps, RyeObject? inj, bool injnow) {
  RyeObject? currentInj = inj; // Keep track of the potentially changing injected value
  bool useInj = injnow;

  while (ps.ser.getPos() < ps.ser.len()) {
    // Use currentInj if useInj is true, otherwise evaluate normally
    if (useInj) {
      ps.res = currentInj; 
      useInj = false; // Injection used for this expression
      maybeEvalOpwordOnRight(ps); // Still check for opwords after injection
    } else {
      rye00_evalExpressionConcrete(ps);
    }

    // Handle potential failure after expression evaluation
    if (tryHandleFailure(ps)) {
      // If tryHandleFailure returns true, it means the failure was unhandled and promoted to error
      return ps; // Propagate error/failure
    }

    // If return flag was raised or an error occurred (either originally or from handler)
    if (ps.returnFlag || ps.errorFlag) {
      // Execute deferred blocks *before* returning from the block evaluation
      // if a return or error occurred.
      if (ps.deferBlocks.isNotEmpty) {
         executeDeferredBlocks(ps);
      }
      return ps; // Propagate return/error
    }
    // Removed invalid duplicate return ps; statements here

    // Check for comma and potentially reset injection flag
    useInj = maybeAcceptComma(ps, currentInj, useInj);
  }
  return ps;
}

// Implements the "_+" builtin function
RyeObject addBuiltin(ProgramState ps, RyeObject? arg0, RyeObject? arg1, RyeObject? arg2, RyeObject? arg3, RyeObject? arg4) {
  // Fast path for the most common case: Integer + Integer
  if (arg0 is Integer && arg1 is Integer) {
    return Integer(arg0.value + arg1.value);
  }
  
  // Type error for arguments
  ps.failureFlag = true;
  return Error("Arguments to _+ must be integers");
}

// Implements the "print" builtin function
RyeObject printBuiltin(ProgramState ps, RyeObject? arg0, RyeObject? arg1, RyeObject? arg2, RyeObject? arg3, RyeObject? arg4) {
  // Check if we have an argument to print
  if (arg0 != null) {
    // Print the argument
    stdout.write("${arg0.print(ps.idx)}\n");
    
    // Return the argument (print is identity function)
    return arg0;
  }
  
  // If no argument is provided, return an error
  ps.failureFlag = true;
  return Error("print requires an argument");
}

// Implements the "flutter_window" builtin function
RyeObject flutterWindowBuiltin(ProgramState ps, RyeObject? arg0, RyeObject? arg1, RyeObject? arg2, RyeObject? arg3, RyeObject? arg4) {
  // Check if we have a title argument
  if (arg0 != null) {
    String title = arg0.print(ps.idx);
    
    // Check if we have a message argument
    if (arg1 != null) {
      String message = arg1.print(ps.idx);
      
      // Launch the Flutter window in a separate isolate
      stdout.writeln("Launching Flutter window with title: '$title' and message: '$message'");
      
      // Use Future.microtask to avoid blocking the main thread
      Future.microtask(() async {
        try {
          await runFlutterApp(
            title: title,
            message: message,
          );
        } catch (e) {
          stdout.writeln("Error launching Flutter window: $e");
        }
      });
      
      // Return a void object
      return const Void();
    }
    
    // If message argument is missing
    ps.failureFlag = true;
    return Error("flutter_window requires a message argument");
  }
  
  // If title argument is missing
  ps.failureFlag = true;
  return Error("flutter_window requires a title argument");
}

// Implements the "loop" builtin function
RyeObject loopBuiltin(ProgramState ps, RyeObject? arg0, RyeObject? arg1, RyeObject? arg2, RyeObject? arg3, RyeObject? arg4) {
  // Check if the first argument is an integer (number of iterations)
  if (arg0 is Integer) {
    // Check if the second argument is a block
    if (arg1 is Block) {
      int iterations = arg0.value;
      RyeObject result = const Void();
      
      // Execute the block 'iterations' times
      for (int i = 0; i < iterations; i++) {
        // Store current series
        TSeries ser = ps.ser;
        
        // Set series to the block's series
        ps.ser = arg1.series;
        
        // Reset the series position to ensure we start from the beginning
        ps.ser.reset();
        
        // Evaluate the block, injecting the 1-based iteration number
        rye00_evalBlockInj(ps, Integer(i + 1), true); 
    
        print(ps);

        // Store the result of the iteration
        result = ps.res ?? const Void();
        
        // Check for errors or failures
        if (ps.errorFlag || ps.failureFlag) {
          return result;
        }
        
        // Restore original series
        ps.ser = ser;
      }
      print(arg0);
      return result;
    }
    
    // If second argument is not a block
    ps.failureFlag = true;
    return Error("Second argument to loop must be a block");
  }
  
  // If first argument is not an integer
  ps.failureFlag = true;
  return Error("First argument to loop must be an integer");
}

// Register all builtins
void registerBuiltins(ProgramState ps) {
  // Register core builtins from builtins.dart
  registerCoreBuiltins(ps);
  
  // Register the _+ builtin
  int plusIdx = ps.idx.indexWord("_+");
  Builtin plusBuiltin = Builtin(addBuiltin, 2, false, true, "Adds two integers");
  ps.ctx.set(plusIdx, plusBuiltin);
  
  // Register the loop builtin
  int loopIdx = ps.idx.indexWord("loop");
  Builtin loopBuiltinObj = Builtin(loopBuiltin, 2, false, false, "Executes a block a specified number of times");
  ps.ctx.set(loopIdx, loopBuiltinObj);
  
  // Register the print builtin
  int printIdx = ps.idx.indexWord("print");
  Builtin printBuiltinObj = Builtin(printBuiltin, 1, false, false, "Prints a value to the console");
  ps.ctx.set(printIdx, printBuiltinObj);
  
  // Register the flutter_window builtin
  int flutterWindowIdx = ps.idx.indexWord("flutter_window");
  Builtin flutterWindowBuiltinObj = Builtin(flutterWindowBuiltin, 2, false, false, "Shows a Flutter window with a title and message");
  ps.ctx.set(flutterWindowIdx, flutterWindowBuiltinObj);
  
  // Register builtins from other modules
  registerCollectionBuiltins(ps);
  registerStringBuiltins(ps);
  registerFlowBuiltins(ps);
  registerNumberBuiltins(ps);
  registerIterationBuiltins(ps);
  registerConditionalBuiltins(ps);
  registerPrintingBuiltins(ps);
  registerTypeBuiltins(ps);
}

// Rye00_MaybeDisplayFailureOrError displays failure or error information if present.
void rye00_maybeDisplayFailureOrError(ProgramState es, Idxs genv, String tag) {
  if (es.failureFlag) {
    stdout.write("\x1b[33mFailure\x1b[0m\n");
    stdout.write("$tag\n");
  }
  if (es.errorFlag) {
    stdout.write("\x1b[31m${es.res!.print(genv)}\n");
    stdout.write("\x1b[0m\n");
    stdout.write("$tag\n");
  }
}

// Checks if the object is an OpWord
bool isOpWord(RyeObject obj) {
  return obj is OpWord;
}

// Checks if the object is a PipeWord
bool isPipeWord(RyeObject obj) {
  return obj is PipeWord;
}

// Checks if the object is an LSetword
bool isLSetword(RyeObject obj) {
  return obj is LSetword;
}

// Checks if the object is an LModword
bool isLModword(RyeObject obj) {
  return obj is LModword;
}

// Checks if the object is a CPath
bool isCPath(RyeObject obj) {
  return obj is CPath;
}

// Checks the next token in the series. If it's an opword or similar, evaluates it using the current result.
void maybeEvalOpwordOnRight(ProgramState ps) {
  if (ps.returnFlag || ps.errorFlag) {
    return;
  }

  RyeObject? nextObj = ps.ser.peek();

  if (nextObj == null) {
    return;
  }

  // Check if the next object is an OpWord
  if (isOpWord(nextObj)) {
    ps.ser.next(); // Consume the opword

    // Find the builtin associated with the opword
    var (found, object, _) = rye00_findWordValue(ps, nextObj);

    if (found && object is Builtin) {
      // Call the builtin, passing the current result (ps.res) as the first argument (arg0_)
      // The 'true' for toLeft indicates the first arg comes from the left (ps.res)
      rye00_callBuiltin(object, ps, ps.res, true, false, null);

      // After evaluating the opword, check if another opword follows
      maybeEvalOpwordOnRight(ps);
    } else {
      setError00(ps, "Opword implementation not found: ${nextObj.print(ps.idx)}");
    }
  } else if (isPipeWord(nextObj)) {
    ps.ser.next(); // Consume the pipeword

    // Find the builtin associated with the pipeword
    var (found, object, _) = rye00_findWordValue(ps, nextObj);

    if (found && object is Builtin) {
      // Call the builtin, passing the current result (ps.res) as the first argument (arg0_)
      // The 'true' for toLeft indicates the first arg comes from the left (ps.res)
      // pipeSecond is false for now, might need adjustment later based on Go logic nuances.
      rye00_callBuiltin(object, ps, ps.res, true, false, null); 

      // After evaluating the pipeword, check if another opword/pipeword follows
      maybeEvalOpwordOnRight(ps);
    } else {
       setError00(ps, "Pipeword implementation not found: ${nextObj.print(ps.idx)}");
    }
  } else if (isLSetword(nextObj)) {
    ps.ser.next(); // Consume the LSetword
    LSetword word = nextObj as LSetword;
    int idx = word.index;
    // TODO: Add ps.AllowMod check if implemented
    bool ok = ps.ctx.setNew(idx, ps.res!); // Use setNew for LSetword
    if (!ok) {
      ps.res = Error("Can't set already set word ${ps.idx.getWord(idx)}, try using ::");
      ps.failureFlag = true;
      ps.errorFlag = true;
      return;
    }
    // After setting, check if another opword/pipeword follows
    maybeEvalOpwordOnRight(ps);
  } else if (isLModword(nextObj)) {
     ps.ser.next(); // Consume the LModword
     LModword word = nextObj as LModword;
     int idx = word.index;
     bool ok = ps.ctx.mod(idx, ps.res!); // Use mod for LModword
     if (!ok) {
       ps.res = Error("Cannot modify constant ${ps.idx.getWord(idx)}, use 'var' to declare it as a variable");
       ps.failureFlag = true;
       ps.errorFlag = true;
       return;
     }
     // After modifying, check if another opword/pipeword follows
     maybeEvalOpwordOnRight(ps);
  } else if (isCPath(nextObj)) {
    CPath path = nextObj as CPath;
    if (path.mode == 1) { // Opword-like CPath
      ps.ser.next(); // Consume CPath
      rye00_evalWord(ps, path, ps.res, false, false); // Evaluate CPath with left value
      maybeEvalOpwordOnRight(ps); // Check for more opwords
    } else if (path.mode == 2) { // Pipeword-like CPath
      ps.ser.next(); // Consume CPath
      rye00_evalWord(ps, path, ps.res, false, false); // Evaluate CPath with left value
      if (ps.returnFlag || ps.errorFlag) return; // Check flags after pipe-like CPath
      maybeEvalOpwordOnRight(ps); // Check for more opwords
    }
    // Mode 0 CPaths are handled in rye00_evalExpressionConcrete
  }
}

// Executes deferred blocks in LIFO order
void executeDeferredBlocks(ProgramState ps) {
  // Execute in reverse order (Last-In, First-Out)
  for (int i = ps.deferBlocks.length - 1; i >= 0; i--) {
    Block block = ps.deferBlocks[i];

    // Save crucial state parts that shouldn't be affected by defer
    TSeries oldSer = ps.ser;
    RyeObject? oldRes = ps.res; // Result might be overwritten, but flags matter more
    bool oldFailureFlag = ps.failureFlag;
    bool oldErrorFlag = ps.errorFlag;
    bool oldReturnFlag = ps.returnFlag; // Defer runs *before* return propagates

    // Evaluate the deferred block in the current context
    ps.ser = block.series;
    ps.ser.reset();
    // Evaluate the block without injection
    rye00_evalBlockInj(ps, null, false); 

    // Restore state, keeping any *new* error from the defer block
    ps.ser = oldSer;
    if (!ps.errorFlag) { // If defer didn't cause an error, restore original result/flags
       ps.res = oldRes;
       ps.failureFlag = oldFailureFlag;
       ps.returnFlag = oldReturnFlag;
    } else {
       // If defer *did* cause an error, keep the error state but potentially
       // restore the original return flag if the defer wasn't meant to stop it.
       // This logic might need refinement based on exact Go defer semantics on error.
       ps.returnFlag = oldReturnFlag; 
    }
  }
  // Clear the list after execution
  ps.deferBlocks.clear();
}

/* // Duplicate definition removed
// Executes deferred blocks in LIFO order
void executeDeferredBlocks(ProgramState ps) {
  // Execute in reverse order (Last-In, First-Out)
  for (int i = ps.deferBlocks.length - 1; i >= 0; i--) {
    Block block = ps.deferBlocks[i];

    // Save crucial state parts that shouldn't be affected by defer
    TSeries oldSer = ps.ser;
    RyeObject? oldRes = ps.res; // Result might be overwritten, but flags matter more
    bool oldFailureFlag = ps.failureFlag;
    bool oldErrorFlag = ps.errorFlag;
    bool oldReturnFlag = ps.returnFlag; // Defer runs *before* return propagates

    // Evaluate the deferred block in the current context
    ps.ser = block.series;
    ps.ser.reset();
    // Evaluate the block without injection
    rye00_evalBlockInj(ps, null, false); 

    // Restore state, keeping any *new* error from the defer block
    ps.ser = oldSer;
    if (!ps.errorFlag) { // If defer didn't cause an error, restore original result/flags
       ps.res = oldRes;
       ps.failureFlag = oldFailureFlag;
       ps.returnFlag = oldReturnFlag;
    } else {
       // If defer *did* cause an error, keep the error state but potentially
       // restore the original return flag if the defer wasn't meant to stop it.
       // This logic might need refinement based on exact Go defer semantics on error.
       ps.returnFlag = oldReturnFlag; 
    }
  }
  // Clear the list after execution
  ps.deferBlocks.clear();
}
*/
