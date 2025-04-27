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
    bool inDoubleString = false;
    bool inSingleString = false; // For lit-words
    StringBuffer currentToken = StringBuffer();
    
    for (int i = 0; i < input.length; i++) {
      final char = input[i];
      
      if (char == '"' && !inSingleString) {
        // Handle double-quoted strings
        inDoubleString = !inDoubleString;
        currentToken.write(char);
      } else if (char == '\'' && !inDoubleString) {
         // Handle single-quoted lit-words
         // Start of lit-word: Add previous token if any, start new one with '
         if (!inSingleString) {
            if (currentToken.isNotEmpty) {
               tokens.add(currentToken.toString());
            }
            currentToken = StringBuffer("'"); // Start token with '
            inSingleString = true;
         } else {
             // End of lit-word: Add the closing quote and finalize token
             currentToken.write(char);
             tokens.add(currentToken.toString());
             currentToken = StringBuffer();
             inSingleString = false;
         }
      } else if (inDoubleString) {
         // Inside double-quoted string
         currentToken.write(char);
      } else if (inSingleString) {
         // Inside single-quoted lit-word (allow spaces and other chars)
         currentToken.write(char);
      } else if (char == '{' || char == '}' || char == '[' || char == ']') { // Add [ and ]
        // Handle braces, brackets, and comma as separate tokens
        if (currentToken.isNotEmpty) {
          tokens.add(currentToken.toString());
          currentToken = StringBuffer();
        }
        tokens.add(char);
      } else if (char == ',') { 
        // Handle comma
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
        // Parse a mode 2 block {}
        final (blockObjects, newIndex) = _parseBlock(tokens, i + 1, '}'); // Expect '}'
        final series = TSeries(blockObjects);
        objects.add(Block(series, 2)); // Mode 2 for {}
        i = newIndex;
      } else if (token == '[') {
         // Parse a mode 1 block []
        final (blockObjects, newIndex) = _parseBlock(tokens, i + 1, ']'); // Expect ']'
        final series = TSeries(blockObjects);
        objects.add(Block(series, 1)); // Mode 1 for []
        i = newIndex;
      } else if (token.endsWith(':')) {
        // Setword
        final wordStr = token.substring(0, token.length - 1);
        final wordIdx = idx.indexWord(wordStr);
        objects.add(Setword(wordIdx));
        i++;
      } else if (token == 'true') {
        // Boolean true
        objects.add(const Boolean(true));
        i++;
      } else if (token == 'false') {
        // Boolean false
        objects.add(const Boolean(false));
        i++;
      } else if (int.tryParse(token) != null) {
        // Integer
        objects.add(Integer(int.parse(token)));
        i++;
      } else if (double.tryParse(token) != null) {
        // Decimal
        objects.add(Decimal(double.parse(token)));
        i++;
      } else if (token.startsWith('"') && token.endsWith('"')) {
        // String
        final str = token.substring(1, token.length - 1);
        // TODO: Handle string escaping
        objects.add(RyeString(str));
        i++;
      } else if (token.startsWith('/') && token.contains('/')) {
         // Basic CPath parsing (assuming /word1/word2 format)
         // TODO: This needs a more robust tokenizer/parser
         List<Word> path = token.split('/')
                              .where((s) => s.isNotEmpty)
                              .map((s) => Word(idx.indexWord(s)))
                              .toList();
         if (path.isNotEmpty) {
            objects.add(CPath(path));
         } else {
            // Handle potential error or default case
            objects.add(Word(idx.indexWord(token))); // Fallback to Word
         }
         i++;
      } else if (token == ',') {
        // Comma
        objects.add(const Comma());
        i++;
      } else if (token.startsWith('\'') && token.endsWith('\'') && token.length >= 2) {
        // LitWord
        // Handle potential spaces inside lit-word if tokenizer allows it (current doesn't robustly)
        final wordStr = token.substring(1, token.length - 1); 
        final wordIdx = idx.indexWord(wordStr);
        objects.add(LitWord(wordIdx));
        i++;
      } else if (token.startsWith('%') || token.contains('://')) {
        // URI (basic check)
        // TODO: Improve robustness, handle potential quotes if tokenizer changes
        try {
          objects.add(Uri.fromString(idx, token));
        } catch (e) {
           // Fallback or error handling if Uri.fromString fails
           objects.add(Word(idx.indexWord(token))); // Fallback to Word
        }
        i++;
      } else {
        // Word, OpWord, PipeWord, LSetword, LModword
        // TODO: Distinguishing LSetword/LModword requires lookbehind or different parsing strategy
        if (token.startsWith('|') && token.length > 1) {
          final wordStr = token.substring(1);
          final wordIdx = idx.indexWord(wordStr);
          objects.add(PipeWord(wordIdx));
        } else if (token.endsWith('::')) { // Placeholder for LModword
           final wordStr = token.substring(0, token.length - 2);
           final wordIdx = idx.indexWord(wordStr);
           objects.add(LModword(wordIdx));
        } else if (token.endsWith(':')) { // Placeholder for LSetword (overlaps with Setword)
           // This currently conflicts with Setword parsing logic placed earlier.
           // A better parser is needed. For now, LSetword won't be parsed correctly here.
           final wordStr = token.substring(0, token.length - 1);
           final wordIdx = idx.indexWord(wordStr);
           objects.add(Word(wordIdx)); // Fallback
        } else {
          final wordIdx = idx.indexWord(token);
          if (_isOpWordToken(token)) {
            objects.add(OpWord(wordIdx));
          } else {
            objects.add(Word(wordIdx));
          }
        }
        i++;
      }
    }
    
    return objects;
  }
  
  /// Parses the content inside a block from tokens until a closing delimiter is found.
  (List<RyeObject>, int) _parseBlock(List<String> tokens, int startIndex, String closingDelimiter) {
    final blockObjects = <RyeObject>[];
    int i = startIndex;
    
    while (i < tokens.length) {
      final token = tokens[i];
      
      if (token == closingDelimiter) {
        // End of block
        return (blockObjects, i + 1);
      } else if (token == '{') {
        // Nested block {}
        final (nestedBlockObjects, newIndex) = _parseBlock(tokens, i + 1, '}');
        final series = TSeries(nestedBlockObjects);
        blockObjects.add(Block(series, 2)); // Mode 2 for {}
        i = newIndex;
      } else if (token == '[') {
        // Nested block []
        final (nestedBlockObjects, newIndex) = _parseBlock(tokens, i + 1, ']');
        final series = TSeries(nestedBlockObjects);
        blockObjects.add(Block(series, 1)); // Mode 1 for []
        i = newIndex;
      } else if (token.endsWith(':')) {
        // Setword
        final wordStr = token.substring(0, token.length - 1);
        final wordIdx = idx.indexWord(wordStr);
        blockObjects.add(Setword(wordIdx));
        i++;
      } else if (token == 'true') {
        // Boolean true
        blockObjects.add(const Boolean(true));
        i++;
      } else if (token == 'false') {
        // Boolean false
        blockObjects.add(const Boolean(false));
        i++;
      } else if (int.tryParse(token) != null) {
        // Integer
        blockObjects.add(Integer(int.parse(token)));
        i++;
      } else if (double.tryParse(token) != null) {
        // Decimal
        blockObjects.add(Decimal(double.parse(token)));
        i++;
      } else if (token.startsWith('"') && token.endsWith('"')) {
        // String
        final str = token.substring(1, token.length - 1);
        // TODO: Handle string escaping
        blockObjects.add(RyeString(str));
        i++;
      } else if (token.startsWith('/') && token.contains('/')) {
         // Basic CPath parsing
         List<Word> path = token.split('/')
                              .where((s) => s.isNotEmpty)
                              .map((s) => Word(idx.indexWord(s)))
                              .toList();
         if (path.isNotEmpty) {
            blockObjects.add(CPath(path));
         } else {
            blockObjects.add(Word(idx.indexWord(token))); // Fallback
         }
         i++;
      } else if (token == ',') {
        // Comma
        blockObjects.add(const Comma());
        i++;
      } else if (token.startsWith('\'') && token.endsWith('\'') && token.length >= 2) {
        // LitWord
        final wordStr = token.substring(1, token.length - 1);
        final wordIdx = idx.indexWord(wordStr);
        blockObjects.add(LitWord(wordIdx));
        i++;
      } else if (token.startsWith('%') || token.contains('://')) {
        // URI (basic check)
         try {
          blockObjects.add(Uri.fromString(idx, token));
        } catch (e) {
           blockObjects.add(Word(idx.indexWord(token))); // Fallback to Word
        }
        i++;
      } else {
        // Word, OpWord, PipeWord, LSetword, LModword
        // TODO: Distinguishing LSetword/LModword requires lookbehind or different parsing strategy
         if (token.startsWith('|') && token.length > 1) {
          final wordStr = token.substring(1);
          final wordIdx = idx.indexWord(wordStr);
          blockObjects.add(PipeWord(wordIdx));
        } else if (token.endsWith('::')) { // Placeholder for LModword
           final wordStr = token.substring(0, token.length - 2);
           final wordIdx = idx.indexWord(wordStr);
           blockObjects.add(LModword(wordIdx));
        } else if (token.endsWith(':')) { // Placeholder for LSetword (overlaps with Setword)
           // This currently conflicts with Setword parsing logic placed earlier.
           final wordStr = token.substring(0, token.length - 1);
           final wordIdx = idx.indexWord(wordStr);
           blockObjects.add(Word(wordIdx)); // Fallback
        } else {
          final wordIdx = idx.indexWord(token);
          if (_isOpWordToken(token)) {
            blockObjects.add(OpWord(wordIdx));
          } else {
            blockObjects.add(Word(wordIdx));
          }
        }
        i++;
      }
    }
    
    // If we reach here, the block was not closed
    throw Exception("Unclosed block");
  }
}

// Helper to check if a token string represents an opword
bool _isOpWordToken(String token) {
  // Starts with '.' or is a specific operator/keyword
  const opChars = ['+', '-', '*', '/', '>', '<', '=', '==', '!=', '>=', '<=']; // Add more as needed
  return token.startsWith('.') || opChars.contains(token) ||
         ['if', 'for', 'loop', 'either', 'switch', 'when', 'while', 'map', 'filter', 'forever', 'forever-with'].contains(token); // Keep keywords for now
}

/// OpWord implementation (distinct from regular Word)
class OpWord implements RyeObject {
  final int index;

  const OpWord(this.index);

  @override
  RyeType type() => RyeType.wordType; // Still treated as a word type for now

  @override
  String print(Idxs idxs) {
    return idxs.getWord(index); // Print like a normal word
  }

  @override
  String inspect(Idxs idxs) {
    return '[OpWord: ${print(idxs)}]';
  }

  @override
  bool equal(RyeObject other) {
    if (other is! OpWord) return false;
    return index == other.index;
  }

  @override
  int getKind() => type().index;
}

/// PipeWord implementation
class PipeWord implements RyeObject {
  final int index;
  final bool force; // Corresponds to * in |word*

  const PipeWord(this.index, [this.force = false]);

  @override
  RyeType type() => RyeType.wordType; // Still treated as a word type for now

  @override
  String print(Idxs idxs) {
    return '|${idxs.getWord(index)}'; 
  }

  @override
  String inspect(Idxs idxs) {
    return '[PipeWord: ${print(idxs)}]';
  }

  @override
  bool equal(RyeObject other) {
    if (other is! PipeWord) return false;
    return index == other.index && force == other.force;
  }

  @override
  int getKind() => type().index;
}

/// LSetword implementation (Setword appearing after an expression)
class LSetword implements RyeObject {
  final int index;

  const LSetword(this.index);

  @override
  RyeType type() => RyeType.wordType; // Still treated as a word type for now

  @override
  String print(Idxs idxs) {
    return '${idxs.getWord(index)}:'; 
  }

  @override
  String inspect(Idxs idxs) {
    return '[LSetword: ${print(idxs)}]';
  }

  @override
  bool equal(RyeObject other) {
    if (other is! LSetword) return false;
    return index == other.index;
  }

  @override
  int getKind() => type().index;
}

/// LModword implementation (Modword appearing after an expression)
class LModword implements RyeObject {
  final int index;

  const LModword(this.index);

  @override
  RyeType type() => RyeType.wordType; // Still treated as a word type for now

  @override
  String print(Idxs idxs) {
    return '${idxs.getWord(index)}::'; 
  }

  @override
  String inspect(Idxs idxs) {
    return '[LModword: ${print(idxs)}]';
  }

  @override
  bool equal(RyeObject other) {
    if (other is! LModword) return false;
    return index == other.index;
  }

  @override
  int getKind() => type().index;
}

/// CPath implementation (Context Path)
class CPath implements RyeObject {
  final List<Word> path;
  final int mode; // 0: normal, 1: opword-like, 2: pipeword-like (based on Go)

  const CPath(this.path, [this.mode = 0]);

  @override
  RyeType type() => RyeType.wordType; // Still treated as a word type for now

  @override
  String print(Idxs idxs) {
    return path.map((w) => '/${w.print(idxs)}').join('');
  }

  @override
  String inspect(Idxs idxs) {
    return '[CPath: ${print(idxs)}]';
  }

  @override
  bool equal(RyeObject other) {
    if (other is! CPath) return false;
    if (path.length != other.path.length || mode != other.mode) return false;
    for (int i = 0; i < path.length; i++) {
      if (!path[i].equal(other.path[i])) return false;
    }
    return true;
  }

  @override
  int getKind() => type().index;
}

/// LitWord implementation (Literal Word)
class LitWord implements RyeObject {
  final int index;

  const LitWord(this.index);

  @override
  RyeType type() => RyeType.wordType; // Still treated as a word type for now

  @override
  String print(Idxs idxs) {
    // Represent lit-words with a leading quote in output for clarity
    return "'${idxs.getWord(index)}"; 
  }

  @override
  String inspect(Idxs idxs) {
    return "[LitWord: ${idxs.getWord(index)}]";
  }

  @override
  bool equal(RyeObject other) {
    if (other is! LitWord) return false;
    return index == other.index;
  }

  @override
  int getKind() => type().index;
}


/// Comma implementation (Expression separator)
class Comma implements RyeObject {
  const Comma();

  @override
  RyeType type() => RyeType.wordType; // Treat as word-like for now

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


/// Setword implementation (Setword appearing before an expression)
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
