# Dart Rye

A Dart implementation of the Rye language evaluator. This is a simplified version that handles integers, blocks, builtins, and includes a parser.

## Overview

This project is a port of the Rye language evaluator from Go to Dart. It implements a subset of the Rye language, focusing on:

- Integer values
- Block values
- Word lookup
- Builtin functions (currently only `_+` is implemented)
- Simple parser for Rye code

## Structure

The project is organized into several files:

- `rye.dart`: Contains all the core components of the Rye evaluator
- `loader.dart`: Implements the parser for Rye code
- `main.dart`: Exports the modules
- `bin/dart_rye.dart`: Entry point that demonstrates the usage of the evaluator
- `bin/loader_example.dart`: Example of using the loader

## Running the Examples

To run the basic example (which evaluates `3 _+ 4`):

```bash
cd _sandbox/dartRye
dart run
```

To run the loader example (which parses and evaluates Rye code):

```bash
cd _sandbox/dartRye
dart run bin/loader_example.dart
```

## Parser Implementation

The parser is implemented using a simple tokenizer and parser approach. It splits the input into tokens and then converts each token into the appropriate Rye object.

The parser supports:
- Integers
- Words
- Setwords
- Blocks
- Basic string literals

Here's an example of how the parser works:

```dart
// Sample Rye code
final sampleCode = '''{ _+ 1 10 }''';

// Load the code
final (result, success) = loader.loadString(sampleCode);

if (success) {
  // Create a program state
  final block = result as Block;
  final series = block.series;
  final ps = ProgramState(series, idx);
  
  // Register builtins
  registerBuiltins(ps);
  
  // Evaluate the program
  rye00_evalBlockInj(ps, null, false);
  
  // Display the result
  stdout.writeln("Result: ${ps.res!.print(idx)}");
}
```

## Adding More Builtins

To add more builtins, you can follow the pattern used for the `_+` builtin in `rye.dart`. For example, to add a subtraction builtin:

1. Implement the builtin function:
```dart
RyeObject subtractBuiltin(ProgramState ps, RyeObject? arg0, RyeObject? arg1, RyeObject? arg2, RyeObject? arg3, RyeObject? arg4) {
  if (arg0 is Integer && arg1 is Integer) {
    return Integer(arg0.value - arg1.value);
  }
  
  ps.failureFlag = true;
  return Error("Arguments to _- must be integers");
}
```

2. Register the builtin in the `registerBuiltins` function:
```dart
void registerBuiltins(ProgramState ps) {
  // Existing code...
  
  int minusIdx = ps.idx.indexWord("_-");
  Builtin minusBuiltin = Builtin(subtractBuiltin, 2, false, true, "Subtracts two integers");
  ps.ctx.set(minusIdx, minusBuiltin);
}
```

## Implementation Details

The implementation follows the same structure as the Go code, with some adaptations for Dart:

1. **RyeObject Interface**: Defines the base interface for all Rye values
2. **Value Types**: Implements Integer, Block, Word, and other value types
3. **Evaluation Logic**: Implements the simplified evaluator that handles integers, blocks, and builtins
4. **Builtin Functions**: Implements the "_+" builtin function as an example
5. **Parser**: Uses a simple tokenizer and parser to convert Rye code into a Block object

## Future Improvements

- Add more value types (decimals, etc.)
- Implement more builtins
- Enhance the parser to support more Rye syntax (nested blocks, etc.)
- Implement more advanced features like contexts and functions
- Add a more sophisticated parser using a library like PetitParser
