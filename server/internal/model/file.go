package model

// FileDiff represents a single changed file in a pull request.
type FileDiff struct {
	Filename    string `json:"filename"`
	Status      string `json:"status"` // added, removed, modified, renamed
	Additions   int    `json:"additions"`
	Deletions   int    `json:"deletions"`
	Patch       string `json:"patch"`
	BlobURL     string `json:"blob_url"`
	ContentsURL string `json:"contents_url"`
	Content     string `json:"-"` // full file content for analysis, not sent to frontend
}

// TotalChanges returns the sum of additions and deletions.
func (f FileDiff) TotalChanges() int {
	return f.Additions + f.Deletions
}

// FileGroup is a named collection of file diffs, produced by a strategy.
type FileGroup struct {
	Name  string     `json:"name"`
	Files []FileDiff `json:"files"`
}
