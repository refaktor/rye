// builtins_registration.dart - Calls registration methods for Rye builtins

import 'env.dart' show ProgramState; // Import ProgramState
import 'builtins_collections.dart' show registerCollectionBuiltins; // Import only the registration function
import 'builtins_strings.dart' show registerStringBuiltins; // Import only the registration function
import 'builtins_flow.dart' show registerFlowBuiltins; // Import only the registration function
import 'builtins_numbers.dart' show registerNumberBuiltins; // Import only the registration function
import 'builtins_iteration.dart' show registerIterationBuiltins; // Import only the registration function
import 'builtins_conditionals.dart' show registerConditionalBuiltins; // Import only the registration function
import 'builtins_printing.dart' show registerPrintingBuiltins; // Import only the registration function
import 'builtins_types.dart' show registerTypeBuiltins; // Import only the registration function

// This file is now primarily for potentially grouping registration calls,
// but the main registration happens in builtins.dart.
// The individual register* functions are defined in their respective files.

// Example of a function that could call all registrations if needed elsewhere:
// void registerAllBuiltinModules(ProgramState ps) {
//   registerCollectionBuiltins(ps);
//   registerStringBuiltins(ps);
//   registerFlowBuiltins(ps);
//   registerNumberBuiltins(ps);
//   registerIterationBuiltins(ps);
//   registerConditionalBuiltins(ps);
//   registerPrintingBuiltins(ps);
//   registerTypeBuiltins(ps);
// }
