// builtins_types.dart - Type-related builtins for the Dart implementation of Rye

import 'env.dart' show ProgramState; // Import ProgramState
// Import specific types needed from types.dart
import 'types.dart' show RyeObject, RyeString, Integer, Decimal, Error, RyeList, Block, TSeries, Word, Builtin, Uri; 

// --- Type Conversion Builtins ---

// Implements the "to-integer" builtin function
RyeObject toIntegerBuiltin(ProgramState ps, RyeObject? arg0, RyeObject? arg1, RyeObject? arg2, RyeObject? arg3, RyeObject? arg4) {
  if (arg0 != null) {
    if (arg0 is RyeString) {
      try {
        int value = int.parse(arg0.value);
        return Integer(value);
      } catch (e) {
        ps.failureFlag = true;
        return Error("Cannot convert string to integer: ${e.toString()}");
      }
    } else if (arg0 is Decimal) {
      return Integer(arg0.value.toInt());
    }
  }
  ps.failureFlag = true;
  return Error("to-integer expects a string or decimal argument");
}

// Implements the "to-decimal" builtin function
RyeObject toDecimalBuiltin(ProgramState ps, RyeObject? arg0, RyeObject? arg1, RyeObject? arg2, RyeObject? arg3, RyeObject? arg4) {
  if (arg0 != null) {
    if (arg0 is RyeString) {
      try {
        double value = double.parse(arg0.value);
        return Decimal(value);
      } catch (e) {
        ps.failureFlag = true;
        return Error("Cannot convert string to decimal: ${e.toString()}");
      }
    } else if (arg0 is Integer) {
      return Decimal(arg0.value.toDouble());
    }
  }
  ps.failureFlag = true;
  return Error("to-decimal expects a string or integer argument");
}

// Implements the "to-string" builtin function
RyeObject toStringBuiltin(ProgramState ps, RyeObject? arg0, RyeObject? arg1, RyeObject? arg2, RyeObject? arg3, RyeObject? arg4) {
  if (arg0 != null) {
    return RyeString(arg0.print(ps.idx));
  }
  ps.failureFlag = true;
  return Error("to-string requires an argument");
}

// Implements the "to-char" builtin function
RyeObject toCharBuiltin(ProgramState ps, RyeObject? arg0, RyeObject? arg1, RyeObject? arg2, RyeObject? arg3, RyeObject? arg4) {
  if (arg0 is Integer) {
    return RyeString(String.fromCharCode(arg0.value));
  }
  ps.failureFlag = true;
  return Error("to-char expects an integer argument");
}

// Implements the "to-block" builtin function
RyeObject toBlockBuiltin(ProgramState ps, RyeObject? arg0, RyeObject? arg1, RyeObject? arg2, RyeObject? arg3, RyeObject? arg4) {
  if (arg0 is RyeList) {
    // Convert list to block
    List<RyeObject?> items = []; // Allow nulls
    for (RyeObject? item in arg0.value) { // Use .value
      items.add(item); // Add item (could be null)
    }
    return Block(TSeries(items));
  }
  ps.failureFlag = true;
  return Error("to-block expects a list argument");
}

// Implements the "to-word" builtin function
RyeObject toWordBuiltin(ProgramState ps, RyeObject? arg0, RyeObject? arg1, RyeObject? arg2, RyeObject? arg3, RyeObject? arg4) {
  if (arg0 is RyeString) {
    int idx = ps.idx.indexWord(arg0.value);
    return Word(idx);
  } else if (arg0 is Word) {
    // For word-like objects, just return a new Word with the same index
    int idx = arg0.index;
    return Word(idx);
  }
  ps.failureFlag = true;
  return Error("to-word expects a string or word-like argument");
}

// --- Type Testing Builtins ---

// Implements the "is-string" builtin function
RyeObject isStringBuiltin(ProgramState ps, RyeObject? arg0, RyeObject? arg1, RyeObject? arg2, RyeObject? arg3, RyeObject? arg4) {
  if (arg0 != null) {
    return Integer(arg0 is RyeString ? 1 : 0);
  }
  ps.failureFlag = true;
  return Error("is-string requires an argument");
}

// Implements the "is-integer" builtin function
RyeObject isIntegerBuiltin(ProgramState ps, RyeObject? arg0, RyeObject? arg1, RyeObject? arg2, RyeObject? arg3, RyeObject? arg4) {
  if (arg0 != null) {
    return Integer(arg0 is Integer ? 1 : 0);
  }
  ps.failureFlag = true;
  return Error("is-integer requires an argument");
}

// Implements the "is-decimal" builtin function
RyeObject isDecimalBuiltin(ProgramState ps, RyeObject? arg0, RyeObject? arg1, RyeObject? arg2, RyeObject? arg3, RyeObject? arg4) {
  if (arg0 != null) {
    return Integer(arg0 is Decimal ? 1 : 0);
  }
  ps.failureFlag = true;
  return Error("is-decimal requires an argument");
}

// Implements the "is-number" builtin function
RyeObject isNumberBuiltin(ProgramState ps, RyeObject? arg0, RyeObject? arg1, RyeObject? arg2, RyeObject? arg3, RyeObject? arg4) {
  if (arg0 != null) {
    return Integer((arg0 is Integer || arg0 is Decimal) ? 1 : 0);
  }
  ps.failureFlag = true;
  return Error("is-number requires an argument");
}

// Implements the "type?" builtin function
RyeObject typeBuiltin(ProgramState ps, RyeObject? arg0, RyeObject? arg1, RyeObject? arg2, RyeObject? arg3, RyeObject? arg4) {
  if (arg0 != null) {
    return Word(arg0.type().index);
  }
  ps.failureFlag = true;
  return Error("type? requires an argument");
}

// Implements the "kind?" builtin function
RyeObject kindBuiltin(ProgramState ps, RyeObject? arg0, RyeObject? arg1, RyeObject? arg2, RyeObject? arg3, RyeObject? arg4) {
  if (arg0 != null) {
    return Word(arg0.getKind());
  }
  ps.failureFlag = true;
  return Error("kind? requires an argument");
}

// Implements the "types?" builtin function
RyeObject typesBuiltin(ProgramState ps, RyeObject? arg0, RyeObject? arg1, RyeObject? arg2, RyeObject? arg3, RyeObject? arg4) {
  if (arg0 is Block) {
    List<RyeObject> types = [];
    for (int i = 0; i < arg0.series.len(); i++) {
      RyeObject? item = arg0.series.get(i);
      if (item != null) {
        types.add(Word(item.type().index));
      }
    }
    return Block(TSeries(types));
  }
  ps.failureFlag = true;
  return Error("types? expects a block argument");
}

// Implements the "dump" builtin function
RyeObject dumpBuiltin(ProgramState ps, RyeObject? arg0, RyeObject? arg1, RyeObject? arg2, RyeObject? arg3, RyeObject? arg4) {
  if (arg0 != null) {
    // For now, just use print as a simple implementation
    return RyeString(arg0.print(ps.idx));
  }
  ps.failureFlag = true;
  return Error("dump requires an argument");
}

// Implements the "mold" builtin function
RyeObject moldBuiltin(ProgramState ps, RyeObject? arg0, RyeObject? arg1, RyeObject? arg2, RyeObject? arg3, RyeObject? arg4) {
  if (arg0 != null) {
    // For now, just use print as a simple implementation
    return RyeString(arg0.print(ps.idx));
  }
  ps.failureFlag = true;
  return Error("mold requires an argument");
}

// --- Registration ---

void registerTypeBuiltins(ProgramState ps) {
  // Register type conversion builtins
  ps.ctx.set(ps.idx.indexWord("to-integer"), 
    Builtin(toIntegerBuiltin, 1, false, true, "Tries to change a Rye value (like string) to integer"));
  
  ps.ctx.set(ps.idx.indexWord("to-decimal"), 
    Builtin(toDecimalBuiltin, 1, false, true, "Tries to change a Rye value (like string) to decimal"));
  
  ps.ctx.set(ps.idx.indexWord("to-string"), 
    Builtin(toStringBuiltin, 1, false, true, "Tries to turn a Rye value to string"));
  
  ps.ctx.set(ps.idx.indexWord("to-char"), 
    Builtin(toCharBuiltin, 1, false, true, "Tries to turn a Rye value (like integer) to ascii character"));
  
  ps.ctx.set(ps.idx.indexWord("to-block"), 
    Builtin(toBlockBuiltin, 1, false, true, "Turns a List to a Block"));
  
  ps.ctx.set(ps.idx.indexWord("to-word"), 
    Builtin(toWordBuiltin, 1, false, true, "Tries to change a Rye value to a word with same name"));
  
  // Register type testing builtins
  ps.ctx.set(ps.idx.indexWord("is-string"), 
    Builtin(isStringBuiltin, 1, false, true, "Returns true if value is a string"));
  
  ps.ctx.set(ps.idx.indexWord("is-integer"), 
    Builtin(isIntegerBuiltin, 1, false, true, "Returns true if value is an integer"));
  
  ps.ctx.set(ps.idx.indexWord("is-decimal"), 
    Builtin(isDecimalBuiltin, 1, false, true, "Returns true if value is a decimal"));
  
  ps.ctx.set(ps.idx.indexWord("is-number"), 
    Builtin(isNumberBuiltin, 1, false, true, "Returns true if value is a number (integer or decimal)"));
  
  ps.ctx.set(ps.idx.indexWord("type?"), 
    Builtin(typeBuiltin, 1, false, true, "Returns the type of Rye value as a word"));
  
  ps.ctx.set(ps.idx.indexWord("kind?"), 
    Builtin(kindBuiltin, 1, false, true, "Returns the kind of Rye value as a word"));
  
  ps.ctx.set(ps.idx.indexWord("types?"), 
    Builtin(typesBuiltin, 1, false, true, "Returns the types of Rye values in a block as a block of words"));
  
  ps.ctx.set(ps.idx.indexWord("dump"), 
    Builtin(dumpBuiltin, 1, false, true, "Returns (dumps) Rye code representing the object"));
  
  ps.ctx.set(ps.idx.indexWord("mold"), 
    Builtin(moldBuiltin, 1, false, true, "Turn value to it's string representation"));
}
