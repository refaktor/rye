# Dart Rye

A Dart implementation of the Rye language evaluator. This is a simplified version that only handles integers, blocks, and builtins.

## Overview

This project is a port of the Rye language evaluator from Go to Dart. It implements a subset of the Rye language, focusing on:

- Integer values
- Block values
- Word lookup
- Builtin functions (currently only `_+` is implemented)

## Structure

The project is organized into a single file for simplicity:

- `rye.dart`: Contains all the core components of the Rye evaluator
- `main.dart`: Exports the rye.dart module
- `bin/dart_rye.dart`: Entry point that demonstrates the usage of the evaluator

## Running the Example

To run the example program (which evaluates `3 _+ 4`):

```bash
cd _sandbox/dartRye
dart run
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

## Future Improvements

- Add more value types (strings, decimals, etc.)
- Implement more builtins
- Add a parser to convert Rye code from text to a series of objects
- Implement more advanced features like contexts and functions
- Split the code into multiple files for better organization
