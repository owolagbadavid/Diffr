package strategy

import (
	"diffr/internal/model"
	"path"
	"sort"
	"strings"
)

func init() { Register(importance{}) }

type importance struct{}

func (importance) Name() string        { return "importance" }
func (importance) Description() string { return "Order files by review priority — critical first" }

func (importance) Organize(files []model.FileDiff) []model.FileGroup {
	type ranked struct {
		file model.FileDiff
		tier int
		orig int
	}

	items := make([]ranked, len(files))
	for i, f := range files {
		items[i] = ranked{file: f, tier: fileTier(f.Filename), orig: i}
	}

	sort.SliceStable(items, func(i, j int) bool {
		if items[i].tier != items[j].tier {
			return items[i].tier < items[j].tier
		}
		return items[i].orig < items[j].orig
	})

	sorted := make([]model.FileDiff, len(items))
	for i, r := range items {
		sorted[i] = r.file
	}

	return []model.FileGroup{{Name: "By review priority", Files: sorted}}
}

// fileTier returns a priority tier (lower = review first).
func fileTier(filename string) int {
	lower := strings.ToLower(filename)
	base := strings.ToLower(path.Base(filename))
	ext := strings.ToLower(path.Ext(filename))

	// Tier 0: Critical — security, auth, middleware, permissions
	for _, kw := range []string{"auth", "security", "middleware", "permission", "rbac", "acl", "secret", "crypt"} {
		if strings.Contains(lower, kw) {
			return 0
		}
	}

	// Tier 1: Core — models, types, interfaces, schemas, migrations
	for _, kw := range []string{"model", "schema", "types", "migration", "entity", "interface"} {
		if strings.Contains(lower, kw) {
			return 1
		}
	}

	// Tier 5: Config & docs (check before logic to catch .yml, .json, .md early)
	switch ext {
	case ".yml", ".yaml", ".toml", ".ini", ".cfg":
		return 5
	case ".md", ".txt", ".rst":
		return 5
	}
	switch base {
	case "dockerfile", "makefile", "rakefile", "cmakelists.txt",
		"go.mod", "go.sum", "package.json", "package-lock.json",
		"yarn.lock", "pnpm-lock.yaml", "cargo.lock", "cargo.toml",
		"requirements.txt", "pipfile", "pipfile.lock", "gemfile",
		"gemfile.lock", ".gitignore", ".dockerignore", ".env",
		".env.example", "license", "readme.md":
		return 5
	}

	// Tier 4: Tests
	if isTestFile(lower) {
		return 4
	}

	// Tier 3: UI — components, pages, styles
	for _, dir := range []string{"/components/", "/pages/", "/views/", "/layouts/"} {
		if strings.Contains(lower, dir) {
			return 3
		}
	}
	switch ext {
	case ".css", ".scss", ".less", ".sass":
		return 3
	}

	// Tier 2: Logic — everything else (handlers, services, regular source)
	return 2
}

func isTestFile(lower string) bool {
	if strings.Contains(lower, "_test.") || strings.Contains(lower, ".test.") || strings.Contains(lower, ".spec.") {
		return true
	}
	if strings.Contains(lower, "__tests__/") || strings.Contains(lower, "/test/") || strings.Contains(lower, "/tests/") {
		return true
	}
	return false
}
