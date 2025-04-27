// builtins_numbers.dart - Number-related builtins for the Dart implementation of Rye

import 'env.dart' show ProgramState; // Import ProgramState
// Import specific types needed from types.dart
import 'types.dart' show RyeObject, Integer, Error, Boolean, Decimal, Time, RyeString, Builtin, RyeType; 

// Implements the "inc" builtin function
RyeObject incBuiltin(ProgramState ps, RyeObject? arg0, RyeObject? arg1, RyeObject? arg2, RyeObject? arg3, RyeObject? arg4) {
  if (arg0 == null) {
    ps.failureFlag = true;
    return Error("inc requires an argument");
  }
  
  if (arg0 is Integer) {
    return Integer(arg0.value + 1);
  }
  
  ps.failureFlag = true;
  return Error("inc expects an integer argument");
}

// Implements the "is-positive" builtin function
RyeObject isPositiveBuiltin(ProgramState ps, RyeObject? arg0, RyeObject? arg1, RyeObject? arg2, RyeObject? arg3, RyeObject? arg4) {
  if (arg0 == null) {
    ps.failureFlag = true;
    return Error("is-positive requires an argument");
  }
  
  if (arg0 is Integer) {
    return Boolean(arg0.value > 0);
  }
  
  if (arg0 is Decimal) {
    return Boolean(arg0.value > 0);
  }
  
  ps.failureFlag = true;
  return Error("is-positive expects a number argument");
}

// Implements the "is-zero" builtin function
RyeObject isZeroBuiltin(ProgramState ps, RyeObject? arg0, RyeObject? arg1, RyeObject? arg2, RyeObject? arg3, RyeObject? arg4) {
  if (arg0 == null) {
    ps.failureFlag = true;
    return Error("is-zero requires an argument");
  }
  
  if (arg0 is Integer) {
    return Boolean(arg0.value == 0);
  }
  
  if (arg0 is Decimal) {
    return Boolean(arg0.value == 0);
  }
  
  ps.failureFlag = true;
  return Error("is-zero expects a number argument");
}

// Implements the "is-odd" builtin function
RyeObject isOddBuiltin(ProgramState ps, RyeObject? arg0, RyeObject? arg1, RyeObject? arg2, RyeObject? arg3, RyeObject? arg4) {
  if (arg0 == null) {
    ps.failureFlag = true;
    return Error("is-odd requires an argument");
  }
  
  if (arg0 is Integer) {
    return Boolean(arg0.value % 2 != 0);
  }
  
  ps.failureFlag = true;
  return Error("is-odd expects an integer argument");
}

// Implements the "is-even" builtin function
RyeObject isEvenBuiltin(ProgramState ps, RyeObject? arg0, RyeObject? arg1, RyeObject? arg2, RyeObject? arg3, RyeObject? arg4) {
  if (arg0 == null) {
    ps.failureFlag = true;
    return Error("is-even requires an argument");
  }
  
  if (arg0 is Integer) {
    return Boolean(arg0.value % 2 == 0);
  }
  
  ps.failureFlag = true;
  return Error("is-even expects an integer argument");
}

// Implements the "mod" builtin function
RyeObject modBuiltin(ProgramState ps, RyeObject? arg0, RyeObject? arg1, RyeObject? arg2, RyeObject? arg3, RyeObject? arg4) {
  if (arg0 == null || arg1 == null) {
    ps.failureFlag = true;
    return Error("mod requires two arguments");
  }
  
  if (arg0 is Integer && arg1 is Integer) {
    if (arg1.value == 0) {
      ps.failureFlag = true;
      return Error("mod: division by zero");
    }
    return Integer(arg0.value % arg1.value);
  }
  
  ps.failureFlag = true;
  return Error("mod expects two integer arguments");
}

// Implements the "_-" (subtraction) builtin function
RyeObject subtractBuiltin(ProgramState ps, RyeObject? arg0, RyeObject? arg1, RyeObject? arg2, RyeObject? arg3, RyeObject? arg4) {
  if (arg0 == null || arg1 == null) {
    ps.failureFlag = true;
    return Error("_- requires two arguments");
  }
  
  // Integer - Integer
  if (arg0 is Integer && arg1 is Integer) {
    return Integer(arg0.value - arg1.value);
  }
  
  // Integer - Decimal
  if (arg0 is Integer && arg1 is Decimal) {
    return Decimal(arg0.value.toDouble() - arg1.value);
  }
  
  // Decimal - Integer
  if (arg0 is Decimal && arg1 is Integer) {
    return Decimal(arg0.value - arg1.value.toDouble());
  }
  
  // Decimal - Decimal
  if (arg0 is Decimal && arg1 is Decimal) {
    return Decimal(arg0.value - arg1.value);
  }
  
  // Time - Integer (subtract milliseconds)
  if (arg0 is Time && arg1 is Integer) {
    return Time(arg0.value.subtract(Duration(milliseconds: arg1.value)));
  }
  
  // Time - Time (returns duration in milliseconds)
  if (arg0 is Time && arg1 is Time) {
    return Integer(arg0.value.difference(arg1.value).inMilliseconds);
  }
  
  ps.failureFlag = true;
  return Error("_- expects numeric arguments");
}

// Implements the "_*" (multiplication) builtin function
RyeObject multiplyBuiltin(ProgramState ps, RyeObject? arg0, RyeObject? arg1, RyeObject? arg2, RyeObject? arg3, RyeObject? arg4) {
  if (arg0 == null || arg1 == null) {
    ps.failureFlag = true;
    return Error("_* requires two arguments");
  }
  
  // Integer * Integer
  if (arg0 is Integer && arg1 is Integer) {
    return Integer(arg0.value * arg1.value);
  }
  
  // Integer * Decimal
  if (arg0 is Integer && arg1 is Decimal) {
    return Decimal(arg0.value.toDouble() * arg1.value);
  }
  
  // Decimal * Integer
  if (arg0 is Decimal && arg1 is Integer) {
    return Decimal(arg0.value * arg1.value.toDouble());
  }
  
  // Decimal * Decimal
  if (arg0 is Decimal && arg1 is Decimal) {
    return Decimal(arg0.value * arg1.value);
  }
  
  ps.failureFlag = true;
  return Error("_* expects numeric arguments");
}

// Implements the "_/" (division) builtin function
RyeObject divideBuiltin(ProgramState ps, RyeObject? arg0, RyeObject? arg1, RyeObject? arg2, RyeObject? arg3, RyeObject? arg4) {
  if (arg0 == null || arg1 == null) {
    ps.failureFlag = true;
    return Error("_/ requires two arguments");
  }
  
  // Check for division by zero
  if ((arg1 is Integer && arg1.value == 0) || (arg1 is Decimal && arg1.value == 0)) {
    ps.failureFlag = true;
    return Error("_/: division by zero");
  }
  
  // Integer / Integer
  if (arg0 is Integer && arg1 is Integer) {
    return Decimal(arg0.value.toDouble() / arg1.value.toDouble());
  }
  
  // Integer / Decimal
  if (arg0 is Integer && arg1 is Decimal) {
    return Decimal(arg0.value.toDouble() / arg1.value);
  }
  
  // Decimal / Integer
  if (arg0 is Decimal && arg1 is Integer) {
    return Decimal(arg0.value / arg1.value.toDouble());
  }
  
  // Decimal / Decimal
  if (arg0 is Decimal && arg1 is Decimal) {
    return Decimal(arg0.value / arg1.value);
  }
  
  ps.failureFlag = true;
  return Error("_/ expects numeric arguments");
}

// Implements the "_//" (integer division) builtin function
RyeObject intDivideBuiltin(ProgramState ps, RyeObject? arg0, RyeObject? arg1, RyeObject? arg2, RyeObject? arg3, RyeObject? arg4) {
  if (arg0 == null || arg1 == null) {
    ps.failureFlag = true;
    return Error("_// requires two arguments");
  }
  
  // Check for division by zero
  if ((arg1 is Integer && arg1.value == 0) || (arg1 is Decimal && arg1.value == 0)) {
    ps.failureFlag = true;
    return Error("_//: division by zero");
  }
  
  // Integer // Integer
  if (arg0 is Integer && arg1 is Integer) {
    return Integer(arg0.value ~/ arg1.value);
  }
  
  // Integer // Decimal
  if (arg0 is Integer && arg1 is Decimal) {
    return Integer(arg0.value ~/ arg1.value.toInt());
  }
  
  // Decimal // Integer
  if (arg0 is Decimal && arg1 is Integer) {
    return Integer(arg0.value.toInt() ~/ arg1.value);
  }
  
  // Decimal // Decimal
  if (arg0 is Decimal && arg1 is Decimal) {
    return Integer(arg0.value.toInt() ~/ arg1.value.toInt());
  }
  
  ps.failureFlag = true;
  return Error("_// expects numeric arguments");
}

// Implements the "_>" (greater than) builtin function
RyeObject greaterThanBuiltin(ProgramState ps, RyeObject? arg0, RyeObject? arg1, RyeObject? arg2, RyeObject? arg3, RyeObject? arg4) {
  if (arg0 == null || arg1 == null) {
    ps.failureFlag = true;
    return Error("_> requires two arguments");
  }
  
  // Integer > Integer
  if (arg0 is Integer && arg1 is Integer) {
    return Boolean(arg0.value > arg1.value);
  }
  
  // Integer > Decimal
  if (arg0 is Integer && arg1 is Decimal) {
    return Boolean(arg0.value.toDouble() > arg1.value);
  }
  
  // Decimal > Integer
  if (arg0 is Decimal && arg1 is Integer) {
    return Boolean(arg0.value > arg1.value.toDouble());
  }
  
  // Decimal > Decimal
  if (arg0 is Decimal && arg1 is Decimal) {
    return Boolean(arg0.value > arg1.value);
  }
  
  // String > String (lexicographical comparison)
  if (arg0 is RyeString && arg1 is RyeString) {
    return Boolean(arg0.value.compareTo(arg1.value) > 0);
  }
  
  ps.failureFlag = true;
  return Error("_> expects comparable arguments");
}

// Implements the "_>=" (greater than or equal) builtin function
RyeObject greaterThanOrEqualBuiltin(ProgramState ps, RyeObject? arg0, RyeObject? arg1, RyeObject? arg2, RyeObject? arg3, RyeObject? arg4) {
  if (arg0 == null || arg1 == null) {
    ps.failureFlag = true;
    return Error("_>= requires two arguments");
  }
  
  // Integer >= Integer
  if (arg0 is Integer && arg1 is Integer) {
    return Boolean(arg0.value >= arg1.value);
  }
  
  // Integer >= Decimal
  if (arg0 is Integer && arg1 is Decimal) {
    return Boolean(arg0.value.toDouble() >= arg1.value);
  }
  
  // Decimal >= Integer
  if (arg0 is Decimal && arg1 is Integer) {
    return Boolean(arg0.value >= arg1.value.toDouble());
  }
  
  // Decimal >= Decimal
  if (arg0 is Decimal && arg1 is Decimal) {
    return Boolean(arg0.value >= arg1.value);
  }
  
  // String >= String (lexicographical comparison)
  if (arg0 is RyeString && arg1 is RyeString) {
    return Boolean(arg0.value.compareTo(arg1.value) >= 0);
  }
  
  // Check for equality first
  if (arg0.equal(arg1)) {
    return Boolean(true);
  }
  
  ps.failureFlag = true;
  return Error("_>= expects comparable arguments");
}

// Implements the "_<" (less than) builtin function
RyeObject lessThanBuiltin(ProgramState ps, RyeObject? arg0, RyeObject? arg1, RyeObject? arg2, RyeObject? arg3, RyeObject? arg4) {
  if (arg0 == null || arg1 == null) {
    ps.failureFlag = true;
    return Error("_< requires two arguments");
  }
  
  // Integer < Integer
  if (arg0 is Integer && arg1 is Integer) {
    return Boolean(arg0.value < arg1.value);
  }
  
  // Integer < Decimal
  if (arg0 is Integer && arg1 is Decimal) {
    return Boolean(arg0.value.toDouble() < arg1.value);
  }
  
  // Decimal < Integer
  if (arg0 is Decimal && arg1 is Integer) {
    return Boolean(arg0.value < arg1.value.toDouble());
  }
  
  // Decimal < Decimal
  if (arg0 is Decimal && arg1 is Decimal) {
    return Boolean(arg0.value < arg1.value);
  }
  
  // String < String (lexicographical comparison)
  if (arg0 is RyeString && arg1 is RyeString) {
    return Boolean(arg0.value.compareTo(arg1.value) < 0);
  }
  
  ps.failureFlag = true;
  return Error("_< expects comparable arguments");
}

// Implements the "_<=" (less than or equal) builtin function
RyeObject lessThanOrEqualBuiltin(ProgramState ps, RyeObject? arg0, RyeObject? arg1, RyeObject? arg2, RyeObject? arg3, RyeObject? arg4) {
  if (arg0 == null || arg1 == null) {
    ps.failureFlag = true;
    return Error("_<= requires two arguments");
  }
  
  // Integer <= Integer
  if (arg0 is Integer && arg1 is Integer) {
    return Boolean(arg0.value <= arg1.value);
  }
  
  // Integer <= Decimal
  if (arg0 is Integer && arg1 is Decimal) {
    return Boolean(arg0.value.toDouble() <= arg1.value);
  }
  
  // Decimal <= Integer
  if (arg0 is Decimal && arg1 is Integer) {
    return Boolean(arg0.value <= arg1.value.toDouble());
  }
  
  // Decimal <= Decimal
  if (arg0 is Decimal && arg1 is Decimal) {
    return Boolean(arg0.value <= arg1.value);
  }
  
  // String <= String (lexicographical comparison)
  if (arg0 is RyeString && arg1 is RyeString) {
    return Boolean(arg0.value.compareTo(arg1.value) <= 0);
  }
  
  // Check for equality first
  if (arg0.equal(arg1)) {
    return Boolean(true);
  }
  
  ps.failureFlag = true;
  return Error("_<= expects comparable arguments");
}

// Register the number-related builtins
void registerNumberBuiltins(ProgramState ps) {
  // Register non-operator builtins directly in context
  ps.ctx.set(ps.idx.indexWord("inc"), 
    Builtin(incBuiltin, 1, false, true, "Increments an integer value by 1"));
  ps.ctx.set(ps.idx.indexWord("is-positive"), 
    Builtin(isPositiveBuiltin, 1, false, true, "Checks if a number is positive (greater than zero)"));
  ps.ctx.set(ps.idx.indexWord("is-zero"), 
    Builtin(isZeroBuiltin, 1, false, true, "Checks if a number is exactly zero"));
  ps.ctx.set(ps.idx.indexWord("is-odd"), 
    Builtin(isOddBuiltin, 1, false, true, "Checks if an integer is odd (not divisible by 2)"));
  ps.ctx.set(ps.idx.indexWord("is-even"), 
    Builtin(isEvenBuiltin, 1, false, true, "Checks if an integer is even (divisible by 2)"));
  ps.ctx.set(ps.idx.indexWord("mod"), 
    Builtin(modBuiltin, 2, false, true, "Calculates the modulo (remainder) when dividing the first integer by the second"));

  // Register operators generically based on the type of the first argument (left-hand side)
  int minusIdx = ps.idx.indexWord("_-");
  Builtin minusBuiltinObj = Builtin(subtractBuiltin, 2, false, true, "Subtracts the second number from the first");
  ps.registerGeneric(RyeType.integerType.index, minusIdx, minusBuiltinObj);
  ps.registerGeneric(RyeType.decimalType.index, minusIdx, minusBuiltinObj);
  ps.registerGeneric(RyeType.timeType.index, minusIdx, minusBuiltinObj); // Time - Int/Time

  int multiplyIdx = ps.idx.indexWord("_*");
  Builtin multiplyBuiltinObj = Builtin(multiplyBuiltin, 2, false, true, "Multiplies two numbers");
  ps.registerGeneric(RyeType.integerType.index, multiplyIdx, multiplyBuiltinObj);
  ps.registerGeneric(RyeType.decimalType.index, multiplyIdx, multiplyBuiltinObj);

  int divideIdx = ps.idx.indexWord("_/");
  Builtin divideBuiltinObj = Builtin(divideBuiltin, 2, false, true, "Divides the first number by the second");
  ps.registerGeneric(RyeType.integerType.index, divideIdx, divideBuiltinObj);
  ps.registerGeneric(RyeType.decimalType.index, divideIdx, divideBuiltinObj);

  int intDivideIdx = ps.idx.indexWord("_//");
  Builtin intDivideBuiltinObj = Builtin(intDivideBuiltin, 2, false, true, "Performs integer division");
  ps.registerGeneric(RyeType.integerType.index, intDivideIdx, intDivideBuiltinObj);
  ps.registerGeneric(RyeType.decimalType.index, intDivideIdx, intDivideBuiltinObj);

  int greaterThanIdx = ps.idx.indexWord("_>");
  Builtin greaterThanBuiltinObj = Builtin(greaterThanBuiltin, 2, false, true, "Compares if the first value is greater than the second value");
  ps.registerGeneric(RyeType.integerType.index, greaterThanIdx, greaterThanBuiltinObj);
  ps.registerGeneric(RyeType.decimalType.index, greaterThanIdx, greaterThanBuiltinObj);
  ps.registerGeneric(RyeType.stringType.index, greaterThanIdx, greaterThanBuiltinObj); // String comparison

  int greaterThanOrEqualIdx = ps.idx.indexWord("_>=");
  Builtin greaterThanOrEqualBuiltinObj = Builtin(greaterThanOrEqualBuiltin, 2, false, true, "Compares if the first value is greater than or equal to the second value");
  ps.registerGeneric(RyeType.integerType.index, greaterThanOrEqualIdx, greaterThanOrEqualBuiltinObj);
  ps.registerGeneric(RyeType.decimalType.index, greaterThanOrEqualIdx, greaterThanOrEqualBuiltinObj);
  ps.registerGeneric(RyeType.stringType.index, greaterThanOrEqualIdx, greaterThanOrEqualBuiltinObj); // String comparison

  int lessThanIdx = ps.idx.indexWord("_<");
  Builtin lessThanBuiltinObj = Builtin(lessThanBuiltin, 2, false, true, "Compares if the first value is less than the second value");
  ps.registerGeneric(RyeType.integerType.index, lessThanIdx, lessThanBuiltinObj);
  ps.registerGeneric(RyeType.decimalType.index, lessThanIdx, lessThanBuiltinObj);
  ps.registerGeneric(RyeType.stringType.index, lessThanIdx, lessThanBuiltinObj); // String comparison

  int lessThanOrEqualIdx = ps.idx.indexWord("_<=");
  Builtin lessThanOrEqualBuiltinObj = Builtin(lessThanOrEqualBuiltin, 2, false, true, "Compares if the first value is less than or equal to the second value");
  ps.registerGeneric(RyeType.integerType.index, lessThanOrEqualIdx, lessThanOrEqualBuiltinObj);
  ps.registerGeneric(RyeType.decimalType.index, lessThanOrEqualIdx, lessThanOrEqualBuiltinObj);
  ps.registerGeneric(RyeType.stringType.index, lessThanOrEqualIdx, lessThanOrEqualBuiltinObj); // String comparison
  
  // Note: '+' is handled in builtins_strings.dart for String and needs specific opword handling
  // in the evaluator to coexist with numeric addition. 
  // We register the numeric version here for Integer + Integer.
  // The evaluator's generic dispatch will pick the correct one based on the left operand's type.
  int plusIdx = ps.idx.indexWord("_+");
  // Assuming addBuiltin is defined in builtins.dart or here, and handles only integers for now
  // If addBuiltin is moved, import it. Let's assume it's accessible.
  // Builtin addBuiltinObj = Builtin(addBuiltin, 2, false, true, "Adds two integers"); 
  // ps.registerGeneric(RyeType.integerType.index, plusIdx, addBuiltinObj);
  // TODO: Add generic registration for Decimal + Decimal, etc. if addBuiltin is updated or new builtins are created.
}

// Assuming addBuiltin is defined here for now if not imported from builtins.dart
RyeObject addBuiltin(ProgramState ps, RyeObject? arg0, RyeObject? arg1, RyeObject? arg2, RyeObject? arg3, RyeObject? arg4) {
  if (arg0 is Integer && arg1 is Integer) {
    return Integer(arg0.value + arg1.value);
  }
  // TODO: Handle Decimal addition
  ps.failureFlag = true;
  return Error("Numeric addition currently only supports Integer + Integer");
}
