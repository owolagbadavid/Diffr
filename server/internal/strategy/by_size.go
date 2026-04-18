package strategy

import "deniro/internal/model"

func init() { Register(bySize{}) }

type bySize struct{}

func (bySize) Name() string        { return "by-size" }
func (bySize) Description() string { return "Group files into small, medium, and large buckets" }

func (bySize) Organize(files []model.FileDiff) []model.FileGroup {
	var small, medium, large []model.FileDiff
	for _, f := range files {
		switch {
		case f.TotalChanges() <= 10:
			small = append(small, f)
		case f.TotalChanges() <= 100:
			medium = append(medium, f)
		default:
			large = append(large, f)
		}
	}

	var groups []model.FileGroup
	if len(small) > 0 {
		groups = append(groups, model.FileGroup{Name: "Small (≤10 lines)", Files: small})
	}
	if len(medium) > 0 {
		groups = append(groups, model.FileGroup{Name: "Medium (11–100 lines)", Files: medium})
	}
	if len(large) > 0 {
		groups = append(groups, model.FileGroup{Name: "Large (>100 lines)", Files: large})
	}
	return groups
}
