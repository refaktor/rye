// builtins_strings.dart - String builtins

import 'env.dart' show ProgramState; // Import ProgramState
// Import specific types needed from types.dart
import 'types.dart' show RyeObject, RyeString, Error, Integer, Builtin, RyeType; 

// --- String Builtins ---

// Implements the "+" builtin for string concatenation
RyeObject stringConcatBuiltin(ProgramState ps, RyeObject? arg0, RyeObject? arg1, RyeObject? arg2, RyeObject? arg3, RyeObject? arg4) {
  // We expect at least two arguments, which can be strings or other printable types
  if (arg0 != null && arg1 != null) {
    // Convert arguments to their string representation for concatenation
    // Note: RyeString.print includes quotes, which we might not want here.
    // Using inspect or a dedicated .toString() might be better, but let's use print for now.
    // A more robust implementation would handle type checking more carefully.
    
    String str0 = "";
    if (arg0 is RyeString) {
      str0 = arg0.value; // Use raw value for concatenation
    } else {
      str0 = arg0.print(ps.idx); // Fallback to print representation
    }

    String str1 = "";
     if (arg1 is RyeString) {
      str1 = arg1.value; // Use raw value for concatenation
    } else {
      str1 = arg1.print(ps.idx); // Fallback to print representation
    }

    return RyeString(str0 + str1);
  }
  ps.failureFlag = true;
  return Error("'+' (string) expects two arguments to concatenate");
}

// Implements the "length" builtin for strings
RyeObject stringLengthBuiltin(ProgramState ps, RyeObject? arg0, RyeObject? arg1, RyeObject? arg2, RyeObject? arg3, RyeObject? arg4) {
  if (arg0 is RyeString) {
    return Integer(arg0.value.length.toInt());
  }
  ps.failureFlag = true;
  return Error("length (string) expects a String");
}

// TODO: Implement other string builtins (split, join, find, replace, etc.)

// --- Registration ---

void registerStringBuiltins(ProgramState ps) {
  // Register string builtins generically
  int concatIdx = ps.idx.indexWord("+"); // Using '+' for now - needs opword handling
  int lengthIdx = ps.idx.indexWord("length");

  // Register '+' for string concatenation generically
  // Note: Proper opword handling in the evaluator is needed for '+' to work correctly alongside numeric '+'
  ps.registerGeneric(RyeType.stringType.index, concatIdx, 
    Builtin(stringConcatBuiltin, 2, false, true, "Concatenates two strings"));

  // Register 'length' for strings generically
  ps.registerGeneric(RyeType.stringType.index, lengthIdx, 
    Builtin(stringLengthBuiltin, 1, false, true, "Returns the length of a string"));
}
