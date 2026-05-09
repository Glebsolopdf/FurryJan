package blob

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"furryjan/internal/db"
)

// BlobWriter aggregates many small files into a single large blob on disk
// to reduce SSD write amplification. It exposes a simple Enqueue API that
// returns a blob reference string the caller can store in the DB.
type BlobWriter struct {
	outPath       string
	indexPath     string
	bufSize       int
	flushInterval time.Duration
	rotationSeq   int
	autoCleanup   bool   // automatically delete blob files after each cleanup
	baseOutPath   string // base path without sequence number
	baseIndexPath string // base path without sequence number
	logLevel      string // logging level: "DEBUG", "INFO", "WARN", "ERROR"
	indexStore    IndexStore

	mu      sync.Mutex
	started bool
	closed  bool

	writeCh chan *blobTask
	doneCh  chan error
	ctx     context.Context
	cancel  context.CancelFunc
}

type blobTask struct {
	postID int
	name   string
	data   []byte
	result chan blobResult
}

type blobResult struct {
	offset int64
	size   int64
	err    error
}

type IndexStore interface {
	UpsertBlobEntry(postID int, blobPath, fileName string, offset, size int64) error
	ListBlobEntries(blobPath string) ([]db.BlobEntry, error)
	DeleteBlobEntries(blobPath string) error
}

// NewBlobWriter creates a BlobWriter instance. bufSize is in bytes.
// If autoCleanup is true, it will delete blob files after flushing.
// logLevel controls logging verbosity: "DEBUG", "INFO", "WARN", "ERROR"
func NewBlobWriter(outPath, indexPath string, bufSize int, autoCleanup bool, logLevel string, indexStore IndexStore) *BlobWriter {
	if logLevel == "" {
		logLevel = "INFO"
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &BlobWriter{
		outPath:       outPath,
		indexPath:     indexPath,
		bufSize:       bufSize,
		flushInterval: 30 * time.Second,
		autoCleanup:   autoCleanup,
		logLevel:      logLevel,
		indexStore:    indexStore,
		baseOutPath:   outPath,
		baseIndexPath: indexPath,
		rotationSeq:   1,
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

func (b *BlobWriter) Enqueue(postID int, name string, data []byte) (string, int64, error) {
	if !b.isStarted() {
		return "", 0, fmt.Errorf("blob writer not started")
	}
	t := &blobTask{postID: postID, name: filepath.ToSlash(name), data: data, result: make(chan blobResult, 1)}
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
	return b.closeWithCleanup(b.autoCleanup)
}

// CloseWithoutCleanup stops the writer but keeps blob artifacts on disk.
func (b *BlobWriter) CloseWithoutCleanup() error {
	return b.closeWithCleanup(false)
}

func (b *BlobWriter) closeWithCleanup(withCleanup bool) error {
	b.mu.Lock()
	if !b.started || b.closed {
		b.mu.Unlock()
		return nil
	}
	b.closed = true
	b.debugLog("[BlobWriter] Close() called, signaling graceful shutdown\n")
	close(b.writeCh)
	b.mu.Unlock()

	err := <-b.doneCh
	b.debugLog("[BlobWriter] Shutdown complete, err=%v\n", err)
	b.cancel()

	if withCleanup && err == nil {
		b.debugLog("[BlobWriter] Cleaning up blob files...\n")
		b.cleanupAllBlobs()
	}

	return err
}

func (b *BlobWriter) Flush() error {
	if !b.isStarted() {
		return nil
	}

	ack := make(chan blobResult, 1)
	select {
	case b.writeCh <- &blobTask{result: ack}:
		res := <-ack
		return res.err
	case <-b.ctx.Done():
		return b.ctx.Err()
	}
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
	indexPath := b.baseIndexPath
	f, err := os.OpenFile(b.baseOutPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		b.doneCh <- fmt.Errorf("open output: %w", err)
		return
	}
	defer f.Close()

	w := bufio.NewWriterSize(f, b.bufSize)
	ticker := time.NewTicker(b.flushInterval)
	defer ticker.Stop()

	b.debugLog("[BlobWriter] Started with %d MB buffer, writing to: %s\n", b.bufSize/(1024*1024), b.baseOutPath)

	idx := make(map[string]struct {
		Offset int64 `json:"offset"`
		Size   int64 `json:"size"`
	})
	var offset int64
	var pendingBytes int64
	fileCount := 0

	flushToDisk := func(reason string) error {
		if pendingBytes == 0 {
			return nil
		}
		if err := w.Flush(); err != nil {
			return fmt.Errorf("flush: %w", err)
		}
		if err := f.Sync(); err != nil {
			return fmt.Errorf("fsync: %w", err)
		}

		if b.indexStore == nil {
			tmp := indexPath + ".tmp"
			tf, err := os.Create(tmp)
			if err != nil {
				return fmt.Errorf("create index tmp: %w", err)
			}
			enc := json.NewEncoder(tf)
			if err := enc.Encode(idx); err != nil {
				tf.Close()
				return fmt.Errorf("encode index: %w", err)
			}
			tf.Close()
			if err := os.Rename(tmp, indexPath); err != nil {
				return fmt.Errorf("rename index: %w", err)
			}
		}

		if !blobWriterSilent {
			fmt.Printf("[BlobWriter] Flushed %.1f MB to disk (%s)\n", float64(pendingBytes)/(1024*1024), reason)
		}
		pendingBytes = 0
		return nil
	}

	for {
		select {
		case <-b.ctx.Done():
			if err := flushToDisk("context cancel"); err != nil {
				b.doneCh <- err
				return
			}
			b.doneCh <- nil
			return
		case <-ticker.C:
			if err := flushToDisk("timer"); err != nil {
				b.doneCh <- err
				return
			}
		case task, ok := <-b.writeCh:
			if !ok {
				if err := flushToDisk("shutdown"); err != nil {
					b.doneCh <- err
					return
				}
				b.doneCh <- nil
				return
			}

			// Empty task is an explicit flush request.
			if task.name == "" && len(task.data) == 0 {
				task.result <- blobResult{err: flushToDisk("manual")}
				continue
			}

			n, err := w.Write(task.data)
			if err != nil {
				task.result <- blobResult{err: err}
				continue
			}

			idx[task.name] = struct {
				Offset int64 `json:"offset"`
				Size   int64 `json:"size"`
			}{Offset: offset, Size: int64(n)}

			if b.indexStore != nil {
				if err := b.indexStore.UpsertBlobEntry(task.postID, b.baseOutPath, task.name, offset, int64(n)); err != nil {
					task.result <- blobResult{err: fmt.Errorf("persist blob index: %w", err)}
					continue
				}
			}

			task.result <- blobResult{offset: offset, size: int64(n), err: nil}
			offset += int64(n)
			pendingBytes += int64(n)
			fileCount++

			if fileCount%10 == 0 {
				b.debugLog("[BlobWriter] Buffered: %d files\n", fileCount)
			}
		}
	}
}
