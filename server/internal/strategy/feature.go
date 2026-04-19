package strategy

import (
	"diffr/internal/model"
	"sort"
	"strings"
	"unicode"
)

func init() { Register(feature{}) }

type feature struct{}

func (feature) Name() string        { return "feature" }
func (feature) Description() string { return "Group files by detected feature across directories" }

func (feature) Organize(files []model.FileDiff) []model.FileGroup {
	if len(files) == 0 {
		return nil
	}

	// Extract keywords for each file.
	fileKeywords := make([][]string, len(files))
	for i, f := range files {
		fileKeywords[i] = extractKeywords(f.Filename)
	}

	// Count how many files each keyword appears in.
	kwCount := map[string][]int{} // keyword → list of file indices
	for i, kws := range fileKeywords {
		seen := map[string]bool{}
		for _, kw := range kws {
			if !seen[kw] {
				seen[kw] = true
				kwCount[kw] = append(kwCount[kw], i)
			}
		}
	}

	// Sort keywords by how many files they group (descending).
	type kwEntry struct {
		word  string
		files []int
	}
	var candidates []kwEntry
	for w, idxs := range kwCount {
		if len(idxs) >= 2 {
			candidates = append(candidates, kwEntry{w, idxs})
		}
	}
	sort.Slice(candidates, func(i, j int) bool {
		return len(candidates[i].files) > len(candidates[j].files)
	})

	// Greedily assign files to groups.
	assigned := make([]bool, len(files))
	var groups []model.FileGroup

	for _, c := range candidates {
		var groupFiles []model.FileDiff
		for _, idx := range c.files {
			if !assigned[idx] {
				groupFiles = append(groupFiles, files[idx])
				assigned[idx] = true
			}
		}
		if len(groupFiles) >= 2 {
			groups = append(groups, model.FileGroup{Name: c.word, Files: groupFiles})
		} else {
			// Unassign if we couldn't form a group.
			for _, idx := range c.files {
				for j, f := range groupFiles {
					if f.Filename == files[idx].Filename {
						_ = j
						assigned[idx] = false
					}
				}
			}
		}
	}

	// Collect unassigned files.
	var other []model.FileDiff
	for i, f := range files {
		if !assigned[i] {
			other = append(other, f)
		}
	}
	if len(other) > 0 {
		groups = append(groups, model.FileGroup{Name: "Other", Files: other})
	}

	return groups
}

var genericKeywords = map[string]bool{
	"src": true, "internal": true, "lib": true, "pkg": true,
	"cmd": true, "components": true, "pages": true, "utils": true,
	"index": true, "main": true, "app": true, "test": true,
	"spec": true, "tests": true, "server": true, "client": true,
	"dist": true, "build": true, "public": true, "assets": true,
	"go": true, "ts": true, "tsx": true, "js": true, "jsx": true,
	"py": true, "rs": true, "css": true, "scss": true, "md": true,
	"json": true, "yaml": true, "yml": true, "toml": true,
	"mod": true, "sum": true, "lock": true,
}

func extractKeywords(filename string) []string {
	// Split path into segments, then split each by separators and camelCase.
	parts := strings.Split(filename, "/")
	var keywords []string

	for _, part := range parts {
		// Remove extension from the last segment.
		if idx := strings.LastIndex(part, "."); idx > 0 {
			ext := part[idx+1:]
			part = part[:idx]
			// Don't add extensions as keywords (already filtered as generic).
			_ = ext
		}

		// Split by underscores and hyphens.
		for _, seg := range strings.FieldsFunc(part, func(r rune) bool {
			return r == '_' || r == '-' || r == '.'
		}) {
			// Split camelCase.
			for _, word := range splitCamelCase(seg) {
				w := strings.ToLower(word)
				if len(w) >= 2 && !genericKeywords[w] {
					keywords = append(keywords, w)
				}
			}
		}
	}
	return keywords
}

func splitCamelCase(s string) []string {
	var words []string
	start := 0
	runes := []rune(s)
	for i := 1; i < len(runes); i++ {
		if unicode.IsUpper(runes[i]) && (i+1 >= len(runes) || unicode.IsLower(runes[i+1]) || unicode.IsLower(runes[i-1])) {
			words = append(words, string(runes[start:i]))
			start = i
		}
	}
	words = append(words, string(runes[start:]))
	return words
}
