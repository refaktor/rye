// types.dart - Additional types for the Dart implementation of Rye

import 'dart:math';
import 'package:intl/intl.dart';
import 'rye.dart';

// Boolean implementation
class Boolean implements RyeObject {
  final bool value;

  const Boolean(this.value);

  @override
  RyeType type() => RyeType.booleanType;

  @override
  String print(Idxs idxs) {
    return value ? "true" : "false";
  }

  @override
  String inspect(Idxs idxs) {
    return '[Boolean: ${print(idxs)}]';
  }

  @override
  bool equal(RyeObject other) {
    if (other.type() != RyeType.booleanType) return false;
    return value == (other as Boolean).value;
  }

  @override
  int getKind() => RyeType.booleanType.index;
}

// Decimal implementation
class Decimal implements RyeObject {
  final double value;

  const Decimal(this.value);

  @override
  RyeType type() => RyeType.decimalType;

  @override
  String print(Idxs idxs) {
    return value.toString();
  }

  @override
  String inspect(Idxs idxs) {
    return '[Decimal: ${print(idxs)}]';
  }

  @override
  bool equal(RyeObject other) {
    if (other.type() != RyeType.decimalType) return false;
    
    // Use epsilon for floating point comparison
    const epsilon = 0.0000000000001;
    return (value - (other as Decimal).value).abs() <= epsilon;
  }

  @override
  int getKind() => RyeType.decimalType.index;
}

// Time implementation
class Time implements RyeObject {
  final DateTime value;

  Time(this.value);

  @override
  RyeType type() => RyeType.timeType;

  @override
  String print(Idxs idxs) {
    return value.toIso8601String();
  }

  @override
  String inspect(Idxs idxs) {
    return '[Time: ${print(idxs)}]';
  }

  @override
  bool equal(RyeObject other) {
    if (other.type() != RyeType.timeType) return false;
    return value.isAtSameMomentAs((other as Time).value);
  }

  @override
  int getKind() => RyeType.timeType.index;
}

// Date implementation
class Date implements RyeObject {
  final DateTime value;

  Date(this.value);

  @override
  RyeType type() => RyeType.dateType;

  @override
  String print(Idxs idxs) {
    return DateFormat('yyyy-MM-dd').format(value);
  }

  @override
  String inspect(Idxs idxs) {
    return '[Date: ${print(idxs)}]';
  }

  @override
  bool equal(RyeObject other) {
    if (other.type() != RyeType.dateType) return false;
    
    DateTime otherDate = (other as Date).value;
    return value.year == otherDate.year && 
           value.month == otherDate.month && 
           value.day == otherDate.day;
  }

  @override
  int getKind() => RyeType.dateType.index;
}

// Uri implementation
class Uri implements RyeObject {
  final String scheme;
  final String path;
  final Word kind;

  Uri(this.scheme, this.path, this.kind);

  factory Uri.fromString(Idxs idxs, String uriString) {
    if (uriString.startsWith('%')) {
      // File URI
      String path = uriString.substring(1);
      String scheme = "file";
      int kindIdx = idxs.indexWord("$scheme-schema");
      return Uri(scheme, path, Word(kindIdx));
    } else if (uriString.contains('://')) {
      // Standard URI
      List<String> parts = uriString.split('://');
      String scheme = parts[0];
      String path = parts[1];
      int kindIdx = idxs.indexWord("$scheme-schema");
      return Uri(scheme, path, Word(kindIdx));
    } else {
      // Default to file URI
      String scheme = "file";
      int kindIdx = idxs.indexWord("$scheme-schema");
      return Uri(scheme, uriString, Word(kindIdx));
    }
  }

  @override
  RyeType type() => RyeType.uriType;

  @override
  String print(Idxs idxs) {
    if (scheme == "file") {
      return "%$path";
    }
    return "$scheme://$path";
  }

  @override
  String inspect(Idxs idxs) {
    return '[Uri: ${print(idxs)}]';
  }

  @override
  bool equal(RyeObject other) {
    if (other.type() != RyeType.uriType) return false;
    
    Uri otherUri = other as Uri;
    return scheme == otherUri.scheme && 
           path == otherUri.path && 
           kind.equal(otherUri.kind);
  }

  @override
  int getKind() => kind.index;
}

// Email implementation
class Email implements RyeObject {
  final String address;

  Email(this.address);

  @override
  RyeType type() => RyeType.emailType;

  @override
  String print(Idxs idxs) {
    return address;
  }

  @override
  String inspect(Idxs idxs) {
    return '[Email: ${print(idxs)}]';
  }

  @override
  bool equal(RyeObject other) {
    if (other.type() != RyeType.emailType) return false;
    return address == (other as Email).address;
  }

  @override
  int getKind() => RyeType.emailType.index;
}

// Vector implementation
class Vector implements RyeObject {
  final List<double> values;
  final Word kind;

  Vector(this.values, [this.kind = const Word(0)]);

  @override
  RyeType type() => RyeType.vectorType;

  @override
  String print(Idxs idxs) {
    return "V[${values.join(', ')}]";
  }

  @override
  String inspect(Idxs idxs) {
    double norm = calculateNorm();
    double mean = calculateMean();
    return '[Vector: Len ${values.length} Norm ${norm.toStringAsFixed(2)} Mean ${mean.toStringAsFixed(2)}]';
  }

  @override
  bool equal(RyeObject other) {
    if (other.type() != RyeType.vectorType) return false;
    
    Vector otherVector = other as Vector;
    if (values.length != otherVector.values.length) return false;
    
    for (int i = 0; i < values.length; i++) {
      if (values[i] != otherVector.values[i]) {
        return false;
      }
    }
    
    return true;
  }

  @override
  int getKind() => RyeType.vectorType.index;
  
  // Helper methods for vector operations
  double calculateNorm() {
    double sumOfSquares = 0;
    for (double value in values) {
      sumOfSquares += value * value;
    }
    return sqrt(sumOfSquares);
  }
  
  double calculateMean() {
    if (values.isEmpty) return 0;
    double sum = 0;
    for (double value in values) {
      sum += value;
    }
    return sum / values.length;
  }
}

// String implementation
class RyeString implements RyeObject {
  final String value;

  const RyeString(this.value);

  @override
  RyeType type() => RyeType.stringType;

  @override
  String print(Idxs idxs) {
    return '"$value"';
  }

  @override
  String inspect(Idxs idxs) {
    return '[String: "$value"]';
  }

  @override
  bool equal(RyeObject other) {
    if (other.type() != RyeType.stringType) return false;
    return value == (other as RyeString).value;
  }

  @override
  int getKind() => RyeType.stringType.index;
}

// List implementation
class RyeList implements RyeObject {
  final List<RyeObject> items;

  RyeList(this.items);

  @override
  RyeType type() => RyeType.listType;

  @override
  String print(Idxs idxs) {
    StringBuffer buffer = StringBuffer();
    buffer.write('[');
    for (int i = 0; i < items.length; i++) {
      buffer.write(items[i].print(idxs));
      if (i < items.length - 1) {
        buffer.write(' ');
      }
    }
    buffer.write(']');
    return buffer.toString();
  }

  @override
  String inspect(Idxs idxs) {
    StringBuffer buffer = StringBuffer();
    buffer.write('[List: ');
    for (int i = 0; i < items.length; i++) {
      buffer.write(items[i].inspect(idxs));
      if (i < items.length - 1) {
        buffer.write(' ');
      }
    }
    buffer.write(']');
    return buffer.toString();
  }

  @override
  bool equal(RyeObject other) {
    if (other.type() != RyeType.listType) return false;
    
    RyeList otherList = other as RyeList;
    if (items.length != otherList.items.length) return false;
    
    for (int i = 0; i < items.length; i++) {
      if (!items[i].equal(otherList.items[i])) {
        return false;
      }
    }
    
    return true;
  }

  @override
  int getKind() => RyeType.listType.index;
  
  // Additional methods for list operations
  RyeObject get(int index) {
    if (index < 0 || index >= items.length) {
      return Error("List index out of bounds: $index");
    }
    return items[index];
  }
  
  void add(RyeObject item) {
    items.add(item);
  }
  
  RyeObject remove(int index) {
    if (index < 0 || index >= items.length) {
      return Error("List index out of bounds: $index");
    }
    return items.removeAt(index);
  }
  
  int length() {
    return items.length;
  }
}

// Dict implementation
class RyeDict implements RyeObject {
  final Map<String, RyeObject> entries;

  RyeDict(this.entries);

  @override
  RyeType type() => RyeType.dictType;

  @override
  String print(Idxs idxs) {
    StringBuffer buffer = StringBuffer();
    buffer.write('{');
    bool first = true;
    entries.forEach((key, value) {
      if (!first) {
        buffer.write(' ');
      }
      buffer.write('"$key": ${value.print(idxs)}');
      first = false;
    });
    buffer.write('}');
    return buffer.toString();
  }

  @override
  String inspect(Idxs idxs) {
    StringBuffer buffer = StringBuffer();
    buffer.write('[Dict: ');
    bool first = true;
    entries.forEach((key, value) {
      if (!first) {
        buffer.write(' ');
      }
      buffer.write('"$key": ${value.inspect(idxs)}');
      first = false;
    });
    buffer.write(']');
    return buffer.toString();
  }

  @override
  bool equal(RyeObject other) {
    if (other.type() != RyeType.dictType) return false;
    
    RyeDict otherDict = other as RyeDict;
    if (entries.length != otherDict.entries.length) return false;
    
    for (String key in entries.keys) {
      if (!otherDict.entries.containsKey(key)) {
        return false;
      }
      
      if (!entries[key]!.equal(otherDict.entries[key]!)) {
        return false;
      }
    }
    
    return true;
  }

  @override
  int getKind() => RyeType.dictType.index;
  
  // Additional methods for dict operations
  RyeObject get(String key) {
    if (!entries.containsKey(key)) {
      return Error("Key not found in dict: $key");
    }
    return entries[key]!;
  }
  
  void set(String key, RyeObject value) {
    entries[key] = value;
  }
  
  RyeObject remove(String key) {
    if (!entries.containsKey(key)) {
      return Error("Key not found in dict: $key");
    }
    return entries.remove(key)!;
  }
  
  bool has(String key) {
    return entries.containsKey(key);
  }
  
  List<String> keys() {
    return entries.keys.toList();
  }
  
  int length() {
    return entries.length;
  }
}

// Context implementation as a RyeObject
class RyeContext implements RyeObject {
  final RyeCtx ctx;

  RyeContext(this.ctx);

  @override
  RyeType type() => RyeType.contextType;

  @override
  String print(Idxs idxs) {
    return '[Context]';
  }

  @override
  String inspect(Idxs idxs) {
    StringBuffer buffer = StringBuffer();
    buffer.write('[Context: ');
    
    // Print the words in the context
    bool first = true;
    ctx.state.forEach((wordIdx, value) {
      if (!first) {
        buffer.write(' ');
      }
      buffer.write('${idxs.getWord(wordIdx)}: ${value.print(idxs)}');
      first = false;
    });
    
    buffer.write(']');
    return buffer.toString();
  }

  @override
  bool equal(RyeObject other) {
    if (other.type() != RyeType.contextType) return false;
    return ctx == (other as RyeContext).ctx; // Reference equality
  }

  @override
  int getKind() => RyeType.contextType.index;
}

// Builtins for the new types
RyeObject makeStringBuiltin(ProgramState ps, RyeObject? arg0, RyeObject? arg1, RyeObject? arg2, RyeObject? arg3, RyeObject? arg4) {
  if (arg0 == null) {
    ps.failureFlag = true;
    return Error("make-string requires an argument");
  }
  
  return RyeString(arg0.print(ps.idx));
}

RyeObject makeListBuiltin(ProgramState ps, RyeObject? arg0, RyeObject? arg1, RyeObject? arg2, RyeObject? arg3, RyeObject? arg4) {
  List<RyeObject> items = [];
  
  // If arg0 is provided, it should be a block of items
  if (arg0 != null && arg0 is Block) {
    // Create a new program state for the block
    ProgramState blockPs = ProgramState(TSeries(List<RyeObject>.from(arg0.series.s)), ps.idx);
    blockPs.ctx = ps.ctx;
    
    // Reset the series position to ensure we start from the beginning
    blockPs.ser.reset();
    
    // Evaluate each item in the block and add it to the list
    while (blockPs.ser.getPos() < blockPs.ser.len()) {
      rye00_evalExpressionConcrete(blockPs);
      
      if (blockPs.errorFlag || blockPs.failureFlag) {
        ps.errorFlag = blockPs.errorFlag;
        ps.failureFlag = blockPs.failureFlag;
        ps.res = blockPs.res;
        return blockPs.res ?? Error("Error in make-list");
      }
      
      if (blockPs.res != null) {
        items.add(blockPs.res!);
      }
    }
  }
  
  return RyeList(items);
}

RyeObject makeDictBuiltin(ProgramState ps, RyeObject? arg0, RyeObject? arg1, RyeObject? arg2, RyeObject? arg3, RyeObject? arg4) {
  Map<String, RyeObject> entries = {};
  
  // If arg0 is provided, it should be a block of key-value pairs
  if (arg0 != null && arg0 is Block) {
    // Create a new program state for the block
    ProgramState blockPs = ProgramState(TSeries(List<RyeObject>.from(arg0.series.s)), ps.idx);
    blockPs.ctx = ps.ctx;
    
    // Reset the series position to ensure we start from the beginning
    blockPs.ser.reset();
    
    // Evaluate key-value pairs in the block and add them to the dict
    while (blockPs.ser.getPos() < blockPs.ser.len()) {
      // Evaluate the key
      rye00_evalExpressionConcrete(blockPs);
      
      if (blockPs.errorFlag || blockPs.failureFlag) {
        ps.errorFlag = blockPs.errorFlag;
        ps.failureFlag = blockPs.failureFlag;
        ps.res = blockPs.res;
        return blockPs.res ?? Error("Error in make-dict (key)");
      }
      
      if (blockPs.res == null) {
        ps.failureFlag = true;
        return Error("make-dict: key cannot be null");
      }
      
      String key = blockPs.res!.print(ps.idx);
      
      // Evaluate the value
      rye00_evalExpressionConcrete(blockPs);
      
      if (blockPs.errorFlag || blockPs.failureFlag) {
        ps.errorFlag = blockPs.errorFlag;
        ps.failureFlag = blockPs.failureFlag;
        ps.res = blockPs.res;
        return blockPs.res ?? Error("Error in make-dict (value)");
      }
      
      if (blockPs.res == null) {
        ps.failureFlag = true;
        return Error("make-dict: value cannot be null");
      }
      
      entries[key] = blockPs.res!;
    }
  }
  
  return RyeDict(entries);
}

RyeObject makeContextBuiltin(ProgramState ps, RyeObject? arg0, RyeObject? arg1, RyeObject? arg2, RyeObject? arg3, RyeObject? arg4) {
  // Create a new context with the current context as parent
  RyeCtx ctx = RyeCtx(ps.ctx);
  
  // If arg0 is provided, it should be a block of key-value pairs
  if (arg0 != null && arg0 is Block) {
    // Create a new program state for the block
    ProgramState blockPs = ProgramState(TSeries(List<RyeObject>.from(arg0.series.s)), ps.idx);
    blockPs.ctx = ps.ctx;
    
    // Reset the series position to ensure we start from the beginning
    blockPs.ser.reset();
    
    // Evaluate key-value pairs in the block and add them to the context
    while (blockPs.ser.getPos() < blockPs.ser.len()) {
      // Evaluate the key (should be a word)
      rye00_evalExpressionConcrete(blockPs);
      
      if (blockPs.errorFlag || blockPs.failureFlag) {
        ps.errorFlag = blockPs.errorFlag;
        ps.failureFlag = blockPs.failureFlag;
        ps.res = blockPs.res;
        return blockPs.res ?? Error("Error in make-context (key)");
      }
      
      if (blockPs.res == null || blockPs.res!.type() != RyeType.wordType) {
        ps.failureFlag = true;
        return Error("make-context: key must be a word");
      }
      
      int wordIdx = (blockPs.res as Word).index;
      
      // Evaluate the value
      rye00_evalExpressionConcrete(blockPs);
      
      if (blockPs.errorFlag || blockPs.failureFlag) {
        ps.errorFlag = blockPs.errorFlag;
        ps.failureFlag = blockPs.failureFlag;
        ps.res = blockPs.res;
        return blockPs.res ?? Error("Error in make-context (value)");
      }
      
      if (blockPs.res == null) {
        ps.failureFlag = true;
        return Error("make-context: value cannot be null");
      }
      
      ctx.set(wordIdx, blockPs.res!);
    }
  }
  
  return RyeContext(ctx);
}

// Builtins for the new types
RyeObject makeBooleanBuiltin(ProgramState ps, RyeObject? arg0, RyeObject? arg1, RyeObject? arg2, RyeObject? arg3, RyeObject? arg4) {
  if (arg0 == null) {
    ps.failureFlag = true;
    return Error("make-boolean requires an argument");
  }
  
  if (arg0 is Boolean) {
    return arg0;
  }
  
  String value = arg0.print(ps.idx).toLowerCase();
  return Boolean(value == "true" || value == "yes" || value == "1");
}

RyeObject makeDecimalBuiltin(ProgramState ps, RyeObject? arg0, RyeObject? arg1, RyeObject? arg2, RyeObject? arg3, RyeObject? arg4) {
  if (arg0 == null) {
    ps.failureFlag = true;
    return Error("make-decimal requires an argument");
  }
  
  if (arg0 is Decimal) {
    return arg0;
  }
  
  if (arg0 is Integer) {
    return Decimal(arg0.value.toDouble());
  }
  
  try {
    double value = double.parse(arg0.print(ps.idx));
    return Decimal(value);
  } catch (e) {
    ps.failureFlag = true;
    return Error("Cannot convert to decimal: ${arg0.print(ps.idx)}");
  }
}

RyeObject makeDateBuiltin(ProgramState ps, RyeObject? arg0, RyeObject? arg1, RyeObject? arg2, RyeObject? arg3, RyeObject? arg4) {
  if (arg0 == null) {
    ps.failureFlag = true;
    return Error("make-date requires an argument");
  }
  
  if (arg0 is Date) {
    return arg0;
  }
  
  if (arg0 is Time) {
    return Date((arg0 as Time).value);
  }
  
  try {
    DateTime value = DateTime.parse(arg0.print(ps.idx).replaceAll('"', ''));
    return Date(value);
  } catch (e) {
    ps.failureFlag = true;
    return Error("Cannot convert to date: ${arg0.print(ps.idx)}");
  }
}

RyeObject makeTimeBuiltin(ProgramState ps, RyeObject? arg0, RyeObject? arg1, RyeObject? arg2, RyeObject? arg3, RyeObject? arg4) {
  if (arg0 == null) {
    ps.failureFlag = true;
    return Error("make-time requires an argument");
  }
  
  if (arg0 is Time) {
    return arg0;
  }
  
  try {
    DateTime value = DateTime.parse(arg0.print(ps.idx).replaceAll('"', ''));
    return Time(value);
  } catch (e) {
    ps.failureFlag = true;
    return Error("Cannot convert to time: ${arg0.print(ps.idx)}");
  }
}

RyeObject makeUriBuiltin(ProgramState ps, RyeObject? arg0, RyeObject? arg1, RyeObject? arg2, RyeObject? arg3, RyeObject? arg4) {
  if (arg0 == null) {
    ps.failureFlag = true;
    return Error("make-uri requires an argument");
  }
  
  if (arg0 is Uri) {
    return arg0;
  }
  
  String uriString = arg0.print(ps.idx).replaceAll('"', '');
  return Uri.fromString(ps.idx, uriString);
}

RyeObject makeEmailBuiltin(ProgramState ps, RyeObject? arg0, RyeObject? arg1, RyeObject? arg2, RyeObject? arg3, RyeObject? arg4) {
  if (arg0 == null) {
    ps.failureFlag = true;
    return Error("make-email requires an argument");
  }
  
  if (arg0 is Email) {
    return arg0;
  }
  
  String emailString = arg0.print(ps.idx).replaceAll('"', '');
  
  // Simple email validation
  if (!emailString.contains('@') || !emailString.contains('.')) {
    ps.failureFlag = true;
    return Error("Invalid email address: $emailString");
  }
  
  return Email(emailString);
}

RyeObject makeVectorBuiltin(ProgramState ps, RyeObject? arg0, RyeObject? arg1, RyeObject? arg2, RyeObject? arg3, RyeObject? arg4) {
  List<double> values = [];
  
  // If arg0 is provided, it should be a block of numbers
  if (arg0 != null && arg0 is Block) {
    // Create a new program state for the block
    ProgramState blockPs = ProgramState(TSeries(List<RyeObject>.from(arg0.series.s)), ps.idx);
    blockPs.ctx = ps.ctx;
    
    // Reset the series position to ensure we start from the beginning
    blockPs.ser.reset();
    
    // Evaluate each item in the block and add it to the vector
    while (blockPs.ser.getPos() < blockPs.ser.len()) {
      rye00_evalExpressionConcrete(blockPs);
      
      if (blockPs.errorFlag || blockPs.failureFlag) {
        ps.errorFlag = blockPs.errorFlag;
        ps.failureFlag = blockPs.failureFlag;
        ps.res = blockPs.res;
        return blockPs.res ?? Error("Error in make-vector");
      }
      
      if (blockPs.res != null) {
        if (blockPs.res is Integer) {
          values.add((blockPs.res as Integer).value.toDouble());
        } else if (blockPs.res is Decimal) {
          values.add((blockPs.res as Decimal).value);
        } else {
          ps.failureFlag = true;
          return Error("Vector elements must be numbers");
        }
      }
    }
  }
  
  return Vector(values);
}

// Implements the "make-function" builtin
RyeObject makeFunctionBuiltin(ProgramState ps, RyeObject? arg0, RyeObject? arg1, RyeObject? arg2, RyeObject? arg3, RyeObject? arg4) {
  if (arg0 is Block && arg1 is Block) {
    // TODO: Add support for flags like //pure, //in-ctx if parser supports them
    // Capture the current context (ps.ctx) for closure
    return RyeFunction(arg0, arg1, ps.ctx); // Use RyeFunction here
  }
  ps.failureFlag = true;
  return Error("make-function expects a spec block and a body block");
}


// Register the new type builtins
void registerTypeBuiltins(ProgramState ps) {
  // Register the make-string builtin
  int makeStringIdx = ps.idx.indexWord("make-string");
  Builtin makeStringBuiltinObj = Builtin(makeStringBuiltin, 1, false, true, "Creates a string from a value");
  ps.ctx.set(makeStringIdx, makeStringBuiltinObj);
  
  // Register the make-list builtin
  int makeListIdx = ps.idx.indexWord("make-list");
  Builtin makeListBuiltinObj = Builtin(makeListBuiltin, 1, false, true, "Creates a list from a block of values");
  ps.ctx.set(makeListIdx, makeListBuiltinObj);
  
  // Register the make-dict builtin
  int makeDictIdx = ps.idx.indexWord("make-dict");
  Builtin makeDictBuiltinObj = Builtin(makeDictBuiltin, 1, false, true, "Creates a dictionary from a block of key-value pairs");
  ps.ctx.set(makeDictIdx, makeDictBuiltinObj);
  
  // Register the make-context builtin
  int makeContextIdx = ps.idx.indexWord("make-context");
  Builtin makeContextBuiltinObj = Builtin(makeContextBuiltin, 1, false, true, "Creates a context from a block of key-value pairs");
  ps.ctx.set(makeContextIdx, makeContextBuiltinObj);
  
  // Register the make-boolean builtin
  int makeBooleanIdx = ps.idx.indexWord("make-boolean");
  Builtin makeBooleanBuiltinObj = Builtin(makeBooleanBuiltin, 1, false, true, "Creates a boolean from a value");
  ps.ctx.set(makeBooleanIdx, makeBooleanBuiltinObj);
  
  // Register the make-decimal builtin
  int makeDecimalIdx = ps.idx.indexWord("make-decimal");
  Builtin makeDecimalBuiltinObj = Builtin(makeDecimalBuiltin, 1, false, true, "Creates a decimal from a value");
  ps.ctx.set(makeDecimalIdx, makeDecimalBuiltinObj);
  
  // Register the make-date builtin
  int makeDateIdx = ps.idx.indexWord("make-date");
  Builtin makeDateBuiltinObj = Builtin(makeDateBuiltin, 1, false, true, "Creates a date from a value");
  ps.ctx.set(makeDateIdx, makeDateBuiltinObj);
  
  // Register the make-time builtin
  int makeTimeIdx = ps.idx.indexWord("make-time");
  Builtin makeTimeBuiltinObj = Builtin(makeTimeBuiltin, 1, false, true, "Creates a time from a value");
  ps.ctx.set(makeTimeIdx, makeTimeBuiltinObj);
  
  // Register the make-uri builtin
  int makeUriIdx = ps.idx.indexWord("make-uri");
  Builtin makeUriBuiltinObj = Builtin(makeUriBuiltin, 1, false, true, "Creates a URI from a value");
  ps.ctx.set(makeUriIdx, makeUriBuiltinObj);
  
  // Register the make-email builtin
  int makeEmailIdx = ps.idx.indexWord("make-email");
  Builtin makeEmailBuiltinObj = Builtin(makeEmailBuiltin, 1, false, true, "Creates an email from a value");
  ps.ctx.set(makeEmailIdx, makeEmailBuiltinObj);
  
  // Register the make-vector builtin
  int makeVectorIdx = ps.idx.indexWord("make-vector");
  Builtin makeVectorBuiltinObj = Builtin(makeVectorBuiltin, 1, false, true, "Creates a vector from a block of numbers");
  ps.ctx.set(makeVectorIdx, makeVectorBuiltinObj);

  // Register the make-function builtin
  int makeFuncIdx = ps.idx.indexWord("make-function");
  Builtin makeFuncBuiltinObj = Builtin(makeFunctionBuiltin, 2, false, true, "Creates a user-defined function");
  ps.ctx.set(makeFuncIdx, makeFuncBuiltinObj);
}
