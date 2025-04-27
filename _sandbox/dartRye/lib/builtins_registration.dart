// builtins_registration.dart - Registration methods for Rye builtins

import 'rye.dart';
import 'builtins_collections.dart';
import 'builtins_strings.dart';
import 'builtins_flow.dart' as flow;
import 'builtins_numbers.dart';
import 'builtins_iteration.dart';
import 'builtins_conditionals.dart';
import 'builtins_printing.dart' as printing; // Import with prefix
import 'builtins_types.dart';

// Register collection builtins
void registerCollectionBuiltins(ProgramState ps) {
  // This function should be implemented in builtins_collections.dart
  // For now, we'll provide a stub implementation
  // In a real implementation, this would register all collection-related builtins
  int listIdx = ps.idx.indexWord("list");
  int dictIdx = ps.idx.indexWord("dict");
  
  // Register any collection-related builtins here
}

// Register string builtins
void registerStringBuiltins(ProgramState ps) {
  // This function should be implemented in builtins_strings.dart
  // For now, we'll provide a stub implementation
  // In a real implementation, this would register all string-related builtins
  int strIdx = ps.idx.indexWord("string");
  
  // Register any string-related builtins here
}

// Register flow builtins - This is now just a wrapper that calls the implementation in builtins_flow.dart
void registerFlowBuiltins(ProgramState ps) {
  // Call the actual implementation from builtins_flow.dart
  flow.registerFlowBuiltins(ps);
}

// Register number builtins
void registerNumberBuiltins(ProgramState ps) {
  // This function should be implemented in builtins_numbers.dart
  // For now, we'll provide a stub implementation
  // In a real implementation, this would register all number-related builtins
  int addIdx = ps.idx.indexWord("_+");
  int subIdx = ps.idx.indexWord("_-");
  int mulIdx = ps.idx.indexWord("_*");
  int divIdx = ps.idx.indexWord("_/");
  
  // Register any number-related builtins here
  // The _+ builtin is already registered in rye.dart
}

// Register iteration builtins
void registerIterationBuiltins(ProgramState ps) {
  // This function should be implemented in builtins_iteration.dart
  // For now, we'll provide a stub implementation
  // In a real implementation, this would register all iteration-related builtins
  int forIdx = ps.idx.indexWord("for");
  int mapIdx = ps.idx.indexWord("map");
  int filterIdx = ps.idx.indexWord("filter");
  
  // Register any iteration-related builtins here
}

// Register conditional builtins
void registerConditionalBuiltins(ProgramState ps) {
  // This function should be implemented in builtins_conditionals.dart
  // For now, we'll provide a stub implementation
  // In a real implementation, this would register all conditional builtins
  int ifIdx = ps.idx.indexWord("if");
  int eitherIdx = ps.idx.indexWord("either");
  int switchIdx = ps.idx.indexWord("switch");
  
  // Register any conditional builtins here
}

// Register printing builtins - This now calls the implementation in builtins_printing.dart
void registerPrintingBuiltins(ProgramState ps) {
  // Call the actual implementation from builtins_printing.dart
  printing.registerPrintingBuiltins(ps);
}

// Register type builtins
void registerTypeBuiltins(ProgramState ps) {
  // This function should be implemented in builtins_types.dart
  // For now, we'll provide a stub implementation
  // In a real implementation, this would register all type-related builtins
  int typeIdx = ps.idx.indexWord("type?");
  
  // Register any type-related builtins here
}
