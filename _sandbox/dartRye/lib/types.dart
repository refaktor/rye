// types.dart - Additional types for the Dart implementation of Rye

import 'rye.dart';

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
}
