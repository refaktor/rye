// builtins_iteration.dart - Iteration builtins for the Dart implementation of Rye

import 'env.dart' show ProgramState; // Import ProgramState
// Import specific types needed from types.dart
import 'types.dart' show RyeObject, Error, Block, RyeList, RyeString, RyeDict, Integer, Builtin, TSeries; // Import TSeries from types.dart
import 'evaldo.dart' show evalBlockInj; // Import the correct evaluator function
import 'builtins_conditionals.dart' show isTruthy; // For isTruthy function

// Implements the "for" builtin function
RyeObject forBuiltin(ProgramState ps, RyeObject? arg0, RyeObject? arg1, RyeObject? arg2, RyeObject? arg3, RyeObject? arg4) {
  if (arg0 == null || arg1 == null) {
    ps.failureFlag = true;
    return Error("for requires two arguments");
  }
  
  // Check if the second argument is a block
  if (arg1 is Block) {
    // Store current series
    TSeries ser = ps.ser;
    
    // Handle different collection types
    if (arg0 is Block) {
      // Iterate over each item in the block
      for (int i = 0; i < arg0.series.len(); i++) {
        RyeObject? item = arg0.series.get(i);
        if (item == null) continue;
        
        // Set series to the action block's series
        ps.ser = arg1.series;
        
        // Reset the series position to ensure we start from the beginning
        ps.ser.reset();
        
        // Evaluate the action block with the current item injected
        evalBlockInj(ps, item, true); // Use evalBlockInj
        
        // Check for errors or return flag
        if (ps.errorFlag || ps.returnFlag) {
          ps.ser = ser;
          return ps.res!;
        }
      }
    } else if (arg0 is RyeList) {
      // Iterate over each item in the list
      for (int i = 0; i < arg0.value.length; i++) { // Use .value instead of .items
        RyeObject? item = arg0.value[i]; // Use .value instead of .items
        if (item == null) continue; // Handle potential null items
        
        // Set series to the action block's series
        ps.ser = arg1.series;
        
        // Reset the series position to ensure we start from the beginning
        ps.ser.reset();
        
        // Evaluate the action block with the current item injected
        evalBlockInj(ps, item, true); // Use evalBlockInj
        
        // Check for errors or return flag
        if (ps.errorFlag || ps.returnFlag) {
          ps.ser = ser;
          return ps.res!;
        }
      }
    } else if (arg0 is RyeString) {
      // Iterate over each character in the string
      for (int i = 0; i < arg0.value.length; i++) {
        RyeObject item = RyeString(arg0.value[i]);
        
        // Set series to the action block's series
        ps.ser = arg1.series;
        
        // Reset the series position to ensure we start from the beginning
        ps.ser.reset();
        
        // Evaluate the action block with the current character injected
        evalBlockInj(ps, item, true); // Use evalBlockInj
        
        // Check for errors or return flag
        if (ps.errorFlag || ps.returnFlag) {
          ps.ser = ser;
          return ps.res!;
        }
      }
    } else if (arg0 is RyeDict) {
      // Iterate over each key in the dictionary
      for (String key in arg0.value.keys) { // Use .value instead of .entries
        RyeObject item = RyeString(key);
        
        // Set series to the action block's series
        ps.ser = arg1.series;
        
        // Reset the series position to ensure we start from the beginning
        ps.ser.reset();
        
        // Evaluate the action block with the current key injected
        evalBlockInj(ps, item, true); // Use evalBlockInj
        
        // Check for errors or return flag
        if (ps.errorFlag || ps.returnFlag) {
          ps.ser = ser;
          return ps.res!;
        }
      }
    } else {
      // Unsupported collection type
      ps.failureFlag = true;
      return Error("for expects a collection (block, list, string, or dict) as its first argument");
    }
    
    // Restore original series
    ps.ser = ser;
    
    // Return the result of the last evaluation
    return ps.res!;
  }
  
  // If second argument is not a block, return an error
  ps.failureFlag = true;
  return Error("for expects a block as its second argument");
}

// Implements the "forever" builtin function
RyeObject foreverBuiltin(ProgramState ps, RyeObject? arg0, RyeObject? arg1, RyeObject? arg2, RyeObject? arg3, RyeObject? arg4) {
  if (arg0 == null) {
    ps.failureFlag = true;
    return Error("forever requires one argument");
  }
  
  // Check if the argument is a block
  if (arg0 is Block) {
    // Store current series
    TSeries ser = ps.ser;
    
    // Loop indefinitely until return flag is set
    for (int i = 0; ; i++) {
      // Set series to the block's series
      ps.ser = arg0.series;
      
      // Reset the series position to ensure we start from the beginning
      ps.ser.reset();
      
      // Evaluate the block with the current iteration number injected
      evalBlockInj(ps, Integer(i), true); // Use evalBlockInj
      
      // Check for errors
      if (ps.errorFlag) {
        ps.ser = ser;
        return ps.res!;
      }
      
      // Check for return flag
      if (ps.returnFlag) {
        ps.returnFlag = false;
        break;
      }
    }
    
    // Restore original series
    ps.ser = ser;
    
    // Return the result of the last evaluation
    return ps.res!;
  }
  
  // If argument is not a block, return an error
  ps.failureFlag = true;
  return Error("forever expects a block as its argument");
}

// Implements the "forever-with" builtin function
RyeObject foreverWithBuiltin(ProgramState ps, RyeObject? arg0, RyeObject? arg1, RyeObject? arg2, RyeObject? arg3, RyeObject? arg4) {
  if (arg0 == null || arg1 == null) {
    ps.failureFlag = true;
    return Error("forever-with requires two arguments");
  }
  
  // Check if the second argument is a block
  if (arg1 is Block) {
    // Store current series
    TSeries ser = ps.ser;
    
    // Loop indefinitely until return flag is set
    while (true) {
      // Set series to the block's series
      ps.ser = arg1.series;
      
      // Reset the series position to ensure we start from the beginning
      ps.ser.reset();
      
      // Evaluate the block with the value injected
      evalBlockInj(ps, arg0, true); // Use evalBlockInj
      
      // Check for errors
      if (ps.errorFlag) {
        ps.ser = ser;
        return ps.res!;
      }
      
      // Check for return flag
      if (ps.returnFlag) {
        ps.returnFlag = false;
        break;
      }
    }
    
    // Restore original series
    ps.ser = ser;
    
    // Return the result of the last evaluation
    return ps.res!;
  }
  
  // If second argument is not a block, return an error
  ps.failureFlag = true;
  return Error("forever-with expects a block as its second argument");
}

// Implements the "map" builtin function
RyeObject mapBuiltin(ProgramState ps, RyeObject? arg0, RyeObject? arg1, RyeObject? arg2, RyeObject? arg3, RyeObject? arg4) {
  if (arg0 == null || arg1 == null) {
    ps.failureFlag = true;
    return Error("map requires two arguments");
  }
  
  // Check if the second argument is a block
  if (arg1 is Block) {
    // Store current series
    TSeries ser = ps.ser;
    
    // Handle different collection types
    if (arg0 is Block) {
      // Create a new list to store the mapped values
      List<RyeObject> mappedItems = [];
      
      // Map each item in the block
      for (int i = 0; i < arg0.series.len(); i++) {
        RyeObject? item = arg0.series.get(i);
        if (item == null) continue;
        
        // Set series to the mapping block's series
        ps.ser = arg1.series;
        
        // Reset the series position to ensure we start from the beginning
        ps.ser.reset();
        
        // Evaluate the mapping block with the current item injected
        evalBlockInj(ps, item, true); // Use evalBlockInj
        
        // Check for errors
        if (ps.errorFlag) {
          ps.ser = ser;
          return ps.res!;
        }
        
        // Add the result to the mapped items
        if (ps.res != null) {
          mappedItems.add(ps.res!);
        }
      }
      
      // Create a new block with the mapped items
      return Block(TSeries(mappedItems));
    } else if (arg0 is RyeList) {
      // Create a new list to store the mapped values
      List<RyeObject> mappedItems = [];
      
      // Map each item in the list
      for (int i = 0; i < arg0.value.length; i++) { // Use .value instead of .items
        RyeObject? item = arg0.value[i]; // Use .value instead of .items
        if (item == null) continue; // Handle potential null items
        
        // Set series to the mapping block's series
        ps.ser = arg1.series;
        
        // Reset the series position to ensure we start from the beginning
        ps.ser.reset();
        
        // Evaluate the mapping block with the current item injected
        evalBlockInj(ps, item, true); // Use evalBlockInj
        
        // Check for errors
        if (ps.errorFlag) {
          ps.ser = ser;
          return ps.res!;
        }
        
        // Add the result to the mapped items
        if (ps.res != null) {
          mappedItems.add(ps.res!);
        }
      }
      
      // Create a new list with the mapped items
      return RyeList(mappedItems);
    } else if (arg0 is RyeString) {
      // Create a new list to store the mapped values
      List<RyeObject> mappedItems = [];
      
      // Map each character in the string
      for (int i = 0; i < arg0.value.length; i++) {
        RyeObject item = RyeString(arg0.value[i]);
        
        // Set series to the mapping block's series
        ps.ser = arg1.series;
        
        // Reset the series position to ensure we start from the beginning
        ps.ser.reset();
        
        // Evaluate the mapping block with the current character injected
        evalBlockInj(ps, item, true); // Use evalBlockInj
        
        // Check for errors
        if (ps.errorFlag) {
          ps.ser = ser;
          return ps.res!;
        }
        
        // Add the result to the mapped items
        if (ps.res != null) {
          mappedItems.add(ps.res!);
        }
      }
      
      // Create a new block with the mapped items
      return Block(TSeries(mappedItems));
    } else {
      // Unsupported collection type
      ps.failureFlag = true;
      return Error("map expects a collection (block, list, or string) as its first argument");
    }
  }
  
  // If second argument is not a block, return an error
  ps.failureFlag = true;
  return Error("map expects a block as its second argument");
}

// Implements the "filter" builtin function
RyeObject filterBuiltin(ProgramState ps, RyeObject? arg0, RyeObject? arg1, RyeObject? arg2, RyeObject? arg3, RyeObject? arg4) {
  if (arg0 == null || arg1 == null) {
    ps.failureFlag = true;
    return Error("filter requires two arguments");
  }
  
  // Check if the second argument is a block
  if (arg1 is Block) {
    // Store current series
    TSeries ser = ps.ser;
    
    // Handle different collection types
    if (arg0 is Block) {
      // Create a new list to store the filtered values
      List<RyeObject> filteredItems = [];
      
      // Filter each item in the block
      for (int i = 0; i < arg0.series.len(); i++) {
        RyeObject? item = arg0.series.get(i);
        if (item == null) continue;
        
        // Set series to the filter block's series
        ps.ser = arg1.series;
        
        // Reset the series position to ensure we start from the beginning
        ps.ser.reset();
        
        // Evaluate the filter block with the current item injected
        evalBlockInj(ps, item, true); // Use evalBlockInj
        
        // Check for errors
        if (ps.errorFlag) {
          ps.ser = ser;
          return ps.res!;
        }
        
        // Add the item to the filtered items if the result is truthy
        if (ps.res != null && isTruthy(ps.res!)) {
          filteredItems.add(item);
        }
      }
      
      // Create a new block with the filtered items
      return Block(TSeries(filteredItems));
    } else if (arg0 is RyeList) {
      // Create a new list to store the filtered values
      List<RyeObject> filteredItems = [];
      
      // Filter each item in the list
      for (int i = 0; i < arg0.value.length; i++) { // Use .value instead of .items
        RyeObject? item = arg0.value[i]; // Use .value instead of .items
         if (item == null) continue; // Handle potential null items

        // Set series to the filter block's series
        ps.ser = arg1.series;
        
        // Reset the series position to ensure we start from the beginning
        ps.ser.reset();
        
        // Evaluate the filter block with the current item injected
        evalBlockInj(ps, item, true); // Use evalBlockInj
        
        // Check for errors
        if (ps.errorFlag) {
          ps.ser = ser;
          return ps.res!;
        }
        
        // Add the item to the filtered items if the result is truthy
        if (ps.res != null && isTruthy(ps.res!)) {
          filteredItems.add(item);
        }
      }
      
      // Create a new list with the filtered items
      return RyeList(filteredItems);
    } else if (arg0 is RyeString) {
      // Create a new list to store the filtered values
      List<RyeObject> filteredItems = [];
      
      // Filter each character in the string
      for (int i = 0; i < arg0.value.length; i++) {
        RyeObject item = RyeString(arg0.value[i]);
        
        // Set series to the filter block's series
        ps.ser = arg1.series;
        
        // Reset the series position to ensure we start from the beginning
        ps.ser.reset();
        
        // Evaluate the filter block with the current character injected
        evalBlockInj(ps, item, true); // Use evalBlockInj
        
        // Check for errors
        if (ps.errorFlag) {
          ps.ser = ser;
          return ps.res!;
        }
        
        // Add the character to the filtered items if the result is truthy
        if (ps.res != null && isTruthy(ps.res!)) {
          filteredItems.add(item);
        }
      }
      
      // Create a new block with the filtered items
      return Block(TSeries(filteredItems));
    } else {
      // Unsupported collection type
      ps.failureFlag = true;
      return Error("filter expects a collection (block, list, or string) as its first argument");
    }
  }
  
  // If second argument is not a block, return an error
  ps.failureFlag = true;
  return Error("filter expects a block as its second argument");
}

// Register the iteration builtins
void registerIterationBuiltins(ProgramState ps) {
  // Register the for builtin
  int forIdx = ps.idx.indexWord("for");
  Builtin forBuiltinObj = Builtin(forBuiltin, 2, false, true, "Iterates over each value in a collection executing a block of code for each value");
  ps.ctx.set(forIdx, forBuiltinObj);
  
  // Register the forever builtin
  int foreverIdx = ps.idx.indexWord("forever");
  Builtin foreverBuiltinObj = Builtin(foreverBuiltin, 1, false, false, "Executes a block of code repeatedly until .return is called within the block");
  ps.ctx.set(foreverIdx, foreverBuiltinObj);
  
  // Register the forever-with builtin
  int foreverWithIdx = ps.idx.indexWord("forever-with");
  Builtin foreverWithBuiltinObj = Builtin(foreverWithBuiltin, 2, false, false, "Accepts a value and a block and executes the block repeatedly with the value until .return is called");
  ps.ctx.set(foreverWithIdx, foreverWithBuiltinObj);
  
  // Register the map builtin
  int mapIdx = ps.idx.indexWord("map");
  Builtin mapBuiltinObj = Builtin(mapBuiltin, 2, false, true, "Maps values of a collection to a new collection by evaluating a block of code");
  ps.ctx.set(mapIdx, mapBuiltinObj);
  
  // Register the filter builtin
  int filterIdx = ps.idx.indexWord("filter");
  Builtin filterBuiltinObj = Builtin(filterBuiltin, 2, false, true, "Filters values from a collection based on return of a injected code block");
  ps.ctx.set(filterIdx, filterBuiltinObj);
}
