package archiver

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"furryjan/internal/config"
)

type Options struct {
	IncludeTags []string
	ExcludeTags []string
	OutputPath  string
	Verbose     bool
}

func Run(cfg *config.Config, opts Options) error {
	filter := NewFilter(opts.IncludeTags, opts.ExcludeTags)

	dirList, err := filter.BuildList(cfg.DownloadDir)
	if err != nil {
		return fmt.Errorf("failed to build directory list: %w", err)
	}

	if len(dirList) == 0 {
		return fmt.Errorf("no directories to archive")
	}

	outFile, err := os.Create(opts.OutputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	zw := zip.NewWriter(outFile)
	defer zw.Close()

	var totalFiles int
	var totalSize int64
	var compressedSize int64

	for _, dirPath := range dirList {
		tagName := filepath.Base(dirPath)

		err := filepath.Walk(dirPath, func(filePath string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			rel, err := filepath.Rel(cfg.DownloadDir, filePath)
			if err != nil {
				return err
			}

			rel = strings.ReplaceAll(rel, string(filepath.Separator), "/")

			header, err := zip.FileInfoHeader(info)
			if err != nil {
				return err
			}
			header.Name = rel
			header.Method = zip.Deflate

			writer, err := zw.CreateHeader(header)
			if err != nil {
				return err
			}

			f, err := os.Open(filePath)
			if err != nil {
				return err
			}
			defer f.Close()

			_, err = io.Copy(writer, f)
			if err != nil {
				return err
			}

			totalFiles++
			totalSize += info.Size()

			if opts.Verbose {
				fmt.Printf("  + %s (%d bytes)\n", rel, info.Size())
			}

			return nil
		})

		if err != nil {
			return fmt.Errorf("error archiving %s: %w", tagName, err)
		}
	}

	err = zw.Close()
	if err != nil {
		return fmt.Errorf("failed to finalize archive: %w", err)
	}

	fileInfo, err := os.Stat(opts.OutputPath)
	if err == nil {
		compressedSize = fileInfo.Size()
	}

	ratio := float64(100)
	if totalSize > 0 {
		ratio = (1.0 - float64(compressedSize)/float64(totalSize)) * 100
	}

	fmt.Printf("✓ Архив создан: %s\n", opts.OutputPath)
	fmt.Printf("   Файлов: %d\n", totalFiles)
	fmt.Printf("   Исходный размер: %.1f MB\n", float64(totalSize)/1024/1024)
	fmt.Printf("   Размер архива: %.1f MB\n", float64(compressedSize)/1024/1024)
	fmt.Printf("   Сжатие: %.1f%%\n", ratio)

	return nil
}

func EstimateSize(downloadDir string, filter *Filter) (int64, error) {
	dirList, err := filter.BuildList(downloadDir)
	if err != nil {
		return 0, err
	}

	var totalSize int64

	for _, dirPath := range dirList {
		err := filepath.Walk(dirPath, func(filePath string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				totalSize += info.Size()
			}
			return nil
		})
		if err != nil {
			return 0, err
		}
	}

	return totalSize, nil
}
