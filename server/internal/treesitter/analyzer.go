package treesitter

import (
	"strings"

	ts "github.com/tree-sitter/go-tree-sitter"
)

// FileSignals contains the imports and definitions extracted from a source file.
type FileSignals struct {
	Imports     []string // imported module/package paths
	Definitions []string // exported function/type/class names
}

// Analyze parses a source file and extracts imports and definitions.
// Returns nil for unsupported languages (not an error).
func Analyze(filename string, content []byte) *FileSignals {
	ld := LangForFile(filename)
	if ld == nil {
		return nil
	}

	parser := ts.NewParser()
	defer parser.Close()
	if err := parser.SetLanguage(ld.language); err != nil {
		return nil
	}

	tree := parser.Parse(content, nil)
	if tree == nil {
		return nil
	}
	defer tree.Close()

	root := tree.RootNode()
	signals := &FileSignals{}

	// Extract imports.
	signals.Imports = runQuery(ld.language, ld.importQ, root, content, "import")

	// Extract definitions.
	signals.Definitions = runQuery(ld.language, ld.defQ, root, content, "def")

	return signals
}

// runQuery runs a tree-sitter query and returns the text of all captures
// matching the given capture name.
func runQuery(lang *ts.Language, queryStr string, root *ts.Node, source []byte, captureName string) []string {
	if queryStr == "" {
		return nil
	}

	query, qErr := ts.NewQuery(lang, queryStr)
	if qErr != nil {
		return nil
	}
	defer query.Close()

	// Find the capture index for the target name.
	captureIdx, ok := query.CaptureIndexForName(captureName)
	if !ok {
		return nil
	}

	cursor := ts.NewQueryCursor()
	defer cursor.Close()
	matches := cursor.Matches(query, root, source)

	var results []string
	seen := map[string]bool{}

	for match := matches.Next(); match != nil; match = matches.Next() {
		for _, capture := range match.Captures {
			if capture.Index != uint32(captureIdx) {
				continue
			}
			text := capture.Node.Utf8Text(source)
			text = cleanImportPath(text)
			if text != "" && !seen[text] {
				seen[text] = true
				results = append(results, text)
			}
		}
	}
	return results
}

// cleanImportPath strips quotes, normalizes separators for cross-language matching.
func cleanImportPath(s string) string {
	// Strip surrounding quotes (single, double, backtick).
	s = strings.Trim(s, "\"'`<>")
	// Normalize C# dots and Rust :: to /
	s = strings.ReplaceAll(s, "::", "/")
	// Strip "crate/" prefix from Rust
	s = strings.TrimPrefix(s, "crate/")
	// Normalize PHP/C# backslashes
	s = strings.ReplaceAll(s, `\`, "/")
	// Normalize Python dots
	if strings.Contains(s, ".") && !strings.Contains(s, "/") {
		s = strings.ReplaceAll(s, ".", "/")
	}
	return strings.TrimSpace(s)
}
