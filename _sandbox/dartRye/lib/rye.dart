// rye.dart - A simplified Rye language evaluator in Dart

import 'dart:io';
import 'dart:async';
import 'flutter/flutter_app.dart';

// Type enum for Rye values
enum RyeType {
  blockType,
  integerType,
  wordType,
  builtinType,
  errorType,
  voidType,
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

  Error(this.message, {this.status = 0, this.parent});

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

  ProgramState(this.ser, this.idx)
      : ctx = RyeCtx(null),
        pCtx = RyeCtx(null),
        args = List.filled(6, 0);
}

// Pre-allocated common error messages to avoid allocations
final Error errMissingValue = Error("Expected Rye value but it's missing");
final Error errExpressionGuard = Error("Expression guard inside expression");
final Error errErrorObject = Error("Error object encountered");
final Error errUnsupportedType = Error("Unsupported type in simplified interpreter");
final Error errArg1Missing = Error("Argument 1 missing for builtin");
final Error errArg2Missing = Error("Argument 2 missing for builtin");
final Error errArg3Missing = Error("Argument 3 missing for builtin");

// Helper function to set error state - uses the shared error variables
void setError00(ProgramState ps, String message) {
  ps.errorFlag = true;

  // Use pre-allocated errors for common messages
  switch (message) {
    case "Expected Rye value but it's missing":
      ps.res = errMissingValue;
      break;
    case "Expression guard inside expression":
      ps.res = errExpressionGuard;
      break;
    case "Error object encountered":
      ps.res = errErrorObject;
      break;
    case "Unsupported type in simplified interpreter":
      ps.res = errUnsupportedType;
      break;
    default:
      ps.res = Error(message);
      break;
  }
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

  switch (object.type()) {
    case RyeType.integerType:
      ps.res = object;
      break;
    case RyeType.blockType:
      ps.res = object;
      break;
    case RyeType.wordType:
      rye00_evalWord(ps, object as Word, null, false, false);
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

// Rye00_EvalWord evaluates a word in the current context.
void rye00_evalWord(ProgramState ps, RyeObject word, RyeObject? leftVal, bool toLeft, bool pipeSecond) {
  var (found, object, session) = rye00_findWordValue(ps, word);

  if (found) {
    rye00_evalObject(ps, object!, leftVal, toLeft, session, pipeSecond, null);
  } else {
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
    default:
      ps.res = object;
  }
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
    // Direct call to avoid function pointer indirection
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

  // Process second argument if needed
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

// Rye00_checkFlagsAfterExpression checks if there are failure flags after evaluating a block.
bool rye00_checkFlagsAfterExpression(ProgramState ps) {
  if ((ps.failureFlag && !ps.returnFlag) || ps.errorFlag) {
    ps.errorFlag = true;
    return true;
  }
  return false;
}

// Rye00_EvalBlockInj evaluates a block with an optional injected value.
ProgramState rye00_evalBlockInj(ProgramState ps, RyeObject? inj, bool injnow) {
  while (ps.ser.getPos() < ps.ser.len()) {
    rye00_evalExpressionConcrete(ps);

    if (rye00_checkFlagsAfterExpression(ps)) {
      return ps;
    }
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
        // Create a new program state for each iteration with a fresh copy of the block's series
        ProgramState blockPs = ProgramState(TSeries(List<RyeObject>.from(arg1.series.s)), ps.idx);
        blockPs.ctx = ps.ctx;
        
        // Reset the series position to ensure we start from the beginning
        blockPs.ser.reset();
        
        // Evaluate the block
        rye00_evalBlockInj(blockPs, null, false);
        
        // Check for errors or failures
        if (blockPs.errorFlag || blockPs.failureFlag) {
          ps.errorFlag = blockPs.errorFlag;
          ps.failureFlag = blockPs.failureFlag;
          ps.res = blockPs.res;
          return blockPs.res ?? Error("Error in loop");
        }
        
        // Store the result of the last iteration
        result = blockPs.res ?? const Void();
      }
      
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

// Register the "_+", "loop", "print", and "flutter_window" builtin functions
void registerBuiltins(ProgramState ps) {
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
