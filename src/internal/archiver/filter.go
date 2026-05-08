package archiver

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Filter handles tag-based filtering
type Filter struct {
	includeTags []string
	excludeTags []string
}

// NewFilter creates a new tag filter
func NewFilter(include, exclude []string) *Filter {
	return &Filter{
		includeTags: include,
		excludeTags: exclude,
	}
}

// BuildList returns list of directories matching the filter
func (f *Filter) BuildList(downloadDir string) ([]string, error) {
	var result []string

	// Read download directory
	entries, err := os.ReadDir(downloadDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read download dir: %w", err)
	}

	// Filter directories
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		tagName := entry.Name()

		// Check if should include this tag
		if !f.matchesFilter(tagName) {
			continue
		}

		dirPath := filepath.Join(downloadDir, tagName)
		result = append(result, dirPath)
	}

	return result, nil
}

// matchesFilter returns true if tag should be included
func (f *Filter) matchesFilter(tag string) bool {
	// If include tags specified, must match one of them
	if len(f.includeTags) > 0 {
		found := false
		for _, incTag := range f.includeTags {
			if strings.Contains(strings.ToLower(tag), strings.ToLower(incTag)) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// If exclude tags specified, must not match any of them
	if len(f.excludeTags) > 0 {
		for _, excTag := range f.excludeTags {
			if strings.Contains(strings.ToLower(tag), strings.ToLower(excTag)) {
				return false
			}
		}
	}

	return true
}
