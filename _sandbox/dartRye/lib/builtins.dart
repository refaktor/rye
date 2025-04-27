// builtins.dart - Core builtins for the Dart implementation of Rye

import 'rye.dart';
import 'types.dart';
import 'builtins_printing.dart';
import 'builtins_types.dart';

// --- Function Creation Builtins ---

// Implements the "does" builtin function
RyeObject doesBuiltin(ProgramState ps, RyeObject? arg0, RyeObject? arg1, RyeObject? arg2, RyeObject? arg3, RyeObject? arg4) {
  if (arg0 is Block) {
    // Create a function with no parameters
    return RyeFunction(Block(TSeries([])), arg0, ps.ctx);
  }
  ps.failureFlag = true;
  return Error("does expects a block argument");
}

// Implements the "fn" builtin function
RyeObject fnBuiltin(ProgramState ps, RyeObject? arg0, RyeObject? arg1, RyeObject? arg2, RyeObject? arg3, RyeObject? arg4) {
  if (arg0 is Block && arg1 is Block) {
    // Process function spec (parameter block)
    // In a full implementation, we would validate the spec block here
    
    // Create a function with the given parameters and body
    return RyeFunction(arg0, arg1, ps.ctx);
  }
  ps.failureFlag = true;
  return Error("fn expects two block arguments: parameters and body");
}

// Implements the "fn1" builtin function
RyeObject fn1Builtin(ProgramState ps, RyeObject? arg0, RyeObject? arg1, RyeObject? arg2, RyeObject? arg3, RyeObject? arg4) {
  if (arg0 is Block) {
    // Create a function with one anonymous parameter
    List<RyeObject> spec = [Word(1)]; // Using index 1 for the anonymous parameter
    return RyeFunction(Block(TSeries(spec)), arg0, ps.ctx);
  }
  ps.failureFlag = true;
  return Error("fn1 expects a block argument");
}

// --- Variable Builtins ---

// Implements the "var" builtin function
RyeObject varBuiltin(ProgramState ps, RyeObject? arg0, RyeObject? arg1, RyeObject? arg2, RyeObject? arg3, RyeObject? arg4) {
  if (arg0 is Word && arg1 != null) {
    // Check if word already exists in context
    if (ps.ctx.has(arg0.index)) {
      ps.failureFlag = true;
      return Error("Cannot redefine existing word '${ps.idx.getWord(arg0.index)}' with var");
    }
    
    // Set the value and mark as variable
    ps.ctx.set(arg0.index, arg1);
    ps.ctx.markAsVariable(arg0.index);
    return arg1;
  // Removed Tagword case as it's not defined
  }
  ps.failureFlag = true;
  return Error("var expects a word/tagword and a value");
}

// Implements the "val" builtin function
RyeObject valBuiltin(ProgramState ps, RyeObject? arg0, RyeObject? arg1, RyeObject? arg2, RyeObject? arg3, RyeObject? arg4) {
  if (arg0 is Word) {
    var (value, exists) = ps.ctx.get(arg0.index);
    if (value != null) {
      return value;
    }
    ps.failureFlag = true;
    return Error("Word not found in context");
  }
  ps.failureFlag = true;
  return Error("val expects a word argument");
}

// --- Context Manipulation Builtins ---


// Implements the "with" builtin function
RyeObject withBuiltin(ProgramState ps, RyeObject? arg0, RyeObject? arg1, RyeObject? arg2, RyeObject? arg3, RyeObject? arg4) {
  if (arg0 != null && arg1 is Block) {
    // Store current series
    TSeries ser = ps.ser;
    
    // Set series to the block's series
    ps.ser = arg1.series;
    
    // Evaluate the block with the value injected
    rye00_evalBlockInj(ps, arg0, true);
    
    // Restore original series
    ps.ser = ser;
    
    // Return the result of the evaluation
    return ps.res!;
  }
  ps.failureFlag = true;
  return Error("with expects a value and a block");
}

// --- Flow Control Builtins ---

// Implements the "return" builtin function
RyeObject returnBuiltin(ProgramState ps, RyeObject? arg0, RyeObject? arg1, RyeObject? arg2, RyeObject? arg3, RyeObject? arg4) {
  if (arg0 != null) {
    ps.returnFlag = true;
    ps.res = arg0;
    return arg0;
  }
  ps.failureFlag = true;
  return Error("return expects an argument");
}

// --- Dictionary Builtins ---

// Implements the "dict" builtin function
RyeObject dictBuiltin(ProgramState ps, RyeObject? arg0, RyeObject? arg1, RyeObject? arg2, RyeObject? arg3, RyeObject? arg4) {
  if (arg0 is Block) {
    // Create a new dictionary from the block
    Map<String, RyeObject> entries = {};
    
    // Process pairs of key-value in the block
    for (int i = 0; i < arg0.series.len() - 1; i += 2) {
      RyeObject? key = arg0.series.get(i);
      RyeObject? value = arg0.series.get(i + 1);
      
      if (key is RyeString && value != null) {
        entries[key.value] = value;
      } else {
        ps.failureFlag = true;
        return Error("dict expects string keys in the block");
      }
    }
    
    return RyeDict(entries);
  }
  ps.failureFlag = true;
  return Error("dict expects a block argument");
}

// Implements the "list" builtin function
RyeObject listBuiltin(ProgramState ps, RyeObject? arg0, RyeObject? arg1, RyeObject? arg2, RyeObject? arg3, RyeObject? arg4) {
  if (arg0 is Block) {
    // Create a new list from the block
    List<RyeObject> items = [];
    
    // Add all items from the block to the list
    for (int i = 0; i < arg0.series.len(); i++) {
      RyeObject? item = arg0.series.get(i);
      if (item != null) {
        items.add(item);
      }
    }
    
    return RyeList(items);
  }
  ps.failureFlag = true;
  return Error("list expects a block argument");
}

// Implements the "change!" builtin function
RyeObject changeBuiltin(ProgramState ps, RyeObject? arg0, RyeObject? arg1, RyeObject? arg2, RyeObject? arg3, RyeObject? arg4) {
  if (arg0 != null && arg1 is Word) {
    // Get the current value of the word
    var (oldValue, exists) = ps.ctx.get(arg1.index);
    if (oldValue == null) {
      ps.failureFlag = true;
      return Error("Word not found in context");
    }
    
    // Check if the word is a variable
    if (!ps.ctx.isVariable(arg1.index)) {
      ps.failureFlag = true;
      return Error("Cannot modify constant '${ps.idx.getWord(arg1.index)}', use 'var' to declare it as a variable");
    }
    
    // Modify the word's value
    ps.ctx.set(arg1.index, arg0);
    
    // Return 1 if the value changed, 0 if it's the same
    if (oldValue.getKind() == arg0.getKind() && oldValue.inspect(ps.idx) == arg0.inspect(ps.idx)) {
      return Integer(0);
    } else {
      return Integer(1);
    }
  }
  ps.failureFlag = true;
  return Error("change! expects a value and a word");
}

// Implements the "set!" builtin function
RyeObject setBuiltin(ProgramState ps, RyeObject? arg0, RyeObject? arg1, RyeObject? arg2, RyeObject? arg3, RyeObject? arg4) {
  if (arg0 != null && arg1 is Word) {
    // Check if the word is a variable
    if (!ps.ctx.isVariable(arg1.index)) {
      ps.failureFlag = true;
      return Error("Cannot modify constant '${ps.idx.getWord(arg1.index)}', use 'var' to declare it as a variable");
    }
    
    // Set the word's value
    ps.ctx.set(arg1.index, arg0);
    return arg0;
  } else if (arg0 is Block && arg1 is Block) {
    // Handle destructuring assignment
    for (int i = 0; i < arg1.series.len(); i++) {
      RyeObject? word = arg1.series.get(i);
      if (word is Word) {
        // Check if we have enough values
        if (i >= arg0.series.len()) {
          ps.failureFlag = true;
          return Error("More words than values in set!");
        }
        
        RyeObject? value = arg0.series.get(i);
        if (value != null) {
          // Check if the word is a variable
          if (!ps.ctx.isVariable(word.index)) {
            ps.failureFlag = true;
            return Error("Cannot modify constant '${ps.idx.getWord(word.index)}', use 'var' to declare it as a variable");
          }
          
          // Set the word's value
          ps.ctx.set(word.index, value);
        }
      } else {
        ps.failureFlag = true;
        return Error("Only words allowed in words block for set!");
      }
    }
    return arg0;
  }
  ps.failureFlag = true;
  return Error("set! expects a value and a word, or two blocks");
}

// Implements the "unset!" builtin function
RyeObject unsetBuiltin(ProgramState ps, RyeObject? arg0, RyeObject? arg1, RyeObject? arg2, RyeObject? arg3, RyeObject? arg4) {
  if (arg0 is Word) {
    // Remove the word from the context
    bool success = ps.ctx.unset(arg0.index);
    if (success) {
      return Void();
    } else {
      ps.failureFlag = true;
      return Error("Word not found in context");
    }
  }
  ps.failureFlag = true;
  return Error("unset! expects a word argument");
}

// --- Registration ---

void registerCoreBuiltins(ProgramState ps) {
  // Register function creation builtins
  ps.ctx.set(ps.idx.indexWord("does"), 
    Builtin(doesBuiltin, 1, false, true, "Creates a function with no arguments that executes the given block when called"));
  
  ps.ctx.set(ps.idx.indexWord("fn"), 
    Builtin(fnBuiltin, 2, false, true, "Creates a function with named parameters specified in the first block and code in the second block"));
  
  ps.ctx.set(ps.idx.indexWord("fn1"), 
    Builtin(fn1Builtin, 1, false, true, "Creates a function that accepts one anonymous argument and executes the given block with that argument"));
  
  // Register variable builtins
  ps.ctx.set(ps.idx.indexWord("var"), 
    Builtin(varBuiltin, 2, false, false, "Declares a word as a variable with the given value, allowing it to be modified"));
  
  ps.ctx.set(ps.idx.indexWord("val"), 
    Builtin(valBuiltin, 1, false, true, "Returns value of the word in context"));
  
  ps.ctx.set(ps.idx.indexWord("change!"), 
    Builtin(changeBuiltin, 2, false, false, "Changes the value of a variable, returns 1 if value changed, 0 otherwise"));
  
  ps.ctx.set(ps.idx.indexWord("set!"), 
    Builtin(setBuiltin, 2, false, false, "Set word to value or words by deconstructing a block"));
  
  ps.ctx.set(ps.idx.indexWord("unset!"), 
    Builtin(unsetBuiltin, 1, false, false, "Unset a word in current context"));
  
  ps.ctx.set(ps.idx.indexWord("with"), 
    Builtin(withBuiltin, 2, true, false, "Takes a value and a block of code. It does the code with the value injected"));
  
  // Register flow control builtins
  ps.ctx.set(ps.idx.indexWord("return"), 
    Builtin(returnBuiltin, 1, false, false, "Accepts one value and returns it"));
  
  // Register dictionary builtins
  ps.ctx.set(ps.idx.indexWord("dict"), 
    Builtin(dictBuiltin, 1, false, true, "Constructs a Dict from the Block of key and value pairs"));
  
  ps.ctx.set(ps.idx.indexWord("list"), 
    Builtin(listBuiltin, 1, false, true, "Constructs a List from the Block of values"));
}
