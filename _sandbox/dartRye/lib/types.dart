// types.dart - Core data types for Dart Rye evaluator

// dart:core is imported automatically
import 'env.dart'; // Import env for RyeCtx and ProgramState
import 'idxs.dart'; // Import idxs for Idxs
// Remove Comma from loader import to break circular dependency
import 'loader.dart' show Setword, OpWord, PipeWord, LSetword, LModword, CPath, LitWord; 

// Type enum for Rye values
enum RyeType {
  blockType,
  integerType,
  wordType,
  builtinType,
  errorType,
  voidType,
  stringType,
  listType,
  dictType,
  contextType,
  booleanType,
  decimalType,
  timeType,
  dateType,
  uriType,
  emailType,
  vectorType,
  commaType, // Added commaType
}

// Base class for all Rye values
abstract class RyeObject {
  RyeType type();
  String print(Idxs idxs);
  String inspect(Idxs idxs);
  bool equal(RyeObject other);
  int getKind();
}

// Integer implementation
class Integer implements RyeObject {
  int value;

  Integer(this.value);

  @override
  RyeType type() => RyeType.integerType;

  @override
  String print(Idxs idxs) {
    return value.toString();
  }

  @override
  String inspect(Idxs idxs) {
    return '[Integer: ${print(idxs)}]';
  }

  @override
  bool equal(RyeObject other) {
    if (other.type() != RyeType.integerType) return false;
    return value == (other as Integer).value;
  }

  @override
  int getKind() => RyeType.integerType.index;
}

// Word implementation
class Word implements RyeObject {
  final int index;

  const Word(this.index);

  @override
  RyeType type() => RyeType.wordType;

  @override
  String print(Idxs idxs) {
    return idxs.getWord(index);
  }

  @override
  String inspect(Idxs idxs) {
    return '[Word: ${print(idxs)}]';
  }

  @override
  bool equal(RyeObject other) {
    if (other.type() != RyeType.wordType) return false;
    return index == (other as Word).index;
  }

  @override
  int getKind() => RyeType.wordType.index;
}

// Error implementation
class Error implements RyeObject {
  String message;
  int status;
  Error? parent;
  RyeCtx? codeContext; // Context where the error occurred
  int? position;      // Position in the series where the error occurred

  Error(this.message, {this.status = 0, this.parent, this.codeContext, this.position});

  @override
  RyeType type() => RyeType.errorType;

  @override
  String print(Idxs idxs) {
    String statusStr = status != 0 ? "($status)" : "";
    StringBuffer b = StringBuffer();
    b.write("Error$statusStr: $message ");
    
    if (parent != null) {
      b.write("\n  ${parent!.print(idxs)}");
    }
    
    // Optionally add position info if available
    if (position != null) {
       b.write(" (at pos $position)");
    }
    
    return b.toString();
  }

  @override
  String inspect(Idxs idxs) {
    return '[${print(idxs)}]';
  }

  @override
  bool equal(RyeObject other) {
    if (other.type() != RyeType.errorType) return false;
    
    Error otherError = other as Error;
    if (status != otherError.status) return false;
    if (message != otherError.message) return false;
    
    // Check if both have parent or both don't have parent
    if ((parent == null) != (otherError.parent == null)) return false;
    
    // If both have parent, check if parents are equal
    if (parent != null && otherError.parent != null) {
      if (!parent!.equal(otherError.parent!)) return false;
    }
    
    return true;
  }

  @override
  int getKind() => RyeType.errorType.index;
}

// Void implementation
class Void implements RyeObject {
  const Void();

  @override
  RyeType type() => RyeType.voidType;

  @override
  String print(Idxs idxs) {
    return '_';
  }

  @override
  String inspect(Idxs idxs) {
    return '[Void]';
  }

  @override
  bool equal(RyeObject other) {
    return other.type() == RyeType.voidType;
  }

  @override
  int getKind() => RyeType.voidType.index;
}

// String implementation
class RyeString implements RyeObject {
  String value;

  RyeString(this.value);

  @override
  RyeType type() => RyeType.stringType;

  @override
  String print(Idxs idxs) {
    // Basic print, might need escaping for inspection
    return value;
  }

  @override
  String inspect(Idxs idxs) {
    // Add quotes for inspection
    return '"${print(idxs)}"'; 
  }

  @override
  bool equal(RyeObject other) {
    if (other.type() != RyeType.stringType) return false;
    return value == (other as RyeString).value;
  }

  @override
  int getKind() => RyeType.stringType.index;
}

// Boolean implementation
class Boolean implements RyeObject {
  final bool value; // Ensure value is final

  // Make constructor const if possible
  const Boolean(this.value); 

  @override
  RyeType type() => RyeType.booleanType;

  @override
  String print(Idxs idxs) {
    return value.toString();
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
  double value;

  Decimal(this.value);

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
    // Consider precision for double comparison if needed
    return value == (other as Decimal).value; 
  }

  @override
  int getKind() => RyeType.decimalType.index;
}

// Uri implementation (using Dart's Uri) - Renamed to RyeUri
class RyeUri implements RyeObject {
  // Store the parsed Dart Uri object
  final Uri value; // Use standard Uri type

  // Private constructor
  RyeUri._(this.value); 

  // Factory constructor using Dart's Uri.parse
  factory RyeUri.fromString(Idxs idxs, String input) { // Rename factory
    try {
      // Use Dart's built-in parser
      return RyeUri._(Uri.parse(input)); // Use standard Uri.parse
    } catch (e) {
      // Rethrow as ArgumentError or return a Rye Error object?
      // Let's rethrow for now, loader might catch it.
      throw ArgumentError("Invalid URI format: $input - $e"); 
    }
  }
  
  // Constructor to create a new Uri from parts (needed for the builtin)
  RyeUri.fromParts({String scheme = '', String path = '', String? query, String? fragment}) // Rename constructor
      : value = Uri(scheme: scheme, path: path, query: query, fragment: fragment); // Use standard Uri


  // Expose parts via getters if needed
  String get scheme => value.scheme;
  String get path => value.path;
  String get query => value.query;
  String get fragment => value.fragment;

  @override
  RyeType type() => RyeType.uriType;

  @override
  String print(Idxs idxs) {
    // Return the full string representation
    return value.toString();
  }

  @override
  String inspect(Idxs idxs) {
    // Use DartUri's toString for inspection representation
    return '[Uri: ${value.toString()}]';
  }

  @override
  bool equal(RyeObject other) {
    if (other is! RyeUri) return false; // Check against RyeUri
    // Compare the underlying Dart Uri objects
    return value == other.value;
  }

  @override
  int getKind() => RyeType.uriType.index;
}

// Email implementation (similar to Uri, store as string)
class Email implements RyeObject {
  String value;

  Email(this.value); // Simple constructor for now

  @override
  RyeType type() => RyeType.emailType;

  @override
  String print(Idxs idxs) {
    return value;
  }

  @override
  String inspect(Idxs idxs) {
    return '[Email: ${print(idxs)}]';
  }

  @override
  bool equal(RyeObject other) {
    if (other.type() != RyeType.emailType) return false;
    return value == (other as Email).value;
  }

  @override
  int getKind() => RyeType.emailType.index;
}

// Date implementation (using DateTime)
class Date implements RyeObject {
  DateTime value;

  Date(this.value);

  @override
  RyeType type() => RyeType.dateType;

  @override
  String print(Idxs idxs) {
    // Format as YYYY-MM-DD
    return "${value.year.toString().padLeft(4, '0')}-${value.month.toString().padLeft(2, '0')}-${value.day.toString().padLeft(2, '0')}";
  }

  @override
  String inspect(Idxs idxs) {
    return '[Date: ${print(idxs)}]';
  }

  @override
  bool equal(RyeObject other) {
    if (other.type() != RyeType.dateType) return false;
    // Compare year, month, day only for Date equality
    DateTime otherValue = (other as Date).value;
    return value.year == otherValue.year && 
           value.month == otherValue.month && 
           value.day == otherValue.day;
  }

  @override
  int getKind() => RyeType.dateType.index;
}

// Time implementation (using DateTime)
class Time implements RyeObject {
  DateTime value;

  Time(this.value);

  @override
  RyeType type() => RyeType.timeType;

  @override
  String print(Idxs idxs) {
    // Format as ISO 8601 string
    return value.toIso8601String(); 
  }

  @override
  String inspect(Idxs idxs) {
    return '[Time: ${print(idxs)}]';
  }

  @override
  bool equal(RyeObject other) {
    if (other.type() != RyeType.timeType) return false;
    // Direct DateTime comparison
    return value == (other as Time).value; 
  }

  @override
  int getKind() => RyeType.timeType.index;
}

// List implementation (wrapping Dart List)
class RyeList implements RyeObject {
  List<RyeObject?> value;

  RyeList(this.value);

  @override
  RyeType type() => RyeType.listType;

  @override
  String print(Idxs idxs) {
    return value.map((e) => e?.print(idxs) ?? '_').join(' ');
  }

  @override
  String inspect(Idxs idxs) {
    return '[List: ${value.map((e) => e?.inspect(idxs) ?? '[Void]').join(' ')}]';
  }

  @override
  bool equal(RyeObject other) {
    if (other.type() != RyeType.listType) return false;
    List<RyeObject?> otherValue = (other as RyeList).value;
    if (value.length != otherValue.length) return false;
    for (int i = 0; i < value.length; i++) {
      if (value[i] == null && otherValue[i] == null) continue;
      if (value[i] == null || otherValue[i] == null) return false;
      if (!value[i]!.equal(otherValue[i]!)) return false;
    }
    return true;
  }

  @override
  int getKind() => RyeType.listType.index;
}

// Dict implementation (wrapping Dart Map)
// Using String keys for simplicity, matching Go's map[string]any approach
class RyeDict implements RyeObject {
  Map<String, RyeObject?> value;

  RyeDict(this.value);

  @override
  RyeType type() => RyeType.dictType;

  @override
  String print(Idxs idxs) {
    return '{ ${value.entries.map((e) => '"${e.key}": ${e.value?.print(idxs) ?? '_'}').join(', ')} }';
  }

  @override
  String inspect(Idxs idxs) {
    return '[Dict: ${value.entries.map((e) => '"${e.key}": ${e.value?.inspect(idxs) ?? '[Void]'}').join(', ')}]';
  }

  @override
  bool equal(RyeObject other) {
    if (other.type() != RyeType.dictType) return false;
    Map<String, RyeObject?> otherValue = (other as RyeDict).value;
    if (value.length != otherValue.length) return false;
    for (var key in value.keys) {
      if (!otherValue.containsKey(key)) return false;
      if (value[key] == null && otherValue[key] == null) continue;
      if (value[key] == null || otherValue[key] == null) return false;
      if (!value[key]!.equal(otherValue[key]!)) return false;
    }
    return true;
  }

  @override
  int getKind() => RyeType.dictType.index;
}


// TSeries represents a series of Objects with a position pointer.
class TSeries {
  List<RyeObject?> s; // The underlying list of Objects
  int pos = 0; // Current position in the series

  TSeries(this.s);

  bool ended() {
    return pos > s.length;
  }

  bool atLast() {
    return pos > s.length - 1;
  }

  int getPos() {
    return pos;
  }

  void next() {
    pos++;
  }

  RyeObject? pop() {
    if (pos >= s.length) {
      return null;
    }
    RyeObject? obj = s[pos];
    pos++;
    return obj;
  }

  bool put(RyeObject obj) {
    if (pos > 0 && pos <= s.length) {
      s[pos - 1] = obj;
      return true;
    }
    return false;
  }

  TSeries append(RyeObject obj) {
    s.add(obj);
    return this;
  }

  void reset() {
    pos = 0;
  }

  void setPos(int position) {
    pos = position;
  }

  List<RyeObject?> getAll() {
    return s;
  }

  RyeObject? peek() {
    if (s.length > pos) {
      return s[pos];
    }
    return null;
  }

  RyeObject? get(int n) {
    if (n >= 0 && n < s.length) {
      return s[n];
    }
    return null;
  }

  int len() {
    return s.length;
  }
}

// Block implementation
class Block implements RyeObject {
  TSeries series;
  int mode;

  Block(this.series, [this.mode = 0]);

  @override
  RyeType type() => RyeType.blockType;

  @override
  String print(Idxs idxs) {
    StringBuffer r = StringBuffer();
    for (int i = 0; i < series.len(); i++) {
      if (series.get(i) != null) {
        r.write(series.get(i)!.print(idxs));
        r.write(' ');
      } else {
        r.write('[NIL]');
      }
    }
    return r.toString();
  }

  @override
  String inspect(Idxs idxs) {
    StringBuffer r = StringBuffer();
    r.write('[Block: ');
    for (int i = 0; i < series.len(); i++) {
      if (series.get(i) != null) {
        if (series.getPos() == i) {
          r.write('^');
        }
        r.write(series.get(i)!.inspect(idxs));
        r.write(' ');
      }
    }
    r.write(']');
    return r.toString();
  }

  @override
  bool equal(RyeObject other) {
    if (other.type() != RyeType.blockType) return false;
    
    Block otherBlock = other as Block;
    if (series.len() != otherBlock.series.len()) return false;
    if (mode != otherBlock.mode) return false;
    
    for (int j = 0; j < series.len(); j++) {
      // Handle potential null values in series comparison
      if (series.get(j) == null && otherBlock.series.get(j) == null) continue;
      if (series.get(j) == null || otherBlock.series.get(j) == null) return false;
      if (!series.get(j)!.equal(otherBlock.series.get(j)!)) {
        return false;
      }
    }
    return true;
  }

  @override
  int getKind() => RyeType.blockType.index;
}

// Builtin function type definition - Moved here temporarily, might move to builtins_base.dart later
typedef BuiltinFunction = RyeObject Function(ProgramState ps, RyeObject? arg0, RyeObject? arg1, RyeObject? arg2, RyeObject? arg3, RyeObject? arg4);

// Builtin implementation
class Builtin implements RyeObject {
  BuiltinFunction fn;
  int argsn;
  RyeObject? cur0;
  RyeObject? cur1;
  RyeObject? cur2;
  RyeObject? cur3;
  RyeObject? cur4;
  bool acceptFailure;
  bool pure;
  String doc;

  Builtin(this.fn, this.argsn, this.acceptFailure, this.pure, this.doc, 
      {this.cur0, this.cur1, this.cur2, this.cur3, this.cur4});

  @override
  RyeType type() => RyeType.builtinType;

  @override
  String print(Idxs idxs) {
    String pureStr = pure ? 'Pure ' : '';
    return '${pureStr}BFunction($argsn): $doc';
  }

  @override
  String inspect(Idxs idxs) {
    return '[${print(idxs)}]';
  }

  @override
  bool equal(RyeObject other) {
    if (other.type() != RyeType.builtinType) return false;
    
    Builtin otherBuiltin = other as Builtin;
    // Function equality check is tricky in Dart. Comparing references is the most practical approach.
    if (!identical(fn, otherBuiltin.fn)) return false; 
    if (argsn != otherBuiltin.argsn) return false;
    // Compare curried arguments - requires RyeObject.equal to handle nulls if necessary
    if (!(cur0?.equal(otherBuiltin.cur0 ?? Void()) ?? (otherBuiltin.cur0 == null))) return false;
    if (!(cur1?.equal(otherBuiltin.cur1 ?? Void()) ?? (otherBuiltin.cur1 == null))) return false;
    if (!(cur2?.equal(otherBuiltin.cur2 ?? Void()) ?? (otherBuiltin.cur2 == null))) return false;
    if (!(cur3?.equal(otherBuiltin.cur3 ?? Void()) ?? (otherBuiltin.cur3 == null))) return false;
    if (!(cur4?.equal(otherBuiltin.cur4 ?? Void()) ?? (otherBuiltin.cur4 == null))) return false;
    if (acceptFailure != otherBuiltin.acceptFailure) return false;
    if (pure != otherBuiltin.pure) return false;
    
    return true;
  }

  @override
  int getKind() => RyeType.builtinType.index;
}

/// Tagword implementation (e.g., 'word) - Similar to LitWord but distinct type?
class Tagword implements RyeObject {
  final int index;

  const Tagword(this.index);

  @override
  RyeType type() => RyeType.wordType; // Still treated as a word type for now

  @override
  String print(Idxs idxs) {
    // Represent tag-words with a leading quote in output for clarity
    return "'${idxs.getWord(index)}"; 
  }

  @override
  String inspect(Idxs idxs) {
    return "[Tagword: ${idxs.getWord(index)}]";
  }

  @override
  bool equal(RyeObject other) {
    if (other is! Tagword) return false;
    return index == other.index;
  }

  @override
  int getKind() => type().index;
}

/// Comma implementation (Expression separator)
class Comma implements RyeObject {
  const Comma();

  @override
  RyeType type() => RyeType.commaType; // Use commaType

  @override
  String print(Idxs idxs) {
    return ','; 
  }

  @override
  String inspect(Idxs idxs) {
    return '[Comma]';
  }

  @override
  bool equal(RyeObject other) {
    return other is Comma;
  }

  @override
  int getKind() => type().index;
}


// Function implementation (user-defined)
class RyeFunction implements RyeObject {
  Block spec; // Block containing argument words
  Block body; // Block containing the function code
  RyeCtx? ctx; // Context where the function was defined (for closure)
  bool pure;   // Is the function pure?
  bool inCtx;  // Should the function execute in its definition context?
  int argsn;   // Number of arguments expected

  RyeFunction(this.spec, this.body, this.ctx, {this.pure = false, this.inCtx = false})
    : argsn = spec.series.len(); // Calculate argsn based on spec length

  @override
  RyeType type() => RyeType.wordType; // Treat as word-like for now (like Go)

  @override
  String print(Idxs idxs) {
    // Basic representation, could be improved
    return "Function($argsn)"; 
  }

  @override
  String inspect(Idxs idxs) {
     return '[Function: Args ${spec.print(idxs)} Body ${body.print(idxs)}]';
  }

  @override
  bool equal(RyeObject other) {
     // Functions are generally compared by reference or definition location,
     // simple equality check might not be meaningful. For now, reference equality.
     return identical(this, other); 
  }

  @override
  int getKind() => type().index;
}
