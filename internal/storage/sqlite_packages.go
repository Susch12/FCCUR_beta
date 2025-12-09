package storage

import (
	"database/sql"
	"time"

	"github.com/jesus/FCCUR/internal/models"
)

// CreatePackage inserts a new package
func (s *SQLiteDB) CreatePackage(pkg *models.Package) (int64, error) {
	query := `
		INSERT INTO packages (name, version, description, category, content_type, course_name,
			file_path, file_size, blake3_hash, sha256_hash, download_url, platform, thumbnail_path)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := s.db.Exec(query,
		pkg.Name, pkg.Version, pkg.Description, pkg.Category, pkg.ContentType, pkg.CourseName,
		pkg.FilePath, pkg.FileSize, pkg.BLAKE3Hash, pkg.SHA256Hash,
		pkg.DownloadURL, pkg.Platform, pkg.ThumbnailPath,
	)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

// GetPackage retrieves a package by ID
func (s *SQLiteDB) GetPackage(id int64) (*models.Package, error) {
	query := `SELECT id, name, version, description, category, content_type, course_name,
		file_path, file_size, blake3_hash, sha256_hash, download_url, platform, thumbnail_path,
		created_at, updated_at FROM packages WHERE id = ?`

	pkg := &models.Package{}
	err := s.db.QueryRow(query, id).Scan(
		&pkg.ID, &pkg.Name, &pkg.Version, &pkg.Description,
		&pkg.Category, &pkg.ContentType, &pkg.CourseName, &pkg.FilePath, &pkg.FileSize,
		&pkg.BLAKE3Hash, &pkg.SHA256Hash, &pkg.DownloadURL,
		&pkg.Platform, &pkg.ThumbnailPath, &pkg.CreatedAt, &pkg.UpdatedAt,
	)

	return pkg, err
}

// FindPackageByHash retrieves a package by BLAKE3 hash
func (s *SQLiteDB) FindPackageByHash(blake3Hash string) (*models.Package, error) {
	query := `SELECT id, name, version, description, category, content_type, course_name,
		file_path, file_size, blake3_hash, sha256_hash, download_url, platform, thumbnail_path,
		created_at, updated_at FROM packages WHERE blake3_hash = ?`

	pkg := &models.Package{}
	err := s.db.QueryRow(query, blake3Hash).Scan(
		&pkg.ID, &pkg.Name, &pkg.Version, &pkg.Description,
		&pkg.Category, &pkg.ContentType, &pkg.CourseName, &pkg.FilePath, &pkg.FileSize,
		&pkg.BLAKE3Hash, &pkg.SHA256Hash, &pkg.DownloadURL,
		&pkg.Platform, &pkg.ThumbnailPath, &pkg.CreatedAt, &pkg.UpdatedAt,
	)

	return pkg, err
}

// GetPackages retrieves all packages
func (s *SQLiteDB) GetPackages() ([]*models.Package, error) {
	query := `SELECT id, name, version, description, category, content_type, course_name,
		file_path, file_size, blake3_hash, sha256_hash, download_url, platform, thumbnail_path,
		created_at, updated_at FROM packages ORDER BY created_at DESC`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	packages := []*models.Package{}
	for rows.Next() {
		pkg := &models.Package{}
		err := rows.Scan(
			&pkg.ID, &pkg.Name, &pkg.Version, &pkg.Description,
			&pkg.Category, &pkg.ContentType, &pkg.CourseName, &pkg.FilePath, &pkg.FileSize,
			&pkg.BLAKE3Hash, &pkg.SHA256Hash, &pkg.DownloadURL,
			&pkg.Platform, &pkg.ThumbnailPath, &pkg.CreatedAt, &pkg.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		packages = append(packages, pkg)
	}

	return packages, nil
}

// ListPackages retrieves packages with filters and pagination
func (s *SQLiteDB) ListPackages(limit, offset int, category, platform, contentType, courseName string) ([]*models.Package, error) {
	query := `SELECT id, name, version, description, category, content_type, course_name,
		file_path, file_size, blake3_hash, sha256_hash, download_url, platform, thumbnail_path,
		created_at, updated_at FROM packages WHERE 1=1`
	args := []interface{}{}

	if category != "" {
		query += ` AND category = ?`
		args = append(args, category)
	}
	if platform != "" {
		query += ` AND platform = ?`
		args = append(args, platform)
	}
	if contentType != "" {
		query += ` AND content_type = ?`
		args = append(args, contentType)
	}
	if courseName != "" {
		query += ` AND course_name = ?`
		args = append(args, courseName)
	}

	query += ` ORDER BY created_at DESC`

	if limit > 0 {
		query += ` LIMIT ?`
		args = append(args, limit)
	}
	if offset > 0 {
		query += ` OFFSET ?`
		args = append(args, offset)
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	packages := []*models.Package{}
	for rows.Next() {
		pkg := &models.Package{}
		err := rows.Scan(
			&pkg.ID, &pkg.Name, &pkg.Version, &pkg.Description,
			&pkg.Category, &pkg.ContentType, &pkg.CourseName, &pkg.FilePath, &pkg.FileSize,
			&pkg.BLAKE3Hash, &pkg.SHA256Hash, &pkg.DownloadURL,
			&pkg.Platform, &pkg.ThumbnailPath, &pkg.CreatedAt, &pkg.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		packages = append(packages, pkg)
	}

	return packages, nil
}

// RecordDownload logs a download event
func (s *SQLiteDB) RecordDownload(packageID int64, ip, userAgent string) error {
	query := `
		INSERT INTO downloads (package_id, ip_address, user_agent)
		VALUES (?, ?, ?)
	`
	_, err := s.db.Exec(query, packageID, ip, userAgent)
	return err
}

// DeletePackage removes a package and its download records
func (s *SQLiteDB) DeletePackage(id int64) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Delete download records first (foreign key constraint)
	_, err = tx.Exec("DELETE FROM downloads WHERE package_id = ?", id)
	if err != nil {
		return err
	}

	// Delete package record
	result, err := tx.Exec("DELETE FROM packages WHERE id = ?", id)
	if err != nil {
		return err
	}

	// Check if package existed
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return tx.Commit()
}

// GetStats retrieves download statistics
func (s *SQLiteDB) GetStats() ([]*models.DownloadStats, error) {
	query := `
		SELECT
			p.id,
			p.name,
			COUNT(d.id) as total,
			COALESCE(MAX(d.downloaded_at), '') as last_download
		FROM packages p
		LEFT JOIN downloads d ON p.id = d.package_id
		GROUP BY p.id, p.name
		ORDER BY total DESC
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats := []*models.DownloadStats{}
	for rows.Next() {
		s := &models.DownloadStats{}
		var lastDownloadStr string

		err := rows.Scan(&s.PackageID, &s.PackageName, &s.TotalDownloads, &lastDownloadStr)
		if err != nil {
			return nil, err
		}

		// Parse time if not empty
		if lastDownloadStr != "" {
			// Try to parse the time
			if t, err := parseTime(lastDownloadStr); err == nil {
				s.LastDownload = t
			}
		}

		stats = append(stats, s)
	}

	return stats, nil
}

// parseTime parses SQLite datetime string
func parseTime(s string) (time.Time, error) {
	// SQLite CURRENT_TIMESTAMP format: "2006-01-02 15:04:05"
	return time.Parse("2006-01-02 15:04:05", s)
}

// GetDownloadCount gets the download count for a package
func (s *SQLiteDB) GetDownloadCount(packageID int64) (int64, error) {
	var count int64
	err := s.db.QueryRow(`SELECT COUNT(*) FROM downloads WHERE package_id = ?`, packageID).Scan(&count)
	return count, err
}

// GetTotalDownloads gets the total number of downloads
func (s *SQLiteDB) GetTotalDownloads() (int64, error) {
	var count int64
	err := s.db.QueryRow(`SELECT COUNT(*) FROM downloads`).Scan(&count)
	return count, err
}

// GetPackageCount gets the total number of packages
func (s *SQLiteDB) GetPackageCount() (int64, error) {
	var count int64
	err := s.db.QueryRow(`SELECT COUNT(*) FROM packages`).Scan(&count)
	return count, err
}

// GetTotalSize gets the total size of all packages
func (s *SQLiteDB) GetTotalSize() (int64, error) {
	var size int64
	err := s.db.QueryRow(`SELECT COALESCE(SUM(file_size), 0) FROM packages`).Scan(&size)
	return size, err
}

// GetRecentPackages gets the most recent packages
func (s *SQLiteDB) GetRecentPackages(limit int) ([]*models.Package, error) {
	query := `SELECT id, name, version, description, category, content_type, course_name,
		file_path, file_size, blake3_hash, sha256_hash, download_url, platform, thumbnail_path,
		created_at, updated_at FROM packages ORDER BY created_at DESC LIMIT ?`

	rows, err := s.db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	packages := []*models.Package{}
	for rows.Next() {
		pkg := &models.Package{}
		err := rows.Scan(
			&pkg.ID, &pkg.Name, &pkg.Version, &pkg.Description,
			&pkg.Category, &pkg.ContentType, &pkg.CourseName, &pkg.FilePath, &pkg.FileSize,
			&pkg.BLAKE3Hash, &pkg.SHA256Hash, &pkg.DownloadURL,
			&pkg.Platform, &pkg.ThumbnailPath, &pkg.CreatedAt, &pkg.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		packages = append(packages, pkg)
	}

	return packages, nil
}
