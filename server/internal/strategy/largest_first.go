package strategy

import (
	"diffr/internal/model"
	"sort"
)

func init() { Register(largestFirst{}) }

type largestFirst struct{}

func (largestFirst) Name() string        { return "largest-first" }
func (largestFirst) Description() string { return "Sort files by change size, largest first" }

func (largestFirst) Organize(files []model.FileDiff) []model.FileGroup {
	sorted := make([]model.FileDiff, len(files))
	copy(sorted, files)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].TotalChanges() > sorted[j].TotalChanges()
	})
	return []model.FileGroup{{Name: "Largest changes first", Files: sorted}}
}
