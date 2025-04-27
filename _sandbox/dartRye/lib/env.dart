// env.dart - Environment and Program State for Dart Rye evaluator

import 'types.dart' show RyeObject, Word, Error, Block, TSeries; // Import specific types needed
import 'idxs.dart';

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
      // In Go, this returns an error object. Replicating that.
      return Error("Can't set already set word, try using modword! FIXME !"); 
    } else {
      state[word] = val;
      return val; // Return the value that was set
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
    
    // Check parent contexts recursively only if the word isn't defined here
    if (parent != null && !state.containsKey(word)) { 
      return parent!.isVariable(word);
    }
    
    // Default to false if not found or defined here but not marked
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
      // Check if the word is marked as a variable *anywhere* in the hierarchy
      if (isVariable(word)) { 
        state[word] = val;
        return true;
      }
      return false; // Word exists here but is not a variable
    }
    
    // Check parent contexts recursively
    if (parent != null) {
      return parent!.mod(word, val);
    }
    
    // Word not found in hierarchy
    return false;
  }
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
      : ctx = RyeCtx(null), // Root context has null parent
        pCtx = RyeCtx(null), // Root pure context has null parent
        args = List.filled(6, 0); // Assuming max 6 args for now

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

  // Helper to set an error state
  void setError00(String message, [RyeObject? value]) {
    errorFlag = true;
    // Wrap the message in an Error object if it's not already one
    if (value is Error) {
      res = value;
      // Optionally enhance the error with context/position if missing
      value.codeContext ??= ctx; 
      value.position ??= ser.getPos() > 0 ? ser.getPos() - 1 : 0;
    } else {
      res = Error(message, 
                  codeContext: ctx, 
                  position: ser.getPos() > 0 ? ser.getPos() - 1 : 0,
                  parent: value is Error ? value : null // Attach original value if it was an error
                 );
    }
  }
}
