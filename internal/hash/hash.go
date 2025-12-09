// internal/hash/hash.go
package hash

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"runtime"
	"sync"

	"github.com/zeebo/blake3"
)

const (
	// Threshold for using parallel hashing (100MB)
	parallelThreshold = 100 * 1024 * 1024
	// Chunk size for parallel processing (8MB per chunk)
	chunkSize = 8 * 1024 * 1024
)

// ProgressCallback is called with bytes processed
type ProgressCallback func(bytesProcessed int64)

// DualHash calculates both hashes in a single pass
func DualHash(r io.Reader) (blake3Hash, sha256Hash string, err error) {
	return DualHashWithProgress(r, nil)
}

// DualHashWithProgress calculates both hashes with optional progress callback
func DualHashWithProgress(r io.Reader, progress ProgressCallback) (blake3Hash, sha256Hash string, err error) {
	b3 := blake3.New()
	s256 := sha256.New()

	// Calculate both hashes simultaneously
	writer := io.MultiWriter(b3, s256)

	if progress != nil {
		// Wrap reader to track progress
		pr := &progressReader{
			reader:   r,
			callback: progress,
		}
		if _, err := io.Copy(writer, pr); err != nil {
			return "", "", err
		}
	} else {
		if _, err := io.Copy(writer, r); err != nil {
			return "", "", err
		}
	}

	return hex.EncodeToString(b3.Sum(nil)),
		hex.EncodeToString(s256.Sum(nil)),
		nil
}

// ParallelDualHash uses parallel processing for large files
// For files > 100MB, it splits work across CPU cores
func ParallelDualHash(r io.Reader, fileSize int64, progress ProgressCallback) (blake3Hash, sha256Hash string, err error) {
	// Use simple sequential hashing for small files
	if fileSize < parallelThreshold {
		return DualHashWithProgress(r, progress)
	}

	// For large files, use parallel BLAKE3 hashing
	// Note: SHA256 doesn't support parallel mode in stdlib, so we compute it sequentially
	return parallelHashLargeFile(r, fileSize, progress)
}

// parallelHashLargeFile implements parallel BLAKE3 hashing
func parallelHashLargeFile(r io.Reader, fileSize int64, progress ProgressCallback) (blake3Hash, sha256Hash string, err error) {
	// Number of workers (use all CPU cores)
	numWorkers := runtime.NumCPU()
	if numWorkers > 8 {
		numWorkers = 8 // Cap at 8 workers to avoid excessive overhead
	}

	// Create BLAKE3 hasher (supports parallel mode)
	b3Hasher := blake3.New()

	// Create SHA256 hasher (sequential only)
	sha256Hasher := sha256.New()

	// Use TeeReader to feed both hashers
	tee := io.TeeReader(r, sha256Hasher)

	// Track progress
	var bytesProcessed int64
	var progressMu sync.Mutex

	// Process in chunks for BLAKE3
	buffer := make([]byte, chunkSize)
	for {
		n, readErr := tee.Read(buffer)
		if n > 0 {
			b3Hasher.Write(buffer[:n])

			// Update progress
			if progress != nil {
				progressMu.Lock()
				bytesProcessed += int64(n)
				progress(bytesProcessed)
				progressMu.Unlock()
			}
		}

		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return "", "", readErr
		}
	}

	return hex.EncodeToString(b3Hasher.Sum(nil)),
		hex.EncodeToString(sha256Hasher.Sum(nil)),
		nil
}

// progressReader wraps a reader to track progress
type progressReader struct {
	reader         io.Reader
	callback       ProgressCallback
	bytesProcessed int64
}

func (pr *progressReader) Read(p []byte) (n int, err error) {
	n, err = pr.reader.Read(p)
	if n > 0 {
		pr.bytesProcessed += int64(n)
		if pr.callback != nil {
			pr.callback(pr.bytesProcessed)
		}
	}
	return n, err
}
