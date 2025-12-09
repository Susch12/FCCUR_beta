package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/jesus/FCCUR/internal/hash"
	"github.com/jesus/FCCUR/internal/models"
)

// Health returns server health status
func (s *Server) Health(w http.ResponseWriter, r *http.Request) {
	// Check database
	if err := s.db.Ping(); err != nil {
		http.Error(w, "Database unhealthy", http.StatusServiceUnavailable)
		return
	}

	// Check packages directory
	if _, err := os.Stat(s.packagesDir); os.IsNotExist(err) {
		http.Error(w, "Packages directory missing", http.StatusServiceUnavailable)
		return
	}

	// Get stats
	stats, _ := s.db.GetStats()
	totalDownloads := 0
	for _, s := range stats {
		totalDownloads += s.TotalDownloads
	}

	response := map[string]interface{}{
		"status":          "ok",
		"total_packages":  len(stats),
		"total_downloads": totalDownloads,
		"uptime":          time.Since(s.startTime).String(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetPackages lists all packages
func (s *Server) GetPackages(w http.ResponseWriter, r *http.Request) {
	// Try to get from cache first
	if packages, ok := s.cache.Get(); ok {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Cache", "HIT")
		json.NewEncoder(w).Encode(packages)
		return
	}

	// Cache miss - fetch from database
	packages, err := s.db.GetPackages()
	if err != nil {
		http.Error(w, "Error fetching packages", http.StatusInternalServerError)
		return
	}

	// Update cache
	s.cache.Set(packages)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Cache", "MISS")
	json.NewEncoder(w).Encode(packages)
}

// GetPackage retrieves a specific package
func (s *Server) GetPackage(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	pkg, err := s.db.GetPackage(id)
	if err != nil {
		http.Error(w, "Package not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pkg)
}

// UploadPackage handles package upload
func (s *Server) UploadPackage(w http.ResponseWriter, r *http.Request) {
	// Limit size to 10GB
	if err := r.ParseMultipartForm(10 << 30); err != nil {
		http.Error(w, "File too large", http.StatusBadRequest)
		return
	}

	// Get file
	file, header, err := r.FormFile("package")
	if err != nil {
		http.Error(w, "Error reading file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Validate file type
	if err := validateFileType(header); err != nil {
		http.Error(w, fmt.Sprintf("File validation error: %v", err), http.StatusBadRequest)
		return
	}

	// Sanitize filename
	sanitizedFilename, err := sanitizeFilename(header.Filename)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid filename: %v", err), http.StatusBadRequest)
		return
	}

	// Get metadata
	name := r.FormValue("name")
	version := r.FormValue("version")
	category := r.FormValue("category")
	platform := r.FormValue("platform")
	description := r.FormValue("description")
	contentType := r.FormValue("content_type") // "tool" or "material"
	courseName := r.FormValue("course_name")    // Optional, for materials

	// Default to "tool" if not specified
	if contentType == "" {
		contentType = "tool"
	}

	// Validate required fields
	if name == "" || version == "" || category == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// For materials, course_name is recommended
	if contentType == "material" && courseName == "" {
		courseName = "General"
	}

	// Check course permissions for professors
	if contentType == "material" && courseName != "" {
		if !s.checkCoursePermission(w, r, courseName) {
			claims, _ := s.getCurrentUser(r)
			user, _ := s.db.GetUserByID(claims.UserID)
			// Only restrict professors, admins can upload to any course
			if user != nil && user.Role == models.RoleProfessor {
				respondJSON(w, http.StatusForbidden, map[string]string{
					"error": "You don't have permission to upload to this course",
				})
				return
			}
		}
	}

	// Save file with sanitized filename
	destPath := filepath.Join(s.packagesDir, sanitizedFilename)
	dest, err := os.Create(destPath)
	if err != nil {
		http.Error(w, "Error saving file", http.StatusInternalServerError)
		return
	}
	defer dest.Close()

	// Copy and calculate hashes
	blake3Hash, sha256Hash, fileSize, err := s.saveAndHash(file, dest)
	if err != nil {
		os.Remove(destPath)
		http.Error(w, "Error processing file", http.StatusInternalServerError)
		return
	}

	// Handle optional thumbnail upload
	var thumbnailPath string
	thumbFile, thumbHeader, err := r.FormFile("thumbnail")
	if err == nil {
		// Thumbnail provided
		defer thumbFile.Close()

		// Validate thumbnail (PNG/JPG only, max 5MB)
		thumbExt := strings.ToLower(filepath.Ext(thumbHeader.Filename))
		if thumbExt != ".png" && thumbExt != ".jpg" && thumbExt != ".jpeg" {
			// Ignore invalid thumbnail, don't fail upload
			log.Printf("Invalid thumbnail format: %s", thumbExt)
		} else if thumbHeader.Size > 5*1024*1024 {
			log.Printf("Thumbnail too large: %d bytes", thumbHeader.Size)
		} else {
			// Save thumbnail
			thumbFilename := fmt.Sprintf("thumb_%d%s", time.Now().UnixNano(), thumbExt)
			thumbPath := filepath.Join(s.packagesDir, "thumbnails", thumbFilename)

			// Ensure thumbnails directory exists
			os.MkdirAll(filepath.Join(s.packagesDir, "thumbnails"), 0755)

			thumbDest, err := os.Create(thumbPath)
			if err == nil {
				defer thumbDest.Close()
				if _, err := io.Copy(thumbDest, thumbFile); err == nil {
					thumbnailPath = thumbPath
				} else {
					log.Printf("Error saving thumbnail: %v", err)
				}
			}
		}
	}

	// Create database record
	pkg := &models.Package{
		Name:          name,
		Version:       version,
		Description:   description,
		Category:      category,
		ContentType:   contentType,
		CourseName:    courseName,
		FilePath:      destPath,
		FileSize:      fileSize,
		BLAKE3Hash:    blake3Hash,
		SHA256Hash:    sha256Hash,
		Platform:      platform,
		ThumbnailPath: thumbnailPath,
	}

	id, err := s.db.CreatePackage(pkg)
	if err != nil {
		os.Remove(destPath)
		http.Error(w, "Error creating package", http.StatusInternalServerError)
		return
	}

	pkg.ID = id

	// Invalidate cache after successful upload
	s.cache.Invalidate()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(pkg)
}

// DownloadPackage streams a package file
func (s *Server) DownloadPackage(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	// Get package metadata
	pkg, err := s.db.GetPackage(id)
	if err != nil {
		http.Error(w, "Package not found", http.StatusNotFound)
		return
	}

	// Open file
	file, err := os.Open(pkg.FilePath)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	defer file.Close()

	// Record download asynchronously
	go s.db.RecordDownload(id, r.RemoteAddr, r.UserAgent())

	// Set headers
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filepath.Base(pkg.FilePath)))
	w.Header().Set("Content-Length", strconv.FormatInt(pkg.FileSize, 10))
	w.Header().Set("X-BLAKE3-Hash", pkg.BLAKE3Hash)
	w.Header().Set("X-SHA256-Hash", pkg.SHA256Hash)

	// Stream file
	io.Copy(w, file)
}

// GetStats returns download statistics
func (s *Server) GetStats(w http.ResponseWriter, r *http.Request) {
	stats, err := s.db.GetStats()
	if err != nil {
		log.Printf("Error fetching stats: %v", err)
		http.Error(w, "Error fetching stats", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// CheckDuplicate checks if a package with the same BLAKE3 hash already exists
func (s *Server) CheckDuplicate(w http.ResponseWriter, r *http.Request) {
	blake3Hash := r.URL.Query().Get("hash")
	if blake3Hash == "" {
		http.Error(w, "Missing hash parameter", http.StatusBadRequest)
		return
	}

	// Validate hash format (should be 64 hex characters)
	if len(blake3Hash) != 64 {
		http.Error(w, "Invalid hash format", http.StatusBadRequest)
		return
	}

	// Check if package with this hash exists
	pkg, err := s.db.FindPackageByHash(blake3Hash)
	if err != nil {
		// No duplicate found
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"duplicate": false,
		})
		return
	}

	// Duplicate found
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"duplicate": true,
		"package":   pkg,
	})
}

// DownloadChecksum generates and downloads checksum file for a package
func (s *Server) DownloadChecksum(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	checksumType := r.URL.Query().Get("type") // "sha256" or "blake3"

	if checksumType == "" {
		checksumType = "sha256"
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	pkg, err := s.db.GetPackage(id)
	if err != nil {
		http.Error(w, "Package not found", http.StatusNotFound)
		return
	}

	// Get filename from path
	filename := filepath.Base(pkg.FilePath)

	var checksumContent string
	var checksumFilename string

	switch checksumType {
	case "blake3":
		checksumContent = fmt.Sprintf("%s  %s\n", pkg.BLAKE3Hash, filename)
		checksumFilename = filename + ".b3sum"
	case "sha256":
		checksumContent = fmt.Sprintf("%s  %s\n", pkg.SHA256Hash, filename)
		checksumFilename = filename + ".sha256"
	default:
		http.Error(w, "Invalid checksum type", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", checksumFilename))
	w.Header().Set("Content-Length", strconv.Itoa(len(checksumContent)))
	w.Write([]byte(checksumContent))
}

// DownloadAllChecksums generates a batch checksum file for all packages
func (s *Server) DownloadAllChecksums(w http.ResponseWriter, r *http.Request) {
	checksumType := r.URL.Query().Get("type") // "sha256" or "blake3"

	if checksumType == "" {
		checksumType = "sha256"
	}

	packages, err := s.db.GetPackages()
	if err != nil {
		http.Error(w, "Error fetching packages", http.StatusInternalServerError)
		return
	}

	var checksumContent strings.Builder
	checksumContent.WriteString(fmt.Sprintf("# FCCUR Checksums (%s)\n", strings.ToUpper(checksumType)))
	checksumContent.WriteString(fmt.Sprintf("# Generated: %s\n", time.Now().Format(time.RFC3339)))
	checksumContent.WriteString(fmt.Sprintf("# Total packages: %d\n\n", len(packages)))

	for _, pkg := range packages {
		filename := filepath.Base(pkg.FilePath)
		switch checksumType {
		case "blake3":
			checksumContent.WriteString(fmt.Sprintf("%s  %s\n", pkg.BLAKE3Hash, filename))
		case "sha256":
			checksumContent.WriteString(fmt.Sprintf("%s  %s\n", pkg.SHA256Hash, filename))
		}
	}

	var checksumFilename string
	switch checksumType {
	case "blake3":
		checksumFilename = "fccur-checksums.b3sum"
	case "sha256":
		checksumFilename = "fccur-checksums.sha256"
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", checksumFilename))
	w.Write([]byte(checksumContent.String()))
}

// GetArchiveContents lists the contents of an archive (ZIP/TAR/TAR.GZ)
func (s *Server) GetArchiveContents(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	pkg, err := s.db.GetPackage(id)
	if err != nil {
		http.Error(w, "Package not found", http.StatusNotFound)
		return
	}

	// Check if file exists
	if _, err := os.Stat(pkg.FilePath); os.IsNotExist(err) {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	// Determine archive type from extension
	ext := strings.ToLower(filepath.Ext(pkg.FilePath))
	var contents interface{}
	var archiveErr error

	switch ext {
	case ".zip":
		contents, archiveErr = listZipContents(pkg.FilePath)
	case ".tar":
		contents, archiveErr = listTarContents(pkg.FilePath, false)
	case ".gz", ".tgz":
		// Check if it's a tar.gz
		if strings.HasSuffix(strings.ToLower(pkg.FilePath), ".tar.gz") || ext == ".tgz" {
			contents, archiveErr = listTarContents(pkg.FilePath, true)
		} else {
			http.Error(w, "Unsupported archive format", http.StatusBadRequest)
			return
		}
	default:
		http.Error(w, "Not an archive file", http.StatusBadRequest)
		return
	}

	if archiveErr != nil {
		log.Printf("Error reading archive: %v", archiveErr)
		http.Error(w, "Error reading archive", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(contents)
}

// ServeThumbnail serves package thumbnail images
func (s *Server) ServeThumbnail(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	pkg, err := s.db.GetPackage(id)
	if err != nil {
		http.Error(w, "Package not found", http.StatusNotFound)
		return
	}

	// Check if package has a thumbnail
	if pkg.ThumbnailPath == "" {
		// Return default icon based on category
		s.serveDefaultIcon(w, pkg.Category)
		return
	}

	// Check if thumbnail file exists
	if _, err := os.Stat(pkg.ThumbnailPath); os.IsNotExist(err) {
		s.serveDefaultIcon(w, pkg.Category)
		return
	}

	// Serve the thumbnail file
	http.ServeFile(w, r, pkg.ThumbnailPath)
}

// serveDefaultIcon serves a default SVG icon based on category
func (s *Server) serveDefaultIcon(w http.ResponseWriter, category string) {
	var svg string
	switch category {
	case "os":
		svg = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100"><rect fill="#2563eb" width="100" height="100"/><text x="50" y="55" font-size="40" text-anchor="middle" fill="white" font-family="Arial">OS</text></svg>`
	case "compiler":
		svg = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100"><rect fill="#10b981" width="100" height="100"/><text x="50" y="55" font-size="35" text-anchor="middle" fill="white" font-family="Arial">⚙️</text></svg>`
	case "ide":
		svg = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100"><rect fill="#f59e0b" width="100" height="100"/><text x="50" y="55" font-size="30" text-anchor="middle" fill="white" font-family="Arial">IDE</text></svg>`
	case "library":
		svg = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100"><rect fill="#8b5cf6" width="100" height="100"/><text x="50" y="60" font-size="50" text-anchor="middle" fill="white">📚</text></svg>`
	default: // "tool" and others
		svg = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100"><rect fill="#6b7280" width="100" height="100"/><text x="50" y="60" font-size="50" text-anchor="middle" fill="white">🛠️</text></svg>`
	}

	w.Header().Set("Content-Type", "image/svg+xml")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	w.Write([]byte(svg))
}

// DeletePackage removes a package (requires authentication)
func (s *Server) DeletePackage(w http.ResponseWriter, r *http.Request) {
	// ALWAYS require authentication for deletion, regardless of global auth settings
	if !s.authConfig.Enabled {
		http.Error(w, "Delete endpoint requires authentication to be configured", http.StatusForbidden)
		return
	}

	idStr := r.URL.Query().Get("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	// Get package metadata before deletion
	pkg, err := s.db.GetPackage(id)
	if err != nil {
		http.Error(w, "Package not found", http.StatusNotFound)
		return
	}

	// Delete from database (will delete download records too)
	if err := s.db.DeletePackage(id); err != nil {
		log.Printf("Error deleting package from database: %v", err)
		http.Error(w, "Error deleting package", http.StatusInternalServerError)
		return
	}

	// Delete the file
	if err := os.Remove(pkg.FilePath); err != nil {
		log.Printf("Warning: Error deleting file %s: %v", pkg.FilePath, err)
		// Continue - database is already updated
	}

	// Invalidate cache after successful deletion
	s.cache.Invalidate()

	// Log deletion event
	log.Printf("Package deleted: ID=%d, Name=%s, Version=%s, File=%s",
		pkg.ID, pkg.Name, pkg.Version, pkg.FilePath)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Package deleted successfully",
		"id":      id,
	})
}

// saveAndHash saves file and calculates both hashes simultaneously
// Uses optimized parallel hashing for files > 100MB
func (s *Server) saveAndHash(src io.Reader, dest io.Writer) (blake3Hash, sha256Hash string, size int64, err error) {
	// Create a tee reader to write to dest and hash at the same time
	pr, pw := io.Pipe()

	// Create a multi-writer to write to both dest and pipe
	mw := io.MultiWriter(dest, pw)

	// Channel to receive hash results
	hashResult := make(chan struct {
		b3Hash   string
		s256Hash string
		err      error
	})

	// Calculate hashes in goroutine with progress tracking
	go func() {
		// Progress callback (could be expanded to send websocket updates)
		progressCallback := func(bytesProcessed int64) {
			// For now, just track internally
			// Future: send progress via websocket or SSE
		}

		b3, s256, err := hash.DualHashWithProgress(pr, progressCallback)
		hashResult <- struct {
			b3Hash   string
			s256Hash string
			err      error
		}{b3, s256, err}
	}()

	// Copy from source to both dest and hash pipe
	written, copyErr := io.Copy(mw, src)
	pw.Close()

	if copyErr != nil {
		return "", "", 0, copyErr
	}

	// Get hash results
	result := <-hashResult
	if result.err != nil {
		return "", "", 0, result.err
	}

	return result.b3Hash, result.s256Hash, written, nil
}
