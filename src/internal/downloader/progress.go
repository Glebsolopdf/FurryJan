package downloader

import "fmt"

// Progress represents download progress
type Progress struct {
	Current   int
	Total     int
	FileName  string
	BytesPerS int64
}

// String returns progress bar representation
func (p *Progress) String() string {
	percentage := 0
	if p.Total > 0 {
		percentage = (p.Current * 100) / p.Total
	}

	barLength := 20
	filledLength := (percentage * barLength) / 100

	bar := ""
	for i := 0; i < barLength; i++ {
		if i < filledLength {
			bar += "█"
		} else {
			bar += "░"
		}
	}

	speedStr := fmt.Sprintf("%.1f MB/s", float64(p.BytesPerS)/1024/1024)
	return fmt.Sprintf("[%s] %3d%% %d/%d %s %s", bar, percentage, p.Current, p.Total, p.FileName, speedStr)
}
