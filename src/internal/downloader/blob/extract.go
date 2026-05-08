package blob

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func ExtractAndCleanupDefaultBlobWriter(baseDir string) error {
	defaultMu.Lock()
	defer defaultMu.Unlock()

	if defaultBlobWriter == nil {
		return nil
	}

	bw := defaultBlobWriter
	basePath := bw.baseOutPath
	baseIndex := bw.baseIndexPath

	if !blobWriterSilent {
		fmt.Printf("[BlobWriter] Extracting and cleaning up blob files...\n")
	}

	if _, err := os.Stat(basePath); err == nil {
		if _, err := os.Stat(baseIndex); err == nil {
			if err := ExtractBlobFile(basePath, baseIndex, baseDir); err != nil {
				if !blobWriterSilent {
					fmt.Printf("[BlobWriter] Warning: Failed to extract base blob: %v\n", err)
				}
			}
		}
		_ = os.Remove(basePath)
		_ = os.Remove(baseIndex)
	}

	for i := 2; i <= bw.rotationSeq; i++ {
		numberedPath := fmt.Sprintf("%s.%d", basePath, i)
		numberedIndex := fmt.Sprintf("%s.%d", baseIndex, i)

		if _, err := os.Stat(numberedPath); err == nil {
			if _, err := os.Stat(numberedIndex); err == nil {
				if err := ExtractBlobFile(numberedPath, numberedIndex, baseDir); err != nil {
					if !blobWriterSilent {
						fmt.Printf("[BlobWriter] Warning: Failed to extract blob %d: %v\n", i, err)
					}
				}
			}
			_ = os.Remove(numberedPath)
			_ = os.Remove(numberedIndex)
		}
	}

	_ = os.Remove(basePath + ".tmp")
	_ = os.Remove(baseIndex + ".tmp")

	if !blobWriterSilent {
		fmt.Printf("[BlobWriter] Extract and cleanup complete\n")
	}
	defaultBlobWriter = nil
	return nil
}

func ExtractBlobFile(blobPath, indexPath, baseDir string) error {
	indexFile, err := os.Open(indexPath)
	if err != nil {
		return fmt.Errorf("failed to open index: %w", err)
	}
	defer indexFile.Close()

	var index map[string]struct {
		Offset int64 `json:"offset"`
		Size   int64 `json:"size"`
	}
	dec := json.NewDecoder(indexFile)
	if err := dec.Decode(&index); err != nil {
		return fmt.Errorf("failed to decode index: %w", err)
	}

	blobFile, err := os.Open(blobPath)
	if err != nil {
		return fmt.Errorf("failed to open blob file: %w", err)
	}
	defer blobFile.Close()

	if !blobWriterSilent {
		fmt.Printf("[BlobWriter] Распаковка: %d файлов\n", len(index))
	}

	extractedCount := 0

	// Extract each file
	for fileName, entry := range index {
		// Create full path relative to baseDir
		fullPath := filepath.Join(baseDir, filepath.FromSlash(fileName))
		dir := filepath.Dir(fullPath)

		if err := os.MkdirAll(dir, 0755); err != nil {
			continue
		}

		// Read data from blob
		data := make([]byte, entry.Size)
		if _, err := blobFile.ReadAt(data, entry.Offset); err != nil {
			continue
		}

		if err := os.WriteFile(fullPath, data, 0644); err != nil {
			continue
		}

		extractedCount++
	}

	if !blobWriterSilent {
		fmt.Printf("[BlobWriter] ✅ Extracted %d files from blob\n", extractedCount)
	}
	return nil
}
