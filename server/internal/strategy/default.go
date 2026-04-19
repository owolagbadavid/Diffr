package strategy

import "diffr/internal/model"

func init() { Register(defaultStrategy{}) }

type defaultStrategy struct{}

func (defaultStrategy) Name() string        { return "default" }
func (defaultStrategy) Description() string { return "Files in the order GitHub returns them" }

func (defaultStrategy) Organize(files []model.FileDiff) []model.FileGroup {
	return []model.FileGroup{{Name: "All files", Files: files}}
}
