package blob

import (
	"fmt"
	"os"
	"sync"
)

// Package-level default writer and helpers
var (
	defaultBlobWriter *BlobWriter
	defaultMu         sync.Mutex
	blobWriterSilent  bool // Set to true to suppress [BlobWriter] logs
)

// StartDefaultBlobWriter initializes and starts the package default BlobWriter
func StartDefaultBlobWriter(outPath, indexPath string, bufSizeBytes int, autoCleanup bool, logLevel string) error {
	defaultMu.Lock()
	defer defaultMu.Unlock()
	if defaultBlobWriter != nil && defaultBlobWriter.isStarted() {
		return fmt.Errorf("default blob writer already running")
	}
	bw := NewBlobWriter(outPath, indexPath, bufSizeBytes, true, autoCleanup, logLevel)
	if err := bw.Start(); err != nil {
		return err
	}
	defaultBlobWriter = bw
	return nil
}

// SetBlobWriterSilent controls whether blob writer logs are suppressed
func SetBlobWriterSilent(silent bool) {
	blobWriterSilent = silent
}

// FlushDefaultBlobWriter waits for data to be written to disk
func FlushDefaultBlobWriter() error {
	defaultMu.Lock()
	bw := defaultBlobWriter
	defaultMu.Unlock()

	if bw == nil {
		return nil
	}

	// Signal writer to stop and close the channel safely
	bw.mu.Lock()
	if bw.stopped {
		// Already stopped, just return
		bw.mu.Unlock()
		return nil
	}
	bw.stopped = true
	bw.stopFlag = true

	// Safely close the channel using recover to avoid panic on double-close
	func() {
		defer func() {
			recover() // Ignore any panic from closing already-closed channel
		}()
		if !bw.channelClosed {
			close(bw.writeCh)
			bw.channelClosed = true
		}
	}()
	bw.mu.Unlock()

	// Wait for writer to finish
	err := <-bw.doneCh
	return err
}

func StopDefaultBlobWriter() error {
	defaultMu.Lock()
	bw := defaultBlobWriter
	defaultBlobWriter = nil
	defaultMu.Unlock()
	if bw == nil {
		return nil
	}
	err := bw.Close()
	return err
}

func CleanupDefaultBlobWriter() error {
	defaultMu.Lock()
	defer defaultMu.Unlock()
	if defaultBlobWriter == nil {
		return nil
	}
	bw := defaultBlobWriter
	basePath := bw.baseOutPath
	baseIndex := bw.baseIndexPath

	// Remove base files
	_ = os.Remove(basePath)
	_ = os.Remove(baseIndex)

	for i := 1; i <= bw.rotationSeq; i++ {
		numberedPath := fmt.Sprintf("%s.%d", basePath, i)
		numberedIndex := fmt.Sprintf("%s.%d", baseIndex, i)
		_ = os.Remove(numberedPath)
		_ = os.Remove(numberedIndex)
	}

	// Remove .tmp files if any
	_ = os.Remove(basePath + ".tmp")
	_ = os.Remove(baseIndex + ".tmp")
	return nil
}

func DefaultBlobActive() bool {
	defaultMu.Lock()
	defer defaultMu.Unlock()
	return defaultBlobWriter != nil && defaultBlobWriter.isStarted()
}

// EnqueueDefaultBlobWriter enqueues data into the default blob writer
func EnqueueDefaultBlobWriter(name string, data []byte) (string, int64, error) {
	defaultMu.Lock()
	bw := defaultBlobWriter
	defaultMu.Unlock()

	if bw == nil {
		return "", 0, fmt.Errorf("blob writer not running")
	}

	return bw.Enqueue(name, data)
}
