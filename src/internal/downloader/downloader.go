package downloader

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/schollz/progressbar/v3"

	"furryjan/internal/api"
	"furryjan/internal/config"
	"furryjan/internal/db"
	"furryjan/internal/downloader/blob"
)

type Options struct {
	Tags        []string
	Limit       int
	DryRun      bool
	Verbose     bool
	DownloadDir string // Custom download directory
}

type DownloadResult struct {
	Downloaded  int
	Skipped     int
	Failed      int
	TotalSize   int64
	DownloadDir string // Where files were saved
}

var fileTypeMap = map[string]string{
	"gif":  "animation",
	"webm": "animation",
	"mp4":  "video",
	"flv":  "video",
	"mkv":  "video",
	"avi":  "video",
	"mov":  "video",
	"wmv":  "video",
}

func GetFileType(ext string) string {
	if fileType, ok := fileTypeMap[strings.ToLower(ext)]; ok {
		return fileType
	}
	return "image"
}

func buildAllowedTypesMap(allowedTypes []string) map[string]bool {
	typeMap := make(map[string]bool, len(allowedTypes))
	for _, t := range allowedTypes {
		typeMap[t] = true
	}
	return typeMap
}

func IsFileAllowed(post api.Post, allowedTypesMap map[string]bool, maxSizeMB int) bool {
	// Check file type
	fileType := GetFileType(post.File.Ext)
	if !allowedTypesMap[fileType] {
		return false
	}

	// Check file size (convert bytes to MB)
	if maxSizeMB > 0 && int(post.File.Size/(1024*1024)) > maxSizeMB {
		return false
	}

	return true
}

func Run(cfg *config.Config, database *db.DB, opts Options) (*DownloadResult, error) {
	client := api.NewClient(cfg.Username, cfg.APIKey, cfg.RateLimitMS)

	result := &DownloadResult{}
	seenPosts := make(map[int]bool)
	allowedTypesMap := buildAllowedTypesMap(cfg.AllowedTypes)
	downloadDir := opts.DownloadDir
	if downloadDir == "" {
		downloadDir = cfg.DownloadDir
	}
	result.DownloadDir = downloadDir

	err := os.MkdirAll(downloadDir, 0755)
	if err != nil {
		return nil, fmt.Errorf("не удалось создать папку: %s\nПожалуйста, измените путь загрузки в настройках (опция 3).\nОригинальная ошибка: %v", downloadDir, err)
	}

	blobWriterActive := blob.DefaultBlobActive() && cfg.BlobWriterEnabled
	if cfg.BlobWriterEnabled && !blob.DefaultBlobActive() {
		fmt.Println("⚠️  ВНИМАНИЕ: Blob Writer включен в настройках, но не активен!")
		fmt.Println("   Это может означать ошибку при старте. Проверьте логи выше.")
		fmt.Println("   Используется режим прямой записи на диск.")
		fmt.Println()
	} else if blobWriterActive {
		fmt.Println("✓ Режим: Blob Writer (оптимизированная запись в памяти)")
	} else {
		fmt.Println("✓ Режим: Прямая запись на диск")
	}
	fmt.Println()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	var cancelled bool

	var perTagLimit int
	if opts.Limit > 0 && len(opts.Tags) > 0 {
		perTagLimit = opts.Limit / len(opts.Tags)
		if perTagLimit == 0 {
			perTagLimit = 1
		}
	}

	progressSize := int64(opts.Limit)
	if progressSize <= 0 {
		// Start with one page and increase max dynamically while scanning pages.
		progressSize = 320
	}

	bar := progressbar.NewOptions64(
		progressSize,
		progressbar.OptionSetDescription("Скачивание"),
		progressbar.OptionShowBytes(false),
		progressbar.OptionSetWidth(30),
		progressbar.OptionShowCount(),
		progressbar.OptionThrottle(120*time.Millisecond),
		progressbar.OptionSetWriter(os.Stderr),
	)

	processed := int64(0)
	knownTotal := progressSize

	go func() {
		<-sigChan
		cancelled = true
		bar.Close()
	}()

	for _, tag := range opts.Tags {
		if cancelled {
			break
		}

		var page = 1
		tagCount := 0

		for {
			if cancelled {
				break
			}

			if opts.Limit > 0 && tagCount >= perTagLimit {
				break
			}

			posts, err := client.GetPosts([]string{tag}, 320, page)
			if err != nil {
				fmt.Printf("Warning: Failed to fetch posts for tag '%s': %v\n", tag, err)
				break
			}

			bar.Describe(fmt.Sprintf("Скачивание | %s | стр:%d | ok:%d", tag, page, result.Downloaded))

			if len(posts) == 0 {
				break
			}

			if opts.Limit <= 0 {
				pageTotal := int64(len(posts))
				if processed+pageTotal > knownTotal {
					knownTotal = processed + pageTotal
					bar.ChangeMax64(knownTotal)
				}
			}

			for _, post := range posts {
				if cancelled {
					break
				}

				if seenPosts[post.ID] {
					if opts.Limit <= 0 {
						processed++
						bar.Add(1)
					}
					continue
				}

				if opts.Limit > 0 && tagCount >= perTagLimit {
					break
				}

				seenPosts[post.ID] = true

				downloaded, err := database.IsDownloaded(post.ID)
				if err != nil {
					if opts.Limit <= 0 {
						processed++
						bar.Add(1)
					}
					result.Failed++
					continue
				}

				if downloaded {
					if opts.Limit <= 0 {
						processed++
						bar.Add(1)
					}
					result.Skipped++
					continue
				}

				if !IsFileAllowed(post, allowedTypesMap, cfg.MaxSizeMB) {
					if opts.Limit <= 0 {
						processed++
						bar.Add(1)
					}
					result.Skipped++
					continue
				}

				if opts.DryRun {
					fmt.Printf("(dry-run) %d - %s\n", post.ID, post.File.URL)
					tagCount++
					result.Downloaded++
					if opts.Limit <= 0 {
						processed++
						bar.Add(1)
					} else {
						bar.Add(1)
					}
					continue
				}

				err = downloadFile(cfg, database, post, opts.Tags, downloadDir)
				if err != nil {
					fmt.Printf("✗ Failed to download %d: %v\n", post.ID, err)
					if opts.Limit <= 0 {
						processed++
						bar.Add(1)
					}
					result.Failed++
					continue
				}

				tagCount++
				result.Downloaded++
				if opts.Limit <= 0 {
					processed++
					bar.Add(1)
				} else {
					bar.Add(1)
				}
			}

			if len(posts) < 320 {
				break
			}

			page++
		}
	}

	bar.Finish()

	if cancelled {
		fmt.Println()
		fmt.Println("Загрузка отменена пользователем")

	}

	return result, nil
}

// downloadFile downloads a single file with tag-based directory selection
func downloadFile(cfg *config.Config, database *db.DB, post api.Post, userTags []string, baseDownloadDir string) error {
	// Check if this is a special order: filter
	if len(userTags) > 0 && strings.HasPrefix(userTags[0], "order:") {
		filterName := strings.TrimPrefix(userTags[0], "order:")
		targetDir := filepath.Join(baseDownloadDir, filterName)
		savedPath, err := DownloadFileToDir(cfg, targetDir, post)
		if err != nil {
			return err
		}

		// Save to database
		allTags := strings.Join(post.Tags.General, " ")
		err = database.SaveDownload(post.ID, allTags, savedPath, post.File.URL, int64(post.File.Size), post.File.Ext, "s")
		if err != nil {
			return fmt.Errorf("failed to save to database: %w", err)
		}
		return nil
	}

	var targetTag string
	if len(userTags) > 0 {
		targetTag = userTags[0]
	} else {
		if len(post.Tags.General) > 0 {
			targetTag = post.Tags.General[0]
		} else {
			targetTag = "untagged"
		}
	}

	targetDir := filepath.Join(baseDownloadDir, targetTag)
	savedPath, err := DownloadFileToDir(cfg, targetDir, post)
	if err != nil {
		return err
	}

	allTags := strings.Join(post.Tags.General, " ")
	err = database.SaveDownload(post.ID, allTags, savedPath, post.File.URL, int64(post.File.Size), post.File.Ext, "s")
	if err != nil {
		return fmt.Errorf("failed to save to database: %w", err)
	}

	return nil
}
