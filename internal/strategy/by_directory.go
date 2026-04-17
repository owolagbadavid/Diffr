package strategy

import (
	"deniro/internal/model"
	"path"
	"sort"
)

func init() { Register(byDirectory{}) }

type byDirectory struct{}

func (byDirectory) Name() string        { return "by-directory" }
func (byDirectory) Description() string { return "Group files by their parent directory" }

func (byDirectory) Organize(files []model.FileDiff) []model.FileGroup {
	grouped := map[string][]model.FileDiff{}
	for _, f := range files {
		dir := path.Dir(f.Filename)
		if dir == "." {
			dir = "/"
		}
		grouped[dir] = append(grouped[dir], f)
	}

	// Stable order by directory name.
	dirs := make([]string, 0, len(grouped))
	for d := range grouped {
		dirs = append(dirs, d)
	}
	sort.Strings(dirs)

	groups := make([]model.FileGroup, 0, len(dirs))
	for _, d := range dirs {
		groups = append(groups, model.FileGroup{Name: d, Files: grouped[d]})
	}
	return groups
}
