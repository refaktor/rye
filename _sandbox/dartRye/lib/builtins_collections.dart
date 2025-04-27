// builtins_collections.dart - Collection builtins (List, Dict)

import 'rye.dart';
import 'types.dart';

// --- List Builtins ---

// Implements the "get" builtin for lists
RyeObject listGetBuiltin(ProgramState ps, RyeObject? arg0, RyeObject? arg1, RyeObject? arg2, RyeObject? arg3, RyeObject? arg4) {
  if (arg0 is RyeList && arg1 is Integer) {
    int index = arg1.value;
    if (index >= 0 && index < arg0.items.length) {
      return arg0.items[index];
    } else {
      ps.failureFlag = true;
      return Error("List index out of bounds: $index");
    }
  }
  ps.failureFlag = true;
  return Error("get (list) expects a List and an Integer index");
}

// Implements the "length" builtin for lists
RyeObject listLengthBuiltin(ProgramState ps, RyeObject? arg0, RyeObject? arg1, RyeObject? arg2, RyeObject? arg3, RyeObject? arg4) {
  if (arg0 is RyeList) {
    return Integer(arg0.items.length.toInt());
  }
  ps.failureFlag = true;
  return Error("length (list) expects a List");
}

// Implements the "append" builtin for lists (modifies in place)
RyeObject listAppendBuiltin(ProgramState ps, RyeObject? arg0, RyeObject? arg1, RyeObject? arg2, RyeObject? arg3, RyeObject? arg4) {
  if (arg0 is RyeList && arg1 != null) {
    arg0.items.add(arg1);
    return arg0; // Return the modified list
  }
  ps.failureFlag = true;
  return Error("append (list) expects a List and a value to append");
}

// --- Dict Builtins --- 

// Implements the "get" builtin for dicts
RyeObject dictGetBuiltin(ProgramState ps, RyeObject? arg0, RyeObject? arg1, RyeObject? arg2, RyeObject? arg3, RyeObject? arg4) {
  if (arg0 is RyeDict && arg1 is RyeString) {
    String key = arg1.value; // Use raw string value as key
    if (arg0.entries.containsKey(key)) {
      return arg0.entries[key]!;
    } else {
      ps.failureFlag = true;
      return Error("Key not found in dict: $key");
    }
  }
  ps.failureFlag = true;
  return Error("get (dict) expects a Dict and a String key");
}

// Implements the "length" builtin for dicts
RyeObject dictLengthBuiltin(ProgramState ps, RyeObject? arg0, RyeObject? arg1, RyeObject? arg2, RyeObject? arg3, RyeObject? arg4) {
  if (arg0 is RyeDict) {
    return Integer(arg0.entries.length.toInt());
  }
  ps.failureFlag = true;
  return Error("length (dict) expects a Dict");
}

// Implements the "set" builtin for dicts (modifies in place)
RyeObject dictSetBuiltin(ProgramState ps, RyeObject? arg0, RyeObject? arg1, RyeObject? arg2, RyeObject? arg3, RyeObject? arg4) {
  if (arg0 is RyeDict && arg1 is RyeString && arg2 != null) {
     String key = arg1.value;
     arg0.entries[key] = arg2;
     return arg0; // Return the modified dict
  }
  ps.failureFlag = true;
  return Error("set (dict) expects a Dict, a String key, and a value");
}

// TODO: Implement other Dict builtins (keys, values, has? etc.)

// --- Registration ---

void registerCollectionBuiltins(ProgramState ps) {
  // Register list builtins generically
  int getIdx = ps.idx.indexWord("get");
  int lengthIdx = ps.idx.indexWord("length");
  int appendIdx = ps.idx.indexWord("append");
  int setIdx = ps.idx.indexWord("set"); // Assuming 'set' might also apply to lists later

  ps.registerGeneric(RyeType.listType.index, getIdx, 
    Builtin(listGetBuiltin, 2, false, true, "Gets item from list at index"));
    
  ps.registerGeneric(RyeType.listType.index, lengthIdx, 
    Builtin(listLengthBuiltin, 1, false, true, "Returns the length of a list"));
    
  ps.registerGeneric(RyeType.listType.index, appendIdx, 
    Builtin(listAppendBuiltin, 2, false, false, "Appends an item to a list (modifies in place)"));

  // Register dict builtins generically
  ps.registerGeneric(RyeType.dictType.index, getIdx, 
    Builtin(dictGetBuiltin, 2, false, true, "Gets value from dict for a given key"));

  ps.registerGeneric(RyeType.dictType.index, lengthIdx, 
    Builtin(dictLengthBuiltin, 1, false, true, "Returns the number of key-value pairs in a dict"));

  ps.registerGeneric(RyeType.dictType.index, setIdx, 
    Builtin(dictSetBuiltin, 3, false, false, "Sets a key-value pair in a dict (modifies in place)"));

  // TODO: Register other Dict builtins (keys, values, has?)
}
