// builtins.dart - Base builtin functions and registration for Dart Rye

import 'dart:io'; // For printBuiltin
import 'dart:async'; // For flutterWindowBuiltin

// Updated Imports:
import 'types.dart' show RyeObject, Integer, Error, Void, Block, TSeries, Builtin; // Import specific types
import 'env.dart' show ProgramState, RyeCtx; // Import specific env types
import 'idxs.dart'; // Import idxs for Idxs
import 'evaldo.dart' show evalBlockInj; // Import specific evaldo functions needed
import 'flutter/flutter_app.dart'; // For flutterWindowBuiltin

// Import other builtin modules
import 'builtins_collections.dart';
import 'builtins_conditionals.dart';
import 'builtins_flow.dart';
import 'builtins_iteration.dart';
import 'builtins_numbers.dart';
import 'builtins_printing.dart';
import 'builtins_registration.dart'; // Assuming this contains registerCoreBuiltins etc.
import 'builtins_strings.dart';
import 'builtins_types.dart';


// --- Builtin Implementations ---

// addBuiltin moved to builtins_numbers.dart
// printBuiltin moved to builtins_printing.dart

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
      RyeObject result = const Void(); // Default result if loop runs 0 times or block is empty
      
      // Execute the block 'iterations' times
      for (int i = 0; i < iterations; i++) {
        // Store current series and context
        TSeries ser = ps.ser;
        RyeCtx ctx = ps.ctx; // Store context in case block evaluation changes it (though it shouldn't normally)
        
        // Set series to the block's series
        ps.ser = arg1.series;
        
        // Reset the series position to ensure we start from the beginning
        ps.ser.reset();
        
        // Evaluate the block, injecting the 1-based iteration number
        // Use a temporary state for block evaluation if needed, or ensure evalBlockInj handles context correctly
        evalBlockInj(ps, Integer(i + 1), true); // Use the main evalBlockInj now
    
        // Store the result of the *last* expression evaluated in the block for this iteration
        result = ps.res ?? const Void(); 
        
        // Check for errors or failures or return flag from the block
        if (ps.errorFlag || ps.failureFlag || ps.returnFlag) {
          // Restore original series and context before returning the error/failure/return value
          ps.ser = ser;
          ps.ctx = ctx; 
          return result; // Propagate error/failure/return value
        }
        
        // Restore original series and context for the next iteration or after loop finishes
        ps.ser = ser;
        ps.ctx = ctx;
      }
      // Return the result of the last iteration (or Void if 0 iterations)
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


// --- Builtin Registration ---

// Register all builtins
void registerBuiltins(ProgramState ps) {
  // Register core builtins from builtins_registration.dart (assuming it exists)
  // registerCoreBuiltins(ps); // Uncomment if registerCoreBuiltins is defined elsewhere
  
  // Register the loop builtin
  int loopIdx = ps.idx.indexWord("loop");
  Builtin loopBuiltinObj = Builtin(loopBuiltin, 2, false, false, "Executes a block a specified number of times");
  ps.ctx.set(loopIdx, loopBuiltinObj);
  
  // Register the flutter_window builtin
  int flutterWindowIdx = ps.idx.indexWord("flutter_window");
  Builtin flutterWindowBuiltinObj = Builtin(flutterWindowBuiltin, 2, false, false, "Shows a Flutter window with a title and message");
  ps.ctx.set(flutterWindowIdx, flutterWindowBuiltinObj);
  
  // Register builtins from other modules (assuming these functions exist)
  registerCollectionBuiltins(ps);
  registerStringBuiltins(ps);
  registerFlowBuiltins(ps);
  registerNumberBuiltins(ps);
  registerIterationBuiltins(ps);
  registerConditionalBuiltins(ps);
  registerPrintingBuiltins(ps);
  registerTypeBuiltins(ps);
}
