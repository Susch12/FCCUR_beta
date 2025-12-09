package models

import "time"

// Package represents a software package or course material in the repository
type Package struct {
	ID            int64     `json:"id"`
	Name          string    `json:"name"`
	Version       string    `json:"version"`
	Description   string    `json:"description,omitempty"`
	Category      string    `json:"category"`
	ContentType   string    `json:"content_type"` // "tool" or "material"
	CourseName    string    `json:"course_name,omitempty"`
	FilePath      string    `json:"file_path"`
	FileSize      int64     `json:"file_size"`
	BLAKE3Hash    string    `json:"blake3_hash"`
	SHA256Hash    string    `json:"sha256_hash"`
	DownloadURL   string    `json:"download_url,omitempty"`
	Platform      string    `json:"platform"`
	ThumbnailPath string    `json:"thumbnail_path,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// DownloadStats represents download statistics for a package
type DownloadStats struct {
	PackageID      int64     `json:"package_id"`
	PackageName    string    `json:"package_name"`
	TotalDownloads int       `json:"total_downloads"`
	LastDownload   time.Time `json:"last_download,omitempty"`
}
