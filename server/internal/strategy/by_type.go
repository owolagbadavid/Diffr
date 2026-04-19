package strategy

import (
	"diffr/internal/model"
	"path"
	"sort"
	"strings"
)

func init() { Register(byType{}) }

type byType struct{}

func (byType) Name() string        { return "by-type" }
func (byType) Description() string { return "Group files by file extension" }

func (byType) Organize(files []model.FileDiff) []model.FileGroup {
	grouped := map[string][]model.FileDiff{}
	for _, f := range files {
		ext := strings.ToLower(path.Ext(f.Filename))
		if ext == "" {
			ext = "(no extension)"
		}
		grouped[ext] = append(grouped[ext], f)
	}

	keys := make([]string, 0, len(grouped))
	for k := range grouped {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	groups := make([]model.FileGroup, 0, len(keys))
	for _, k := range keys {
		groups = append(groups, model.FileGroup{Name: k, Files: grouped[k]})
	}
	return groups
}
