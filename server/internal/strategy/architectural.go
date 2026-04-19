package strategy

import (
	"diffr/internal/model"
	"path"
	"strings"
)

func init() { Register(architectural{}) }

type architectural struct{}

func (architectural) Name() string        { return "architectural" }
func (architectural) Description() string { return "Group files by architectural layer" }

// layers are checked in order; first match wins.
var layers = []struct {
	name    string
	matcher func(string, string, string) bool // (lowerPath, base, ext)
}{
	{"API / Routes", func(lp, base, _ string) bool {
		for _, kw := range []string{"/api/", "/handler", "/router", "/route", "/controller", "/endpoint"} {
			if strings.Contains(lp, kw) {
				return true
			}
		}
		return false
	}},
	{"Models / Types", func(lp, base, _ string) bool {
		for _, kw := range []string{"/model/", "/models/", "/types/", "/schema/", "/entity/", "/entities/"} {
			if strings.Contains(lp, kw) {
				return true
			}
		}
		for _, kw := range []string{"types", "schema", "entity"} {
			if base == kw {
				return true
			}
		}
		return false
	}},
	{"Tests", func(lp, _, _ string) bool {
		return isTestFile(lp)
	}},
	{"Styles", func(_, _, ext string) bool {
		switch ext {
		case ".css", ".scss", ".less", ".sass", ".styled":
			return true
		}
		return false
	}},
	{"UI / Components", func(lp, _, ext string) bool {
		for _, dir := range []string{"/components/", "/pages/", "/views/", "/layouts/", "/templates/"} {
			if strings.Contains(lp, dir) {
				return true
			}
		}
		switch ext {
		case ".tsx", ".jsx", ".vue", ".svelte":
			return true
		}
		return false
	}},
	{"Services / Logic", func(lp, _, _ string) bool {
		for _, dir := range []string{"/service/", "/services/", "/usecase/", "/domain/", "/logic/", "/core/"} {
			if strings.Contains(lp, dir) {
				return true
			}
		}
		return false
	}},
	{"Configuration", func(lp, base, ext string) bool {
		switch ext {
		case ".yml", ".yaml", ".toml", ".ini", ".cfg", ".json":
			return true
		}
		switch base {
		case "dockerfile", "makefile", "cmakelists.txt",
			"go.mod", "go.sum", "package.json", "package-lock.json",
			"yarn.lock", "pnpm-lock.yaml", "cargo.lock", "cargo.toml",
			"requirements.txt", "pipfile", "gemfile",
			".gitignore", ".dockerignore", ".env", ".env.example",
			".eslintrc", ".prettierrc", "tsconfig.json", "vite.config.ts":
			return true
		}
		return false
	}},
	{"Documentation", func(_, _, ext string) bool {
		switch ext {
		case ".md", ".txt", ".rst", ".adoc":
			return true
		}
		return false
	}},
}

func (architectural) Organize(files []model.FileDiff) []model.FileGroup {
	buckets := make([][]model.FileDiff, len(layers)+1) // +1 for "Other"

	for _, f := range files {
		lp := strings.ToLower(f.Filename)
		base := strings.ToLower(path.Base(f.Filename))
		ext := strings.ToLower(path.Ext(f.Filename))

		matched := false
		for i, l := range layers {
			if l.matcher(lp, base, ext) {
				buckets[i] = append(buckets[i], f)
				matched = true
				break
			}
		}
		if !matched {
			buckets[len(layers)] = append(buckets[len(layers)], f)
		}
	}

	var groups []model.FileGroup
	for i, b := range buckets {
		if len(b) == 0 {
			continue
		}
		name := "Other"
		if i < len(layers) {
			name = layers[i].name
		}
		groups = append(groups, model.FileGroup{Name: name, Files: b})
	}
	return groups
}
