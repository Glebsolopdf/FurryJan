package downloader

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"furryjan/internal/api"
	"furryjan/internal/config"
	"furryjan/internal/downloader/blob"
)

func DownloadFileToDir(ctx context.Context, cfg *config.Config, targetDir string, post api.Post) (string, error) {
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	fileName := fmt.Sprintf("%d.%s", post.ID, post.File.Ext)
	filePath := filepath.Join(targetDir, fileName)

	blobActive := blob.DefaultBlobActive() && cfg.BlobWriterEnabled
	if blobActive {
		client := api.NewClientWithTimeout(cfg.Username, cfg.APIKey, cfg.RateLimitMS, api.DownloadTimeout)
		resp, err := client.DownloadFileWithProgressCtx(ctx, post.File.URL, post.File.Size)
		if err != nil {
			return "", fmt.Errorf("failed to download: %w", err)
		}
		defer resp.Body.Close()

		var buf bytes.Buffer
		if _, err := io.Copy(&buf, resp.Body); err != nil {
			return "", fmt.Errorf("failed to read response body: %w", err)
		}

		rel := filepath.ToSlash(filepath.Join(filepath.Base(targetDir), fileName))
		ref, _, err := blob.EnqueueDefaultBlobWriter(post.ID, rel, buf.Bytes())
		if err != nil {
			return "", fmt.Errorf("blob enqueue failed: %w", err)
		}
		return ref, nil
	}

	if _, err := os.Stat(filePath); err == nil {
		return filePath, nil
	}

	outFile, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer outFile.Close()

	client := api.NewClientWithTimeout(cfg.Username, cfg.APIKey, cfg.RateLimitMS, api.DownloadTimeout)
	resp, err := client.DownloadFileWithProgressCtx(ctx, post.File.URL, post.File.Size)
	if err != nil {
		os.Remove(filePath)
		return "", fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	written, err := io.Copy(outFile, resp.Body)
	outFile.Close()

	if err != nil {
		os.Remove(filePath)
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	if written == 0 {
		os.Remove(filePath)
		return "", fmt.Errorf("no data written")
	}

	// Verify file size matches expected size
	if post.File.Size > 0 && written != int64(post.File.Size) {
		os.Remove(filePath)
		return "", fmt.Errorf("incomplete download: got %d bytes, expected %d bytes", written, post.File.Size)
	}

	return filePath, nil
}
