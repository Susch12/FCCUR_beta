package hash

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"testing"
	"time"
)

// Test basic dual hashing
func TestDualHash(t *testing.T) {
	data := []byte("Hello, FCCUR!")
	reader := bytes.NewReader(data)

	blake3Hash, sha256Hash, err := DualHash(reader)
	if err != nil {
		t.Fatalf("DualHash failed: %v", err)
	}

	if blake3Hash == "" {
		t.Error("BLAKE3 hash is empty")
	}

	if sha256Hash == "" {
		t.Error("SHA256 hash is empty")
	}

	// Verify hash lengths
	if len(blake3Hash) != 64 {
		t.Errorf("BLAKE3 hash length = %d, want 64", len(blake3Hash))
	}

	if len(sha256Hash) != 64 {
		t.Errorf("SHA256 hash length = %d, want 64", len(sha256Hash))
	}
}

// Test progress callback
func TestDualHashWithProgress(t *testing.T) {
	size := 10 * 1024 * 1024 // 10MB
	data := make([]byte, size)
	rand.Read(data)

	var progressCalls int
	var lastProgress int64

	progress := func(bytesProcessed int64) {
		progressCalls++
		lastProgress = bytesProcessed
	}

	reader := bytes.NewReader(data)
	_, _, err := DualHashWithProgress(reader, progress)

	if err != nil {
		t.Fatalf("DualHashWithProgress failed: %v", err)
	}

	if progressCalls == 0 {
		t.Error("Progress callback was never called")
	}

	if lastProgress != int64(size) {
		t.Errorf("Last progress = %d, want %d", lastProgress, size)
	}
}

// Test parallel hashing threshold
func TestParallelDualHash(t *testing.T) {
	tests := []struct {
		name string
		size int64
	}{
		{"Small file (1MB)", 1 * 1024 * 1024},
		{"Medium file (50MB)", 50 * 1024 * 1024},
		{"Large file (150MB)", 150 * 1024 * 1024},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := make([]byte, tt.size)
			rand.Read(data)

			reader := bytes.NewReader(data)
			blake3Hash, sha256Hash, err := ParallelDualHash(reader, tt.size, nil)

			if err != nil {
				t.Fatalf("ParallelDualHash failed: %v", err)
			}

			if blake3Hash == "" || sha256Hash == "" {
				t.Error("Hash is empty")
			}
		})
	}
}

// Benchmark sequential hashing (small file)
func BenchmarkDualHash_1MB(b *testing.B) {
	size := 1 * 1024 * 1024
	data := make([]byte, size)
	rand.Read(data)

	b.ResetTimer()
	b.SetBytes(int64(size))

	for i := 0; i < b.N; i++ {
		reader := bytes.NewReader(data)
		_, _, err := DualHash(reader)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark sequential hashing (medium file)
func BenchmarkDualHash_50MB(b *testing.B) {
	size := 50 * 1024 * 1024
	data := make([]byte, size)
	rand.Read(data)

	b.ResetTimer()
	b.SetBytes(int64(size))

	for i := 0; i < b.N; i++ {
		reader := bytes.NewReader(data)
		_, _, err := DualHash(reader)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark sequential hashing (large file)
func BenchmarkDualHash_200MB(b *testing.B) {
	size := 200 * 1024 * 1024
	data := make([]byte, size)
	rand.Read(data)

	b.ResetTimer()
	b.SetBytes(int64(size))

	for i := 0; i < b.N; i++ {
		reader := bytes.NewReader(data)
		_, _, err := DualHash(reader)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark parallel hashing (large file)
func BenchmarkParallelDualHash_200MB(b *testing.B) {
	size := int64(200 * 1024 * 1024)
	data := make([]byte, size)
	rand.Read(data)

	b.ResetTimer()
	b.SetBytes(size)

	for i := 0; i < b.N; i++ {
		reader := bytes.NewReader(data)
		_, _, err := ParallelDualHash(reader, size, nil)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Comparative benchmark showing improvement
func BenchmarkComparison(b *testing.B) {
	sizes := []struct {
		name string
		size int64
	}{
		{"10MB", 10 * 1024 * 1024},
		{"50MB", 50 * 1024 * 1024},
		{"100MB", 100 * 1024 * 1024},
		{"200MB", 200 * 1024 * 1024},
	}

	for _, s := range sizes {
		data := make([]byte, s.size)
		rand.Read(data)

		b.Run(fmt.Sprintf("Sequential_%s", s.name), func(b *testing.B) {
			b.SetBytes(s.size)
			start := time.Now()

			for i := 0; i < b.N; i++ {
				reader := bytes.NewReader(data)
				_, _, _ = DualHash(reader)
			}

			elapsed := time.Since(start)
			b.ReportMetric(float64(s.size)/elapsed.Seconds()/1024/1024, "MB/s")
		})

		b.Run(fmt.Sprintf("Parallel_%s", s.name), func(b *testing.B) {
			b.SetBytes(s.size)
			start := time.Now()

			for i := 0; i < b.N; i++ {
				reader := bytes.NewReader(data)
				_, _, _ = ParallelDualHash(reader, s.size, nil)
			}

			elapsed := time.Since(start)
			b.ReportMetric(float64(s.size)/elapsed.Seconds()/1024/1024, "MB/s")
		})
	}
}
