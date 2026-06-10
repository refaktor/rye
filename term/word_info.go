package term

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/refaktor/rye/env"
	"github.com/refaktor/rye/loader"
)

// findWordInfo searches for documentation about a word in RYE_HOME/info/*.info.rye files
func findWordInfo(ps *env.ProgramState, word string) string {
	if word == "" {
		return ""
	}

	// Strip any prefix operators
	cleanWord := word
	if strings.HasPrefix(word, ".") || strings.HasPrefix(word, "|") {
		cleanWord = word[1:]
	}

	if cleanWord == "" {
		return ""
	}

	// Find RYE_HOME or use current directory structure
	ryeHome := os.Getenv("RYE_HOME")
	if ryeHome == "" {
		// Try to find info directory relative to current working directory
		if wd, err := os.Getwd(); err == nil {
			// Look for info directory in current directory or parent directories
			testDirs := []string{
				filepath.Join(wd, "info"),
				filepath.Join(wd, "tests"), // Since the example files are in tests/ directory
			}
			for _, dir := range testDirs {
				if _, err := os.Stat(dir); err == nil {
					ryeHome = filepath.Dir(dir)
					break
				}
			}
		}
	}

	if ryeHome == "" {
		return ""
	}

	// Look for info files
	infoDirs := []string{
		filepath.Join(ryeHome, "info"),
		filepath.Join(ryeHome, "tests"), // Fallback for the current structure
	}

	for _, infoDir := range infoDirs {
		if info := searchInfoInDirectory(ps, infoDir, cleanWord); info != "" {
			return info
		}
	}

	return ""
}

// searchInfoInDirectory searches for word info in *.info.rye files within a directory
func searchInfoInDirectory(ps *env.ProgramState, infoDir, word string) string {
	if _, err := os.Stat(infoDir); os.IsNotExist(err) {
		return ""
	}

	files, err := filepath.Glob(filepath.Join(infoDir, "*.info.rye"))
	if err != nil {
		return ""
	}

	for _, file := range files {
		if info := searchWordInFile(ps, file, word); info != "" {
			return info
		}
	}

	return ""
}

// searchWordInFile searches for word documentation in a specific info.rye file
func searchWordInFile(ps *env.ProgramState, filename, word string) string {
	content, err := os.ReadFile(filename)
	if err != nil {
		return ""
	}

	// Parse the file using Rye's no_peg parser
	block, _ := loader.LoadStringNoPEG(string(content), false)
	blockObj, ok := block.(env.Block)
	if !ok {
		return ""
	}

	// Search for the word in the parsed structure
	return searchWordInBlock(ps, blockObj, word)
}

// searchWordInBlock recursively searches for a word's documentation in a parsed block
func searchWordInBlock(ps *env.ProgramState, block env.Block, targetWord string) string {
	series := block.Series
	items := series.GetAll()

	for i := 0; i < len(items); i++ {
		item := items[i]

		// Look for 'group' followed by the word name
		if word, ok := item.(env.Word); ok {
			if ps.Idx.GetWord(word.Index) == "group" && i+1 < len(items) {
				// Next item should be the word name
				if nextItem, ok := items[i+1].(env.String); ok {
					wordName := nextItem.Value
					if wordName == targetWord {
						// Found the target word! Now collect its documentation
						return extractWordDocumentation(ps, items, i, targetWord)
					}
				}
			}
		}

		// If this is a block, search recursively
		if subBlock, ok := item.(env.Block); ok {
			if result := searchWordInBlock(ps, subBlock, targetWord); result != "" {
				return result
			}
		}
	}

	return ""
}

// extractWordDocumentation extracts the documentation for a word starting at the given index
func extractWordDocumentation(ps *env.ProgramState, items []env.Object, startIdx int, wordName string) string {
	var result strings.Builder
	
	result.WriteString(fmt.Sprintf("\n\033[1;36m%s\033[0m\n", wordName))
	result.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")

	// Look for the description (string after the word name)
	if startIdx+2 < len(items) {
		if desc, ok := items[startIdx+2].(env.String); ok {
			result.WriteString("\033[33mDescription:\033[0m ")
			result.WriteString(desc.Value)
			result.WriteString("\n\n")
		}
	}

	// Look for the specification block (next block after description)
	if startIdx+3 < len(items) {
		if specBlock, ok := items[startIdx+3].(env.Block); ok {
			extractSpecification(ps, &result, specBlock)
		}
	}

	// Look for test examples (next block after specification)
	if startIdx+4 < len(items) {
		if testBlock, ok := items[startIdx+4].(env.Block); ok {
			extractExamples(ps, &result, testBlock)
		}
	}

	return result.String()
}

// extractSpecification extracts argument and return information from the specification block
func extractSpecification(ps *env.ProgramState, result *strings.Builder, block env.Block) {
	items := block.Series.GetAll()
	
	for i := 0; i < len(items); i++ {
		item := items[i]
		
		if word, ok := item.(env.Word); ok {
			wordName := ps.Idx.GetWord(word.Index)
			
			switch wordName {
			case "argsn":
				if i+1 < len(items) {
					if argCount, ok := items[i+1].(env.Integer); ok {
						result.WriteString(fmt.Sprintf("\033[32mArguments:\033[0m %d\n", argCount.Value))
					}
				}
			case "arg":
				if i+1 < len(items) {
					if argDesc, ok := items[i+1].(env.String); ok {
						result.WriteString(fmt.Sprintf("  • %s\n", argDesc.Value))
					}
				}
			case "returns":
				if i+1 < len(items) {
					if returnDesc, ok := items[i+1].(env.String); ok {
						result.WriteString(fmt.Sprintf("\033[35mReturns:\033[0m %s\n", returnDesc.Value))
					}
				}
			case "pure":
				result.WriteString("\033[36mPure function\033[0m (no side effects)\n")
			case "argtypes":
				if i+1 < len(items) {
					if typeBlock, ok := items[i+1].(env.Block); ok {
						extractArgTypes(ps, result, typeBlock)
					}
				}
			}
		}
	}
	result.WriteString("\n")
}

// extractArgTypes extracts argument type information
func extractArgTypes(ps *env.ProgramState, result *strings.Builder, block env.Block) {
	items := block.Series.GetAll()
	argNum := 1
	
	for _, item := range items {
		if argBlock, ok := item.(env.Block); ok {
			if len(argBlock.Series.GetAll()) > 1 {
				if num, ok := argBlock.Series.GetAll()[0].(env.Integer); ok {
					argNum = int(num.Value)
				}
				if typeList, ok := argBlock.Series.GetAll()[1].(env.Block); ok {
					types := make([]string, 0)
					for _, typeItem := range typeList.Series.GetAll() {
						if typeWord, ok := typeItem.(env.Word); ok {
							types = append(types, ps.Idx.GetWord(typeWord.Index))
						}
					}
					if len(types) > 0 {
						result.WriteString(fmt.Sprintf("  Arg %d types: %s\n", argNum, strings.Join(types, ", ")))
					}
				}
			}
		}
	}
}

// extractExamples extracts test examples from the test block
func extractExamples(ps *env.ProgramState, result *strings.Builder, block env.Block) {
	items := block.Series.GetAll()
	
	if len(items) > 0 {
		result.WriteString("\033[33mExamples:\033[0m\n")
		
		for _, item := range items {
			if testBlock, ok := item.(env.Block); ok {
				testItems := testBlock.Series.GetAll()
				
				for _, testItem := range testItems {
					if word, ok := testItem.(env.Word); ok {
						testType := ps.Idx.GetWord(word.Index)
						if testType == "equal" {
							// Handle 'equal' test
							if len(testItems) >= 3 {
								if codeBlock, ok := testItems[1].(env.Block); ok {
									expectedValue := testItems[2]
									codeStr := formatBlockForDisplay(ps, codeBlock)
									expectedStr := formatValueForDisplay(ps, expectedValue)
									result.WriteString(fmt.Sprintf("  %s  ; => %s\n", codeStr, expectedStr))
								}
							}
						} else if testType == "error" {
							// Handle 'error' test
							if len(testItems) >= 2 {
								if codeBlock, ok := testItems[1].(env.Block); ok {
									codeStr := formatBlockForDisplay(ps, codeBlock)
									result.WriteString(fmt.Sprintf("  %s  ; => ERROR\n", codeStr))
								}
							}
						}
					}
				}
			}
		}
		result.WriteString("\n")
	}
}

// formatBlockForDisplay formats a block for display in examples
func formatBlockForDisplay(ps *env.ProgramState, block env.Block) string {
	var result strings.Builder
	items := block.Series.GetAll()
	
	for i, item := range items {
		if i > 0 {
			result.WriteString(" ")
		}
		result.WriteString(formatValueForDisplay(ps, item))
	}
	
	return result.String()
}

// formatValueForDisplay formats a value for display
func formatValueForDisplay(ps *env.ProgramState, value env.Object) string {
	switch v := value.(type) {
	case env.String:
		return fmt.Sprintf(`"%s"`, v.Value)
	case env.Integer:
		return fmt.Sprintf("%d", v.Value)
	case env.Decimal:
		return fmt.Sprintf("%.1f", v.Value)
	case env.Word:
		return ps.Idx.GetWord(v.Index)
	case env.Block:
		return "{ " + formatBlockForDisplay(ps, v) + " }"
	default:
		return fmt.Sprintf("%v", v)
	}
}