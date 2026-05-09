package db

import "time"

type BlobEntry struct {
	PostID    int
	BlobPath  string
	FileName  string
	Offset    int64
	Size      int64
	CreatedAt time.Time
}

func (d *DB) UpsertBlobEntry(postID int, blobPath, fileName string, offset, size int64) error {
	_, err := d.conn.Exec(`
		INSERT INTO blob_entries (post_id, blob_path, file_name, offset, size, created_at)
		VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(post_id, blob_path)
		DO UPDATE SET file_name = excluded.file_name, offset = excluded.offset, size = excluded.size, created_at = CURRENT_TIMESTAMP
	`, postID, blobPath, fileName, offset, size)
	return err
}

func (d *DB) ListBlobEntries(blobPath string) ([]BlobEntry, error) {
	rows, err := d.conn.Query(`
		SELECT post_id, blob_path, file_name, offset, size, created_at
		FROM blob_entries
		WHERE blob_path = ?
	`, blobPath)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	entries := make([]BlobEntry, 0)
	for rows.Next() {
		var e BlobEntry
		if err := rows.Scan(&e.PostID, &e.BlobPath, &e.FileName, &e.Offset, &e.Size, &e.CreatedAt); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return entries, nil
}

func (d *DB) DeleteBlobEntries(blobPath string) error {
	_, err := d.conn.Exec("DELETE FROM blob_entries WHERE blob_path = ?", blobPath)
	return err
}
