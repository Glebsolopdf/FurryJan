package db

import "fmt"

// Stats holds database statistics
type Stats struct {
	TotalFiles    int64
	TotalSize     int64
	TagCount      int
	FirstDownload string
	LastDownload  string
}

// GetStats returns database statistics
func (d *DB) GetStats() (*Stats, error) {
	stats := &Stats{}

	// Get total files and size
	err := d.conn.QueryRow(`
		SELECT COUNT(*), COALESCE(SUM(file_size), 0)
		FROM downloads
	`).Scan(&stats.TotalFiles, &stats.TotalSize)
	if err != nil {
		return nil, err
	}

	// Get unique tags count
	err = d.conn.QueryRow(`
		SELECT COUNT(DISTINCT file_path)
		FROM downloads
	`).Scan(&stats.TagCount)
	if err != nil {
		return nil, err
	}

	// Get first and last download
	err = d.conn.QueryRow(`
		SELECT 
			COALESCE(MIN(downloaded_at), ''),
			COALESCE(MAX(downloaded_at), '')
		FROM downloads
	`).Scan(&stats.FirstDownload, &stats.LastDownload)
	if err != nil {
		return nil, err
	}

	return stats, nil
}

// FormatBytes formats bytes to human-readable format
func FormatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
