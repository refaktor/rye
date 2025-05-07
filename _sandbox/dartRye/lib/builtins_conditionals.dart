// builtins_conditionals.dart - Conditional builtins for the Dart implementation of Rye

import 'env.dart' show ProgramState; // Import ProgramState
// Import specific types needed from types.dart
import 'types.dart' show RyeObject, Boolean, Integer, RyeString, Block, RyeList, RyeDict, Error, Void, TSeries, Builtin; 
import 'evaldo.dart' show evalBlockInj; // Import the correct evaluator function

// Helper function to check if a value is truthy
bool isTruthy(RyeObject obj) {
  if (obj is Boolean) {
    return obj.value;
  } else if (obj is Integer) {
    return obj.value != 0;
  } else if (obj is RyeString) {
    return obj.value.isNotEmpty;
  } else if (obj is Block) {
    return obj.series.len() > 0;
  } else if (obj is RyeList) {
    return obj.value.isNotEmpty; // Use .value
  } else if (obj is RyeDict) {
    return obj.value.isNotEmpty; // Use .value (assuming this was also intended)
  }
  
  return false;
}

// Implements the "if" builtin function
RyeObject ifBuiltin(ProgramState ps, RyeObject? arg0, RyeObject? arg1, RyeObject? arg2, RyeObject? arg3, RyeObject? arg4) {
  if (arg0 == null || arg1 == null) {
    ps.failureFlag = true;
    return Error("if requires two arguments");
  }
  
  // Check if the first argument is a boolean
  if (arg0 is Boolean) {
    // Check if the second argument is a block
    if (arg1 is Block) {
      // If condition is true, execute the block
      if (arg0.value) {
        // Store current series
        TSeries ser = ps.ser;
        
        // Set series to the block's series
        ps.ser = arg1.series;
        
        // Reset the series position to ensure we start from the beginning
        ps.ser.reset();
        
        // Evaluate the block with the condition value injected
        evalBlockInj(ps, arg0, true); // Use evalBlockInj
        
        // Restore original series
        ps.ser = ser;
        
        // Return the result of the block execution
        return ps.res!;
      }
      
      // If condition is false, return false
      return Boolean(false);
    }
    
    // If second argument is not a block, return an error
    ps.failureFlag = true;
    return Error("if expects a block as its second argument");
  }
  
  // If first argument is not a boolean, return an error
  ps.failureFlag = true;
  return Error("if expects a boolean as its first argument");
}

// Implements the "either" builtin function (if/else)
RyeObject eitherBuiltin(ProgramState ps, RyeObject? arg0, RyeObject? arg1, RyeObject? arg2, RyeObject? arg3, RyeObject? arg4) {
  if (arg0 == null || arg1 == null || arg2 == null) {
    ps.failureFlag = true;
    return Error("either requires three arguments");
  }
  
  // Check if the first argument is a boolean
  if (arg0 is Boolean) {
    // Check if both the second and third arguments are blocks
    if (arg1 is Block && arg2 is Block) {
      // Store current series
      TSeries ser = ps.ser;
      
      // Choose which block to execute based on the condition
      if (arg0.value) {
        ps.ser = arg1.series;
      } else {
        ps.ser = arg2.series;
      }
      
      // Reset the series position to ensure we start from the beginning
      ps.ser.reset();
      
      // Evaluate the chosen block with the condition value injected
      evalBlockInj(ps, arg0, true); // Use evalBlockInj
      
      // Restore original series
      ps.ser = ser;
      
      // Return the result of the block execution
      return ps.res!;
    } else if (arg1 is! Block && arg2 is! Block) {
      // If neither argument is a block, treat them as literal values
      return arg0.value ? arg1 : arg2;
    }
    
    // If one argument is a block and the other isn't, return an error
    ps.failureFlag = true;
    return Error("either expects both the second and third arguments to be of the same type (either both blocks or both values)");
  }
  
  // If first argument is not a boolean, return an error
  ps.failureFlag = true;
  return Error("either expects a boolean as its first argument");
}

// Implements the "switch" builtin function
RyeObject switchBuiltin(ProgramState ps, RyeObject? arg0, RyeObject? arg1, RyeObject? arg2, RyeObject? arg3, RyeObject? arg4) {
  if (arg0 == null || arg1 == null) {
    ps.failureFlag = true;
    return Error("switch requires two arguments");
  }
  
  // Check if the second argument is a block
  if (arg1 is Block) {
    // The value to match against
    RyeObject value = arg0;
    
    // Flag to track if any case matched
    bool anyFound = false;
    
    // The block to execute if a match is found
    RyeObject? codeToExecute;
    
    // Iterate through the case-handler pairs in the block
    for (int i = 0; i < arg1.series.len(); i += 2) {
      // Check if we have enough elements for a case-handler pair
      if (i + 1 >= arg1.series.len()) {
        ps.failureFlag = true;
        return Error("switch: malformed switch block (odd number of elements)");
      }
      
      // Get the case value and handler
      RyeObject? caseValue = arg1.series.get(i);
      RyeObject? handler = arg1.series.get(i + 1);
      
      if (caseValue == null || handler == null) {
        ps.failureFlag = true;
        return Error("switch: malformed switch block (null elements)");
      }
      
      // Check if this is a default case (Void)
      if (caseValue is Void) {
        if (!anyFound) {
          codeToExecute = handler;
          anyFound = true;
        }
        continue;
      }
      
      // Check if the case value matches the input value
      if (value.equal(caseValue)) {
        codeToExecute = handler;
        anyFound = true;
        break;
      }
    }
    
    // If a match was found, execute the corresponding handler
    if (anyFound && codeToExecute != null) {
      if (codeToExecute is Block) {
        // Store current series
        TSeries ser = ps.ser;
        
        // Set series to the handler block's series
        ps.ser = codeToExecute.series;
        
        // Reset the series position to ensure we start from the beginning
        ps.ser.reset();
        
        // Evaluate the handler block with the value injected
        evalBlockInj(ps, value, true); // Use evalBlockInj
        
        // Restore original series
        ps.ser = ser;
        
        // Return the result of the handler execution
        return ps.res!;
      } else {
        ps.failureFlag = true;
        return Error("switch: handler must be a block");
      }
    }
    
    // If no match was found, return the original value
    return value;
  }
  
  // If second argument is not a block, return an error
  ps.failureFlag = true;
  return Error("switch expects a block as its second argument");
}

// Implements the "when" builtin function
RyeObject whenBuiltin(ProgramState ps, RyeObject? arg0, RyeObject? arg1, RyeObject? arg2, RyeObject? arg3, RyeObject? arg4) {
  if (arg0 == null || arg1 == null || arg2 == null) {
    ps.failureFlag = true;
    return Error("when requires three arguments");
  }
  
  // Check if the second and third arguments are blocks
  if (arg1 is Block && arg2 is Block) {
    // Store current series
    TSeries ser = ps.ser;
    
    // Set series to the condition block's series
    ps.ser = arg1.series;
    
    // Reset the series position to ensure we start from the beginning
    ps.ser.reset();
    
    // Evaluate the condition block with the value injected
    evalBlockInj(ps, arg0, true); // Use evalBlockInj
    
    // Check if the condition is truthy
    if (isTruthy(ps.res!)) {
      // Set series to the action block's series
      ps.ser = arg2.series;
      
      // Reset the series position to ensure we start from the beginning
      ps.ser.reset();
      
      // Evaluate the action block with the value injected
      evalBlockInj(ps, arg0, true); // Use evalBlockInj
    }
    
    // Restore original series
    ps.ser = ser;
    
    // Return the original value regardless of whether the action was executed
    return arg0;
  }
  
  // If second or third argument is not a block, return an error
  ps.failureFlag = true;
  return Error("when expects blocks as its second and third arguments");
}

// Implements the "while" builtin function
RyeObject whileBuiltin(ProgramState ps, RyeObject? arg0, RyeObject? arg1, RyeObject? arg2, RyeObject? arg3, RyeObject? arg4) {
  if (arg0 == null || arg1 == null) {
    ps.failureFlag = true;
    return Error("while requires two arguments");
  }
  
  // Check if both arguments are blocks
  if (arg0 is Block && arg1 is Block) {
    // Store current series
    TSeries ser = ps.ser;
    
    // Loop until the condition becomes false
    while (true) {
      // Evaluate the condition block
      ps.ser = arg0.series;
      ps.ser.reset();
      evalBlockInj(ps, null, false); // Use evalBlockInj
      
      // Check for errors
      if (ps.errorFlag) {
        ps.ser = ser;
        return ps.res!;
      }
      
      // Check if the condition is truthy
      if (!isTruthy(ps.res!)) {
        break;
      }
      
      // Evaluate the body block
      ps.ser = arg1.series;
      ps.ser.reset();
      evalBlockInj(ps, null, false); // Use evalBlockInj
      
      // Check for errors or return flag
      if (ps.errorFlag || ps.returnFlag) {
        ps.ser = ser;
        return ps.res!;
      }
    }
    
    // Restore original series
    ps.ser = ser;
    
    // Return the result of the last evaluation
    return ps.res!;
  }
  
  // If either argument is not a block, return an error
  ps.failureFlag = true;
  return Error("while expects blocks as both arguments");
}

// Register the conditional builtins
void registerConditionalBuiltins(ProgramState ps) {
  // Register the if builtin
  int ifIdx = ps.idx.indexWord("if");
  Builtin ifBuiltinObj = Builtin(ifBuiltin, 2, false, true, "Executes a block of code only if the condition is true");
  ps.ctx.set(ifIdx, ifBuiltinObj);
  
  // Register the either builtin
  int eitherIdx = ps.idx.indexWord("either");
  Builtin eitherBuiltinObj = Builtin(eitherBuiltin, 3, false, true, "Executes one of two blocks based on a boolean condition");
  ps.ctx.set(eitherIdx, eitherBuiltinObj);
  
  // Register the switch builtin
  int switchIdx = ps.idx.indexWord("switch");
  Builtin switchBuiltinObj = Builtin(switchBuiltin, 2, true, true, "Pattern matching construct that executes a block of code corresponding to the first matching case value");
  ps.ctx.set(switchIdx, switchBuiltinObj);
  
  // Register the when builtin
  int whenIdx = ps.idx.indexWord("when");
  Builtin whenBuiltinObj = Builtin(whenBuiltin, 3, false, true, "Conditionally executes an action block if a condition block evaluates to true");
  ps.ctx.set(whenIdx, whenBuiltinObj);
  
  // Register the while builtin
  int whileIdx = ps.idx.indexWord("while");
  Builtin whileBuiltinObj = Builtin(whileBuiltin, 2, false, false, "Executes a block of code repeatedly while a condition is true");
  ps.ctx.set(whileIdx, whileBuiltinObj);
}
