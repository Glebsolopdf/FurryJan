package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const (
	SchemaVersion = 1
)

// DB wraps SQLite database
type DB struct {
	conn *sql.DB
}

func Open(dbPath string) (*DB, error) {
	dir := filepath.Dir(dbPath)
	err := os.MkdirAll(dir, 0700)
	if err != nil {
		return nil, fmt.Errorf("cannot create database directory '%s': %w", dir, err)
	}

	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("cannot open database '%s': %w", dbPath, err)
	}

	err = conn.Ping()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("cannot connect to database '%s': %w", dbPath, err)
	}

	db := &DB{conn: conn}

	err = db.migrate()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("database migration failed: %w", err)
	}

	return db, nil
}

func (d *DB) Close() error {
	if d.conn != nil {
		return d.conn.Close()
	}
	return nil
}

func (d *DB) migrate() error {
	_, err := d.conn.Exec(`
		CREATE TABLE IF NOT EXISTS schema_version (
			version INTEGER PRIMARY KEY
		)
	`)
	if err != nil {
		return err
	}

	var version int
	err = d.conn.QueryRow("SELECT version FROM schema_version LIMIT 1").Scan(&version)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	if version < SchemaVersion {
		err = d.migrateV1()
		if err != nil {
			return err
		}

		if version == 0 {
			_, err = d.conn.Exec("INSERT INTO schema_version (version) VALUES (?)", SchemaVersion)
		} else {
			_, err = d.conn.Exec("UPDATE schema_version SET version = ?", SchemaVersion)
		}
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *DB) migrateV1() error {
	_, err := d.conn.Exec(`
		CREATE TABLE IF NOT EXISTS downloads (
			post_id       INTEGER PRIMARY KEY,
			tags          TEXT        NOT NULL,
			file_path     TEXT        NOT NULL,
			file_url      TEXT        NOT NULL,
			downloaded_at DATETIME    NOT NULL DEFAULT CURRENT_TIMESTAMP,
			file_size     INTEGER     NOT NULL DEFAULT 0,
			file_ext      TEXT        NOT NULL DEFAULT '',
			rating        TEXT        NOT NULL DEFAULT 's'
		)
	`)
	if err != nil {
		return err
	}

	_, err = d.conn.Exec(`
		CREATE INDEX IF NOT EXISTS idx_downloaded_at ON downloads(downloaded_at)
	`)
	return err
}

type Download struct {
	PostID       int
	Tags         string
	FilePath     string
	FileURL      string
	DownloadedAt time.Time
	FileSize     int64
	FileExt      string
	Rating       string
}

func (d *DB) IsDownloaded(postID int) (bool, error) {
	var exists int
	err := d.conn.QueryRow("SELECT 1 FROM downloads WHERE post_id = ? LIMIT 1", postID).Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (d *DB) SaveDownload(postID int, tags, filePath, fileURL string, fileSize int64, fileExt, rating string) error {
	_, err := d.conn.Exec(`
		INSERT INTO downloads (post_id, tags, file_path, file_url, downloaded_at, file_size, file_ext, rating)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`,
		postID,
		tags,
		filePath,
		fileURL,
		time.Now(),
		fileSize,
		fileExt,
		rating,
	)
	return err
}

func (d *DB) QueryHistory(limit int) ([]Download, error) {
	var downloads []Download

	var query string
	if limit > 0 {
		query = `
			SELECT post_id, tags, file_path, file_url, downloaded_at, file_size, file_ext, rating
			FROM downloads
			ORDER BY downloaded_at DESC
			LIMIT ?
		`
	} else {
		query = `
			SELECT post_id, tags, file_path, file_url, downloaded_at, file_size, file_ext, rating
			FROM downloads
			ORDER BY downloaded_at DESC
		`
	}

	var rows *sql.Rows
	var err error
	if limit > 0 {
		rows, err = d.conn.Query(query, limit)
	} else {
		rows, err = d.conn.Query(query)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var dl Download
		err := rows.Scan(&dl.PostID, &dl.Tags, &dl.FilePath, &dl.FileURL, &dl.DownloadedAt, &dl.FileSize, &dl.FileExt, &dl.Rating)
		if err != nil {
			return nil, err
		}
		downloads = append(downloads, dl)
	}

	return downloads, rows.Err()
}

func (d *DB) QueryHistoryByTag(tag string, limit int) ([]Download, error) {
	var downloads []Download

	query := `
		SELECT post_id, tags, file_path, file_url, downloaded_at, file_size, file_ext, rating
		FROM downloads
		WHERE tags LIKE ?
		ORDER BY downloaded_at DESC
	`
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	pattern := fmt.Sprintf("%%%s%%", tag)
	rows, err := d.conn.Query(query, pattern)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var dl Download
		err := rows.Scan(&dl.PostID, &dl.Tags, &dl.FilePath, &dl.FileURL, &dl.DownloadedAt, &dl.FileSize, &dl.FileExt, &dl.Rating)
		if err != nil {
			return nil, err
		}
		downloads = append(downloads, dl)
	}

	return downloads, rows.Err()
}
