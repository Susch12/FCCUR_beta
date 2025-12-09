package api

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"
)

// Allowed MIME types for uploads
var allowedMIMETypes = map[string]bool{
	// Archives
	"application/zip":                true,
	"application/x-tar":              true,
	"application/gzip":               true,
	"application/x-gzip":             true,
	"application/x-bzip2":            true,
	"application/x-7z-compressed":    true,
	"application/x-rar-compressed":   true,
	"application/x-xz":               true,

	// Executables and installers
	"application/x-executable":       true,
	"application/x-msdos-program":    true,
	"application/x-msdownload":       true,
	"application/vnd.microsoft.portable-executable": true,
	"application/x-deb":              true,
	"application/x-rpm":              true,
	"application/x-apple-diskimage":  true,

	// ISO images
	"application/x-iso9660-image":    true,
	"application/x-cd-image":         true,

	// Documents
	"application/pdf":                true,
	"application/msword":             true,
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
	"application/vnd.ms-excel":       true,
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": true,
	"application/vnd.ms-powerpoint":  true,
	"application/vnd.openxmlformats-officedocument.presentationml.presentation": true,
	"text/plain":                     true,
	"text/markdown":                  true,
	"text/html":                      true,

	// Media (for course materials)
	"image/png":                      true,
	"image/jpeg":                     true,
	"image/gif":                      true,
	"image/svg+xml":                  true,
	"video/mp4":                      true,
	"video/mpeg":                     true,
	"audio/mpeg":                     true,

	// Generic binary
	"application/octet-stream":       true,
}

// Allowed file extensions
var allowedExtensions = map[string]bool{
	// Archives
	".zip": true, ".tar": true, ".gz": true, ".tgz": true, ".bz2": true,
	".7z": true, ".rar": true, ".xz": true,

	// Executables
	".exe": true, ".msi": true, ".deb": true, ".rpm": true,
	".dmg": true, ".pkg": true, ".appimage": true,

	// ISO images
	".iso": true, ".img": true,

	// Documents
	".pdf": true, ".doc": true, ".docx": true, ".xls": true, ".xlsx": true,
	".ppt": true, ".pptx": true, ".txt": true, ".md": true, ".html": true,
	".odt": true, ".ods": true, ".odp": true,

	// Media
	".png": true, ".jpg": true, ".jpeg": true, ".gif": true, ".svg": true,
	".mp4": true, ".mpeg": true, ".mp3": true, ".wav": true,

	// Source code (allowed for course materials)
	".c": true, ".cpp": true, ".h": true, ".py": true, ".java": true,
	".js": true, ".go": true, ".rs": true, ".rb": true, ".sh": true,
}

// Dangerous patterns to reject
var dangerousPatterns = []*regexp.Regexp{
	regexp.MustCompile(`\.\.`),              // Directory traversal
	regexp.MustCompile(`^/`),                // Absolute paths
	regexp.MustCompile(`\\`),                // Windows path separators
	regexp.MustCompile(`[<>:"|?*]`),         // Invalid Windows characters
	regexp.MustCompile(`^\s+`),              // Leading whitespace
	regexp.MustCompile(`\s+$`),              // Trailing whitespace
}

// validateFileType checks if the file type is allowed
func validateFileType(header *multipart.FileHeader) error {
	// Check file extension
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if !allowedExtensions[ext] {
		return fmt.Errorf("file type not allowed: %s", ext)
	}

	// Check MIME type (if available)
	if len(header.Header["Content-Type"]) > 0 {
		mimeType := header.Header["Content-Type"][0]
		// Some MIME types include charset, extract just the type
		mimeType = strings.Split(mimeType, ";")[0]
		mimeType = strings.TrimSpace(mimeType)

		if !allowedMIMETypes[mimeType] {
			return fmt.Errorf("MIME type not allowed: %s", mimeType)
		}
	}

	return nil
}

// sanitizeFilename cleans and validates a filename
func sanitizeFilename(filename string) (string, error) {
	if filename == "" {
		return "", fmt.Errorf("filename cannot be empty")
	}

	// Check for dangerous patterns
	for _, pattern := range dangerousPatterns {
		if pattern.MatchString(filename) {
			return "", fmt.Errorf("filename contains invalid characters or patterns")
		}
	}

	// Get base name (remove any directory components)
	filename = filepath.Base(filename)

	// Limit filename length (255 bytes is typical filesystem limit)
	if len(filename) > 255 {
		ext := filepath.Ext(filename)
		nameWithoutExt := strings.TrimSuffix(filename, ext)
		// Keep extension and truncate name
		maxNameLen := 255 - len(ext)
		if maxNameLen > 0 {
			filename = nameWithoutExt[:maxNameLen] + ext
		} else {
			return "", fmt.Errorf("filename too long")
		}
	}

	// Replace problematic characters with underscores
	replacer := strings.NewReplacer(
		" ", "_",
		"(", "_",
		")", "_",
		"[", "_",
		"]", "_",
		"{", "_",
		"}", "_",
		"&", "and",
		"'", "",
		"\"", "",
	)
	filename = replacer.Replace(filename)

	// Remove multiple consecutive underscores
	multiUnderscore := regexp.MustCompile(`_{2,}`)
	filename = multiUnderscore.ReplaceAllString(filename, "_")

	// Trim leading/trailing underscores
	filename = strings.Trim(filename, "_")

	// Final validation
	if filename == "" || filename == "." || filename == ".." {
		return "", fmt.Errorf("invalid filename after sanitization")
	}

	return filename, nil
}

// detectContentType tries to detect the content type from file content
func detectContentType(file multipart.File) (string, error) {
	// Read first 512 bytes for detection
	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && n == 0 {
		return "", err
	}

	// Reset file pointer
	if seeker, ok := file.(interface{ Seek(int64, int) (int64, error) }); ok {
		seeker.Seek(0, 0)
	}

	// Detect content type
	contentType := http.DetectContentType(buffer[:n])
	return contentType, nil
}
