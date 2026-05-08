package blob

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/schollz/progressbar/v3"
)

// BlobWriter aggregates many small files into a single large blob on disk
// to reduce SSD write amplification. It exposes a simple Enqueue API that
// returns a blob reference string the caller can store in the DB.
type BlobWriter struct {
	outPath       string
	indexPath     string
	bufSize       int
	rotationSeq   int    // current blob file sequence number
	autoRotate    bool   // automatically restart after flush to free RAM
	autoCleanup   bool   // automatically delete blob files after each flush
	baseOutPath   string // base path without sequence number
	baseIndexPath string // base path without sequence number
	stopFlag      bool   // signal to stop instead of auto-rotate
	logLevel      string // logging level: "DEBUG", "INFO", "WARN", "ERROR"

	mu      sync.Mutex
	started bool

	writeCh chan *blobTask
	doneCh  chan error
	ctx     context.Context
	cancel  context.CancelFunc
}

type blobTask struct {
	name   string
	data   []byte
	result chan blobResult
}

type blobResult struct {
	offset int64
	size   int64
	err    error
}

// NewBlobWriter creates a BlobWriter instance. bufSize is in bytes.
// If autoRotate is true, it will automatically restart and create new blob files after each flush.
// If autoCleanup is true, it will delete blob files after flushing.
// logLevel controls logging verbosity: "DEBUG", "INFO", "WARN", "ERROR"
func NewBlobWriter(outPath, indexPath string, bufSize int, autoRotate bool, autoCleanup bool, logLevel string) *BlobWriter {
	if logLevel == "" {
		logLevel = "INFO"
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &BlobWriter{
		outPath:       outPath,
		indexPath:     indexPath,
		bufSize:       bufSize,
		autoRotate:    autoRotate,
		autoCleanup:   autoCleanup,
		logLevel:      logLevel,
		baseOutPath:   outPath,
		baseIndexPath: indexPath,
		rotationSeq:   1,
		stopFlag:      false,
		writeCh:       make(chan *blobTask, 256),
		doneCh:        make(chan error, 1),
		ctx:           ctx,
		cancel:        cancel,
	}
}

func (b *BlobWriter) Start() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.started {
		return fmt.Errorf("blob writer already started")
	}
	b.started = true
	go b.run()
	return nil
}

func (b *BlobWriter) Enqueue(name string, data []byte) (string, int64, error) {
	if !b.isStarted() {
		return "", 0, fmt.Errorf("blob writer not started")
	}
	t := &blobTask{name: filepath.ToSlash(name), data: data, result: make(chan blobResult, 1)}
	select {
	case b.writeCh <- t:
		res := <-t.result
		if res.err != nil {
			return "", 0, res.err
		}
		ref := fmt.Sprintf("blob://%s?offset=%d&size=%d", b.outPath, res.offset, res.size)
		return ref, res.size, nil
	case <-b.ctx.Done():
		return "", 0, fmt.Errorf("blob writer shutting down")
	}
}

func (b *BlobWriter) isStarted() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.started
}

// debugLog outputs debug-level log messages (only if logLevel is DEBUG)
func (b *BlobWriter) debugLog(format string, args ...interface{}) {
	if b.logLevel == "DEBUG" {
		fmt.Printf(format, args...)
	}
}

// Close signals the writer to finish and waits for it to flush and write index.
// Close gracefully shuts down the writer and waits for final flush.
func (b *BlobWriter) Close() error {
	b.mu.Lock()
	if !b.started {
		b.mu.Unlock()
		return nil
	}
	b.debugLog("[BlobWriter] Close() called, signaling graceful shutdown\n")
	b.stopFlag = true
	close(b.writeCh)
	b.mu.Unlock()

	// Wait for goroutine to finish (with implicit timeout from context)
	// Don't cancel context yet - let goroutine finish naturally
	err := <-b.doneCh
	b.debugLog("[BlobWriter] Shutdown complete, err=%v\n", err)
	b.cancel()

	if b.autoCleanup && err == nil {
		b.debugLog("[BlobWriter] Cleaning up blob files...\n")
		b.cleanupAllBlobs()
	}

	return err
}

func (b *BlobWriter) cleanupAllBlobs() {
	for i := 1; i <= b.rotationSeq; i++ {
		var blobPath, indexPath string
		if i == 1 {
			blobPath = b.baseOutPath
			indexPath = b.baseIndexPath
		} else {
			blobPath = fmt.Sprintf("%s.%d", b.baseOutPath, i)
			indexPath = fmt.Sprintf("%s.%d", b.baseIndexPath, i)
		}

		if err := os.Remove(blobPath); err == nil && !blobWriterSilent {
			fmt.Printf("[BlobWriter] Deleted blob: %s\n", blobPath)
		}
		if err := os.Remove(indexPath); err == nil && !blobWriterSilent {
			fmt.Printf("[BlobWriter] Deleted index: %s\n", indexPath)
		}
	}
	if !blobWriterSilent {
		fmt.Printf("[BlobWriter] All blob files cleaned up\n")
	}
}

func (b *BlobWriter) run() {
	for {
		outPath := b.baseOutPath
		indexPath := b.baseIndexPath
		if b.rotationSeq > 1 {
			outPath = fmt.Sprintf("%s.%d", b.baseOutPath, b.rotationSeq)
			indexPath = fmt.Sprintf("%s.%d", b.baseIndexPath, b.rotationSeq)
		}

		f, err := os.OpenFile(outPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			b.doneCh <- fmt.Errorf("open output: %w", err)
			return
		}

		w := bufio.NewWriterSize(f, b.bufSize)
		b.debugLog("[BlobWriter] Started with %d MB buffer, writing to: %s (seq=%d)\n", b.bufSize/(1024*1024), outPath, b.rotationSeq)
		idx := make(map[string]struct {
			Offset int64 `json:"offset"`
			Size   int64 `json:"size"`
		})
		var offset int64
		var bufferedSize int64
		fileCount := 0

		// Drain loop: read until channel closed or rotation needed
		for task := range b.writeCh {
			n, err := w.Write(task.data)
			if err != nil {
				task.result <- blobResult{err: err}
				f.Close()
				continue
			}
			idx[task.name] = struct {
				Offset int64 `json:"offset"`
				Size   int64 `json:"size"`
			}{Offset: offset, Size: int64(n)}
			task.result <- blobResult{offset: offset, size: int64(n), err: nil}
			offset += int64(n)
			bufferedSize += int64(n)
			fileCount++
			if fileCount%10 == 0 {
				b.debugLog("[BlobWriter] Buffered: %d files, %.1f MB (no disk write yet)\n", fileCount, float64(bufferedSize)/(1024*1024))
			}
		}

		// Flush and fsync with progress bar
		b.debugLog("[BlobWriter] Flushing %d files (%.1f MB total) to disk...\n", fileCount, float64(bufferedSize)/(1024*1024))

		// Create progress bar for writing
		bar := progressbar.NewOptions64(
			bufferedSize,
			progressbar.OptionSetDescription("Завершение и запись на диск"),
			progressbar.OptionShowBytes(true),
			progressbar.OptionSetWidth(30),
		)

		if err := w.Flush(); err != nil {
			bar.Close()
			f.Close()
			b.doneCh <- fmt.Errorf("flush: %w", err)
			return
		}
		bar.Add64(bufferedSize / 2) // Show half progress after flush

		if err := f.Sync(); err != nil {
			bar.Close()
			f.Close()
			b.doneCh <- fmt.Errorf("fsync: %w", err)
			return
		}
		bar.Add64(bufferedSize - bufferedSize/2) // Complete the progress bar
		bar.Finish()
		if !blobWriterSilent {
			fmt.Printf("[BlobWriter] ✅ Successfully wrote %.1f MB to disk\n", float64(bufferedSize)/(1024*1024))
		}
		f.Close()

		// Write index atomically
		tmp := indexPath + ".tmp"
		tf, err := os.Create(tmp)
		if err != nil {
			b.doneCh <- fmt.Errorf("create index tmp: %w", err)
			return
		}
		enc := json.NewEncoder(tf)
		if err := enc.Encode(idx); err != nil {
			tf.Close()
			b.doneCh <- fmt.Errorf("encode index: %w", err)
			return
		}
		tf.Close()
		if err := os.Rename(tmp, indexPath); err != nil {
			b.doneCh <- fmt.Errorf("rename index: %w", err)
			return
		}

		if !blobWriterSilent {
			fmt.Printf("[BlobWriter] Index written to: %s, memory freed\n", indexPath)
		}

		// Note: Don't delete blob files here - they'll be extracted and cleaned up in ExtractAndCleanup()

		// Check if we should stop or continue rotating
		b.mu.Lock()
		shouldStop := b.stopFlag
		b.mu.Unlock()

		if shouldStop {
			b.debugLog("[BlobWriter] Stop flag set, exiting\n")
			b.doneCh <- nil
			return
		}

		// Auto-rotate: restart the loop with a new blob file and sequence number
		b.debugLog("[BlobWriter] Auto-rotating to next blob file...\n")
		b.rotationSeq++

		// Recreate writeCh for next iteration (old one is closed)
		// Check context first to avoid unnecessary allocation
		select {
		case <-b.ctx.Done():
			b.debugLog("[BlobWriter] Context cancelled during rotation\n")
			b.doneCh <- nil
			return
		default:
			b.writeCh = make(chan *blobTask, 256)
		}
	}
}
