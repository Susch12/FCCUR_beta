package api

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ArchiveFile represents a file within an archive
type ArchiveFile struct {
	Name string `json:"name"`
	Size int64  `json:"size"`
	IsDir bool   `json:"is_dir"`
}

// ArchiveContents holds information about an archive
type ArchiveContents struct {
	TotalFiles int64          `json:"total_files"`
	TotalSize  int64          `json:"total_size"`
	Files      []ArchiveFile  `json:"files"`
	Readme     string         `json:"readme,omitempty"`
}

// listZipContents lists the contents of a ZIP file
func listZipContents(filePath string) (*ArchiveContents, error) {
	r, err := zip.OpenReader(filePath)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	contents := &ArchiveContents{
		Files: make([]ArchiveFile, 0),
	}

	var readme string

	for _, f := range r.File {
		contents.TotalFiles++
		contents.TotalSize += int64(f.UncompressedSize64)

		contents.Files = append(contents.Files, ArchiveFile{
			Name:  f.Name,
			Size:  int64(f.UncompressedSize64),
			IsDir: f.FileInfo().IsDir(),
		})

		// Check for README
		if readme == "" && isReadmeFile(f.Name) {
			readme, _ = readZipFile(f)
		}
	}

	contents.Readme = readme
	return contents, nil
}

// listTarContents lists the contents of a TAR file (optionally gzipped)
func listTarContents(filePath string, isGzipped bool) (*ArchiveContents, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var tarReader *tar.Reader

	if isGzipped {
		gzReader, err := gzip.NewReader(file)
		if err != nil {
			return nil, err
		}
		defer gzReader.Close()
		tarReader = tar.NewReader(gzReader)
	} else {
		tarReader = tar.NewReader(file)
	}

	contents := &ArchiveContents{
		Files: make([]ArchiveFile, 0),
	}

	var readme string

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		contents.TotalFiles++
		contents.TotalSize += header.Size

		contents.Files = append(contents.Files, ArchiveFile{
			Name:  header.Name,
			Size:  header.Size,
			IsDir: header.FileInfo().IsDir(),
		})

		// Check for README
		if readme == "" && isReadmeFile(header.Name) && header.Size < 50000 { // Max 50KB
			readmeBytes := make([]byte, header.Size)
			if _, err := io.ReadFull(tarReader, readmeBytes); err == nil {
				readme = string(readmeBytes)
			}
		}
	}

	contents.Readme = readme
	return contents, nil
}

// isReadmeFile checks if a filename is a README file
func isReadmeFile(name string) bool {
	basename := strings.ToLower(filepath.Base(name))
	return strings.HasPrefix(basename, "readme")
}

// readZipFile reads the contents of a file from a ZIP archive
func readZipFile(f *zip.File) (string, error) {
	// Only read small files (max 50KB)
	if f.UncompressedSize64 > 50000 {
		return "", nil
	}

	rc, err := f.Open()
	if err != nil {
		return "", err
	}
	defer rc.Close()

	content := make([]byte, f.UncompressedSize64)
	_, err = io.ReadFull(rc, content)
	if err != nil {
		return "", err
	}

	return string(content), nil
}
