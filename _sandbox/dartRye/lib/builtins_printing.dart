// builtins_printing.dart - Printing builtins for the Dart implementation of Rye

import 'dart:io';
import 'rye.dart';
import 'types.dart';

// --- Printing Builtins ---

// Implements the "prns" builtin function
RyeObject prnsBuiltin(ProgramState ps, RyeObject? arg0, RyeObject? arg1, RyeObject? arg2, RyeObject? arg3, RyeObject? arg4) {
  if (arg0 != null) {
    // Print the value followed by a space
    if (arg0 is RyeString) {
      stdout.write("${arg0.value} ");
    } else {
      stdout.write("${arg0.print(ps.idx)} ");
    }
    return arg0;
  }
  ps.failureFlag = true;
  return Error("prns requires an argument");
}

// Implements the "prn" builtin function
RyeObject prnBuiltin(ProgramState ps, RyeObject? arg0, RyeObject? arg1, RyeObject? arg2, RyeObject? arg3, RyeObject? arg4) {
  if (arg0 != null) {
    // Print the value without a newline
    if (arg0 is RyeString) {
      stdout.write(arg0.value);
    } else {
      stdout.write(arg0.print(ps.idx));
    }
    return arg0;
  }
  ps.failureFlag = true;
  return Error("prn requires an argument");
}

// Implements the "print" builtin function
RyeObject printBuiltin(ProgramState ps, RyeObject? arg0, RyeObject? arg1, RyeObject? arg2, RyeObject? arg3, RyeObject? arg4) {
  if (arg0 != null) {
    // Print the value followed by a newline
    if (arg0 is RyeString) {
      stdout.writeln(arg0.value);
    } else {
      stdout.writeln(arg0.print(ps.idx));
    }
    return arg0;
  }
  ps.failureFlag = true;
  return Error("print requires an argument");
}

// Implements the "print2" builtin function
RyeObject print2Builtin(ProgramState ps, RyeObject? arg0, RyeObject? arg1, RyeObject? arg2, RyeObject? arg3, RyeObject? arg4) {
  if (arg0 != null && arg1 != null) {
    // Print the first value
    if (arg0 is RyeString) {
      stdout.write(arg0.value);
    } else {
      stdout.write(arg0.print(ps.idx));
    }
    
    // Print a space
    stdout.write(" ");
    
    // Print the second value followed by a newline
    if (arg1 is RyeString) {
      stdout.writeln(arg1.value);
    } else {
      stdout.writeln(arg1.print(ps.idx));
    }
    
    return arg1;
  }
  ps.failureFlag = true;
  return Error("print2 requires two arguments");
}

// Implements the "format" builtin function
RyeObject formatBuiltin(ProgramState ps, RyeObject? arg0, RyeObject? arg1, RyeObject? arg2, RyeObject? arg3, RyeObject? arg4) {
  if (arg0 != null && arg1 is RyeString) {
    String format = arg1.value;
    String result = "";
    
    // Simple format implementation - replace %d, %s, %f with the string representation of arg0
    if (arg0 is Integer) {
      result = format.replaceAll("%d", arg0.value.toString());
    } else if (arg0 is Decimal) {
      result = format.replaceAll("%f", arg0.value.toString());
    } else if (arg0 is RyeString) {
      result = format.replaceAll("%s", arg0.value);
    } else {
      result = format.replaceAll(RegExp(r'%[dfs]'), arg0.print(ps.idx));
    }
    
    return RyeString(result);
  }
  ps.failureFlag = true;
  return Error("format requires a value and a format string");
}

// Implements the "prnf" builtin function
RyeObject prnfBuiltin(ProgramState ps, RyeObject? arg0, RyeObject? arg1, RyeObject? arg2, RyeObject? arg3, RyeObject? arg4) {
  if (arg0 != null && arg1 is RyeString) {
    String format = arg1.value;
    String result = "";
    
    // Simple format implementation - replace %d, %s, %f with the string representation of arg0
    if (arg0 is Integer) {
      result = format.replaceAll("%d", arg0.value.toString());
    } else if (arg0 is Decimal) {
      result = format.replaceAll("%f", arg0.value.toString());
    } else if (arg0 is RyeString) {
      result = format.replaceAll("%s", arg0.value);
    } else {
      result = format.replaceAll(RegExp(r'%[dfs]'), arg0.print(ps.idx));
    }
    
    stdout.write(result);
    return arg0;
  }
  ps.failureFlag = true;
  return Error("prnf requires a value and a format string");
}

// Implements the "embed" builtin function
RyeObject embedBuiltin(ProgramState ps, RyeObject? arg0, RyeObject? arg1, RyeObject? arg2, RyeObject? arg3, RyeObject? arg4) {
  if (arg0 != null && arg1 != null) {
    String value = arg0.print(ps.idx);
    
    if (arg1 is RyeString) {
      String template = arg1.value;
      String result = template.replaceAll("{}", value);
      return RyeString(result);
    } else if (arg1 is Uri) {
      String path = arg1.path;
      String newPath = path.replaceAll("{}", value);
      return Uri(arg1.scheme, newPath, arg1.kind);
    }
  }
  ps.failureFlag = true;
  return Error("embed requires a value and a template string or URI");
}

// Implements the "prnv" builtin function
RyeObject prnvBuiltin(ProgramState ps, RyeObject? arg0, RyeObject? arg1, RyeObject? arg2, RyeObject? arg3, RyeObject? arg4) {
  if (arg0 != null && arg1 is RyeString) {
    String value = arg0.print(ps.idx);
    String template = arg1.value;
    String result = template.replaceAll("{}", value);
    
    stdout.write(result);
    return arg0;
  }
  ps.failureFlag = true;
  return Error("prnv requires a value and a template string");
}

// Implements the "printv" builtin function
RyeObject printvBuiltin(ProgramState ps, RyeObject? arg0, RyeObject? arg1, RyeObject? arg2, RyeObject? arg3, RyeObject? arg4) {
  if (arg0 != null && arg1 is RyeString) {
    String value = arg0.print(ps.idx);
    String template = arg1.value;
    String result = template.replaceAll("{}", value);
    
    stdout.writeln(result);
    return arg0;
  }
  ps.failureFlag = true;
  return Error("printv requires a value and a template string");
}

// Implements the "probe" builtin function
RyeObject probeBuiltin(ProgramState ps, RyeObject? arg0, RyeObject? arg1, RyeObject? arg2, RyeObject? arg3, RyeObject? arg4) {
  if (arg0 != null) {
    // Print detailed information about the value
    stdout.writeln(arg0.inspect(ps.idx));
    return arg0;
  }
  ps.failureFlag = true;
  return Error("probe requires an argument");
}

// Implements the "inspect" builtin function
RyeObject inspectBuiltin(ProgramState ps, RyeObject? arg0, RyeObject? arg1, RyeObject? arg2, RyeObject? arg3, RyeObject? arg4) {
  if (arg0 != null) {
    // Return detailed information about the value as a string
    return RyeString(arg0.inspect(ps.idx));
  }
  ps.failureFlag = true;
  return Error("inspect requires an argument");
}

// --- Registration ---

void registerPrintingBuiltins(ProgramState ps) {
  // Register printing builtins
  ps.ctx.set(ps.idx.indexWord("prns"), 
    Builtin(prnsBuiltin, 1, false, false, "Prints a value followed by a space, returning the input value"));
  
  ps.ctx.set(ps.idx.indexWord("prn"), 
    Builtin(prnBuiltin, 1, false, false, "Prints a value without adding a newline, returning the input value"));
  
  ps.ctx.set(ps.idx.indexWord("print"), 
    Builtin(printBuiltin, 1, false, false, "Prints a value followed by a newline, returning the input value"));
  
  ps.ctx.set(ps.idx.indexWord("print2"), 
    Builtin(print2Builtin, 2, false, false, "Prints two values separated by a space and followed by a newline, returning the second value"));
  
  ps.ctx.set(ps.idx.indexWord("format"), 
    Builtin(formatBuiltin, 2, false, true, "Formats a value according to format specifiers, returning the formatted string"));
  
  ps.ctx.set(ps.idx.indexWord("prnf"), 
    Builtin(prnfBuiltin, 2, false, false, "Formats a value according to format specifiers and prints it without a newline, returning the input value"));
  
  ps.ctx.set(ps.idx.indexWord("embed"), 
    Builtin(embedBuiltin, 2, false, true, "Embeds a value into a string or URI by replacing {} placeholder with the string representation of the value"));
  
  ps.ctx.set(ps.idx.indexWord("prnv"), 
    Builtin(prnvBuiltin, 2, false, false, "Embeds a value into a string by replacing {} placeholder and prints it without a newline, returning the input value"));
  
  ps.ctx.set(ps.idx.indexWord("printv"), 
    Builtin(printvBuiltin, 2, false, false, "Embeds a value into a string by replacing {} placeholder and prints it followed by a newline, returning the input value"));
  
  ps.ctx.set(ps.idx.indexWord("probe"), 
    Builtin(probeBuiltin, 1, false, false, "Prints detailed type and value information about a value, followed by a newline, returning the input value"));
  
  ps.ctx.set(ps.idx.indexWord("inspect"), 
    Builtin(inspectBuiltin, 1, false, true, "Returns a string containing detailed type and value information about a value"));
}
