// rye.dart - Main library file for the Dart Rye evaluator

// Export core types
export 'types.dart';

// Export environment and state
export 'env.dart';

// Export word indexing
export 'idxs.dart';

// Export evaluator functions
export 'evaldo.dart';

// Export builtin functions and registration
export 'builtins.dart';

// Export loader functions/classes (if needed by external users)
export 'loader.dart';

// Export individual builtin files if direct access to specific builtins is needed.
// Avoid exporting builtins_registration.dart as its functions are now just wrappers.
export 'builtins_collections.dart' show registerCollectionBuiltins; // Example: only export registration
export 'builtins_conditionals.dart' show registerConditionalBuiltins;
export 'builtins_flow.dart' show registerFlowBuiltins;
export 'builtins_iteration.dart' show registerIterationBuiltins;
export 'builtins_numbers.dart' show registerNumberBuiltins; // addBuiltin is also here
export 'builtins_printing.dart' show registerPrintingBuiltins; // printBuiltin is also here
// export 'builtins_registration.dart'; // DO NOT EXPORT - causes conflicts
export 'builtins_strings.dart' show registerStringBuiltins;
export 'builtins_types.dart' show registerTypeBuiltins;

// Note: builtins.dart exports registerBuiltins which calls all the above.
// Exporting builtins.dart is sufficient for standard use.
// The individual exports above are only needed if someone wants to call
// a specific registration function directly.

// --- Main Setup/Initialization (if any) ---
// This file could potentially contain top-level functions 
// to initialize a ProgramState with standard builtins, etc.
// For now, it mainly serves to export the separated components.

// Example initialization function (optional)
/*
import 'types.dart';
import 'env.dart';
import 'idxs.dart';
import 'builtins.dart';

ProgramState createInitialProgramState() {
  Idxs idxs = Idxs();
  // Create an empty initial series or load from somewhere
  TSeries initialSeries = TSeries([]); 
  ProgramState ps = ProgramState(initialSeries, idxs);
  registerBuiltins(ps); // Register all standard builtins
  return ps;
}
*/
