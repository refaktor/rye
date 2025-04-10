// loader.dart - A simple parser for the Rye language

import 'dart:io';
import 'rye.dart';
import 'types.dart';

/// RyeLoader is responsible for parsing Rye code into a Block object
class RyeLoader {
  final Idxs idx;

  /// Creates a new RyeLoader with the given index
  RyeLoader(this.idx);

  /// Loads a string of Rye code and returns a Block object
  (RyeObject, bool) loadString(String input) {
    // Remove shebang line if present
    input = _removeBangLine(input);

    // Wrap input in a block if it doesn't start with one
    input = _wrapInBlock(input);

    try {
      // Parse the input using a simple tokenizer
      final tokens = _tokenize(input);
      final objects = _parseTokens(tokens);
      
      // Create a block from the parsed objects
      final series = TSeries(objects);
      return (Block(series), true);
    } catch (e) {
      return (Error("Parse error: $e"), false);
    }
  }

  /// Removes the shebang line from the input if present
  String _removeBangLine(String content) {
    if (content.startsWith('#!')) {
      final newlineIndex = content.indexOf('\n');
      if (newlineIndex != -1) {
        return content.substring(newlineIndex + 1);
      }
    }
    return content;
  }

  /// Wraps the input in a block if it doesn't start with one
  String _wrapInBlock(String input) {
    final trimmed = input.trim();
    if (trimmed.isEmpty || !trimmed.startsWith('{')) {
      return '{ $input }';
    }
    return input;
  }

  /// Tokenizes the input string into a list of tokens
  List<String> _tokenize(String input) {
    // Split the input into tokens
    final tokens = <String>[];
    bool inString = false;
    StringBuffer currentToken = StringBuffer();
    
    for (int i = 0; i < input.length; i++) {
      final char = input[i];
      
      if (char == '"') {
        // Handle strings
        inString = !inString;
        currentToken.write(char);
      } else if (inString) {
        // Inside a string, add all characters
        currentToken.write(char);
      } else if (char == '{' || char == '}') {
        // Handle braces as separate tokens
        if (currentToken.isNotEmpty) {
          tokens.add(currentToken.toString());
          currentToken = StringBuffer();
        }
        tokens.add(char);
      } else if (char.trim().isEmpty) {
        // Handle whitespace
        if (currentToken.isNotEmpty) {
          tokens.add(currentToken.toString());
          currentToken = StringBuffer();
        }
      } else {
        // Add character to current token
        currentToken.write(char);
      }
    }
    
    if (currentToken.isNotEmpty) {
      tokens.add(currentToken.toString());
    }
    
    return tokens;
  }

  /// Parses the tokens into a list of RyeObjects
  List<RyeObject> _parseTokens(List<String> tokens) {
    final objects = <RyeObject>[];
    int i = 0;
    
    while (i < tokens.length) {
      final token = tokens[i];
      
      if (token == '{') {
        // Parse a block
        final (blockObjects, newIndex) = _parseBlock(tokens, i + 1);
        final series = TSeries(blockObjects);
        objects.add(Block(series));
        i = newIndex;
      } else if (token.endsWith(':')) {
        // Setword
        final wordStr = token.substring(0, token.length - 1);
        final wordIdx = idx.indexWord(wordStr);
        objects.add(Setword(wordIdx));
        i++;
      } else if (int.tryParse(token) != null) {
        // Integer
        objects.add(Integer(int.parse(token)));
        i++;
      } else if (token.startsWith('"') && token.endsWith('"')) {
        // String
        final str = token.substring(1, token.length - 1);
        // TODO: Handle string escaping
        objects.add(RyeString(str));
        i++;
      } else {
        // Word
        final wordIdx = idx.indexWord(token);
        objects.add(Word(wordIdx));
        i++;
      }
    }
    
    return objects;
  }
  
  /// Parses a block from tokens
  (List<RyeObject>, int) _parseBlock(List<String> tokens, int startIndex) {
    final blockObjects = <RyeObject>[];
    int i = startIndex;
    
    while (i < tokens.length) {
      final token = tokens[i];
      
      if (token == '}') {
        // End of block
        return (blockObjects, i + 1);
      } else if (token == '{') {
        // Nested block
        final (nestedBlockObjects, newIndex) = _parseBlock(tokens, i + 1);
        final series = TSeries(nestedBlockObjects);
        blockObjects.add(Block(series));
        i = newIndex;
      } else if (token.endsWith(':')) {
        // Setword
        final wordStr = token.substring(0, token.length - 1);
        final wordIdx = idx.indexWord(wordStr);
        blockObjects.add(Setword(wordIdx));
        i++;
      } else if (int.tryParse(token) != null) {
        // Integer
        blockObjects.add(Integer(int.parse(token)));
        i++;
      } else if (token.startsWith('"') && token.endsWith('"')) {
        // String
        final str = token.substring(1, token.length - 1);
        // TODO: Handle string escaping
        blockObjects.add(RyeString(str));
        i++;
      } else {
        // Word
        final wordIdx = idx.indexWord(token);
        blockObjects.add(Word(wordIdx));
        i++;
      }
    }
    
    // If we reach here, the block was not closed
    throw Exception("Unclosed block");
  }
}

/// Setword implementation
class Setword implements RyeObject {
  final int index;

  const Setword(this.index);

  @override
  RyeType type() => RyeType.wordType; // Using wordType for now

  @override
  String print(Idxs idxs) {
    return '${idxs.getWord(index)}:';
  }

  @override
  String inspect(Idxs idxs) {
    return '[Setword: ${print(idxs)}]';
  }

  @override
  bool equal(RyeObject other) {
    if (other.type() != type()) return false;
    return index == (other as Setword).index;
  }

  @override
  int getKind() => type().index;
}
