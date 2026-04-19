package strategy

import (
	"diffr/internal/model"
	"diffr/internal/treesitter"
	"path"
	"strings"
)

func init() { Register(topological{}) }

type topological struct{}

func (topological) Name() string { return "topological" }
func (topological) Description() string {
	return "Order by dependencies — review imported files first"
}

func (topological) Organize(files []model.FileDiff) []model.FileGroup {
	n := len(files)
	if n == 0 {
		return nil
	}

	// Map filename stems and directory segments to file indices (multimap).
	// Supports namespace-based imports (C#, Java, PHP) that refer to
	// directories rather than individual files.
	idx := map[string][]int{}
	add := func(key string, i int) {
		idx[key] = append(idx[key], i)
	}
	for i, f := range files {
		add(f.Filename, i)
		base := strings.TrimSuffix(path.Base(f.Filename), path.Ext(f.Filename))
		add(base, i)
		dir := path.Dir(f.Filename)
		if dir != "." {
			add(path.Base(dir)+"/"+base, i)
			// Register directory segments for namespace matching.
			// e.g., "src/Models/User.cs" registers "Models" and "src/Models"
			parts := strings.Split(dir, "/")
			for j := len(parts) - 1; j >= 0; j-- {
				segment := strings.Join(parts[j:], "/")
				add(segment, i)
			}
		}
	}

	// Parse each file with tree-sitter to extract imports.
	fileImports := make([][]string, n)
	for i, f := range files {
		if f.Content == "" {
			continue
		}
		signals := treesitter.Analyze(f.Filename, []byte(f.Content))
		if signals != nil {
			fileImports[i] = signals.Imports
		}
	}

	// Build adjacency: deps[i] = set of file indices that file i imports.
	deps := make([]map[int]bool, n)
	for i := range deps {
		deps[i] = map[int]bool{}
	}

	for i, imports := range fileImports {
		for _, imp := range imports {
			for _, j := range matchImport(imp, idx, i) {
				deps[i][j] = true
			}
		}
	}

	// Kahn's algorithm: files with no in-PR dependencies come first.
	inDegree := make([]int, n)
	for i := range deps {
		inDegree[i] = len(deps[i])
	}

	queue := make([]int, 0, n)
	for i := 0; i < n; i++ {
		if inDegree[i] == 0 {
			queue = append(queue, i)
		}
	}

	sorted := make([]model.FileDiff, 0, n)
	visited := make([]bool, n)
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		if visited[cur] {
			continue
		}
		visited[cur] = true
		sorted = append(sorted, files[cur])

		for i := range deps {
			if deps[i][cur] {
				inDegree[i]--
				if inDegree[i] == 0 {
					queue = append(queue, i)
				}
			}
		}
	}

	// Append any files not reached (cycles or no imports).
	for i, f := range files {
		if !visited[i] {
			sorted = append(sorted, f)
		}
	}

	return []model.FileGroup{{Name: "Dependency order", Files: sorted}}
}

// matchImport tries to match an import path against known filenames.
// Returns all matched file indices, excluding self.
func matchImport(imp string, idx map[string][]int, self int) []int {
	// Try exact match first, then basename, then dir/base.
	// Return the first level that produces matches.
	if matches := lookup(idx, imp, self); len(matches) > 0 {
		return matches
	}
	base := path.Base(imp)
	base = strings.TrimSuffix(base, path.Ext(base))
	if matches := lookup(idx, base, self); len(matches) > 0 {
		return matches
	}
	dir := path.Dir(imp)
	if dir != "." {
		key := path.Base(dir) + "/" + base
		if matches := lookup(idx, key, self); len(matches) > 0 {
			return matches
		}
	}
	return nil
}

// lookup returns all indices for a key, excluding self.
func lookup(idx map[string][]int, key string, self int) []int {
	entries := idx[key]
	if len(entries) == 0 {
		return nil
	}
	var out []int
	for _, j := range entries {
		if j != self {
			out = append(out, j)
		}
	}
	return out
}
