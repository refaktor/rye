// idxs.dart - Word indexing for Dart Rye evaluator

// Idxs is a bidirectional mapping between words (strings) and their indices.
class Idxs {
  List<String> words1 = [""];
  Map<String, int> words2 = {};

  Idxs() {
    // Register some basic words
    indexWord("_+");
    indexWord("integer");
    indexWord("word");
    indexWord("block");
    indexWord("builtin");
    indexWord("error");
    indexWord("void");
    // Add other common words if needed during refactoring
  }

  int indexWord(String w) {
    int? idx = words2[w];
    if (idx != null) {
      return idx;
    } else {
      words1.add(w);
      words2[w] = words1.length - 1;
      return words1.length - 1;
    }
  }

  (int, bool) getIndex(String w) {
    int? idx = words2[w];
    if (idx != null) {
      return (idx, true);
    }
    return (0, false);
  }

  String getWord(int i) {
    if (i < 0 || i >= words1.length) { // Added bounds check
      return "invalid-index!"; 
    }
    return words1[i];
  }

  int getWordCount() {
    return words1.length;
  }
}
