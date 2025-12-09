package storage

import (
	"database/sql"

	"github.com/jesus/FCCUR/internal/models"
)

// CreatePackage creates a new package record
func (p *PostgresDB) CreatePackage(pkg *models.Package) (int64, error) {
	ctx, cancel := p.getContext()
	defer cancel()

	var id int64
	err := p.pool.QueryRow(ctx, `
		INSERT INTO packages (
			name, version, description, category, content_type, course_name,
			file_path, file_size, blake3_hash, sha256_hash, download_url, platform, thumbnail_path
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id
	`, pkg.Name, pkg.Version, pkg.Description, pkg.Category, pkg.ContentType,
		pkg.CourseName, pkg.FilePath, pkg.FileSize, pkg.BLAKE3Hash, pkg.SHA256Hash,
		pkg.DownloadURL, pkg.Platform, pkg.ThumbnailPath).Scan(&id)

	return id, err
}

// GetPackage retrieves a package by ID
func (p *PostgresDB) GetPackage(id int64) (*models.Package, error) {
	ctx, cancel := p.getContext()
	defer cancel()

	pkg := &models.Package{}
	err := p.pool.QueryRow(ctx, `
		SELECT id, name, version, description, category, content_type, course_name,
		       file_path, file_size, blake3_hash, sha256_hash, download_url, platform,
		       thumbnail_path, created_at, updated_at
		FROM packages WHERE id = $1
	`, id).Scan(
		&pkg.ID, &pkg.Name, &pkg.Version, &pkg.Description, &pkg.Category,
		&pkg.ContentType, &pkg.CourseName, &pkg.FilePath, &pkg.FileSize,
		&pkg.BLAKE3Hash, &pkg.SHA256Hash, &pkg.DownloadURL, &pkg.Platform,
		&pkg.ThumbnailPath, &pkg.CreatedAt, &pkg.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrPackageNotFound
	}
	return pkg, err
}

// GetPackages retrieves all packages
func (p *PostgresDB) GetPackages() ([]*models.Package, error) {
	ctx, cancel := p.getContext()
	defer cancel()

	rows, err := p.pool.Query(ctx, `
		SELECT id, name, version, description, category, content_type, course_name,
		       file_path, file_size, blake3_hash, sha256_hash, download_url, platform,
		       thumbnail_path, created_at, updated_at
		FROM packages
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var packages []*models.Package
	for rows.Next() {
		pkg := &models.Package{}
		err := rows.Scan(
			&pkg.ID, &pkg.Name, &pkg.Version, &pkg.Description, &pkg.Category,
			&pkg.ContentType, &pkg.CourseName, &pkg.FilePath, &pkg.FileSize,
			&pkg.BLAKE3Hash, &pkg.SHA256Hash, &pkg.DownloadURL, &pkg.Platform,
			&pkg.ThumbnailPath, &pkg.CreatedAt, &pkg.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		packages = append(packages, pkg)
	}

	return packages, rows.Err()
}

// ListPackages retrieves packages with filters and pagination
func (p *PostgresDB) ListPackages(limit, offset int, category, platform, contentType, courseName string) ([]*models.Package, error) {
	ctx, cancel := p.getContext()
	defer cancel()

	query := `
		SELECT id, name, version, description, category, content_type, course_name,
		       file_path, file_size, blake3_hash, sha256_hash, download_url, platform,
		       thumbnail_path, created_at, updated_at
		FROM packages
		WHERE 1=1
	`
	args := []interface{}{}
	argPos := 1

	if category != "" {
		query += ` AND category = $` + string(rune(argPos+'0'))
		args = append(args, category)
		argPos++
	}
	if platform != "" {
		query += ` AND platform = $` + string(rune(argPos+'0'))
		args = append(args, platform)
		argPos++
	}
	if contentType != "" {
		query += ` AND content_type = $` + string(rune(argPos+'0'))
		args = append(args, contentType)
		argPos++
	}
	if courseName != "" {
		query += ` AND course_name = $` + string(rune(argPos+'0'))
		args = append(args, courseName)
		argPos++
	}

	query += ` ORDER BY created_at DESC`

	if limit > 0 {
		query += ` LIMIT $` + string(rune(argPos+'0'))
		args = append(args, limit)
		argPos++
	}
	if offset > 0 {
		query += ` OFFSET $` + string(rune(argPos+'0'))
		args = append(args, offset)
	}

	rows, err := p.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var packages []*models.Package
	for rows.Next() {
		pkg := &models.Package{}
		err := rows.Scan(
			&pkg.ID, &pkg.Name, &pkg.Version, &pkg.Description, &pkg.Category,
			&pkg.ContentType, &pkg.CourseName, &pkg.FilePath, &pkg.FileSize,
			&pkg.BLAKE3Hash, &pkg.SHA256Hash, &pkg.DownloadURL, &pkg.Platform,
			&pkg.ThumbnailPath, &pkg.CreatedAt, &pkg.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		packages = append(packages, pkg)
	}

	return packages, rows.Err()
}

// DeletePackage deletes a package
func (p *PostgresDB) DeletePackage(id int64) error {
	ctx, cancel := p.getContext()
	defer cancel()

	_, err := p.pool.Exec(ctx, `DELETE FROM packages WHERE id = $1`, id)
	return err
}

// FindPackageByHash finds a package by BLAKE3 or SHA256 hash
func (p *PostgresDB) FindPackageByHash(hash string) (*models.Package, error) {
	ctx, cancel := p.getContext()
	defer cancel()

	pkg := &models.Package{}
	err := p.pool.QueryRow(ctx, `
		SELECT id, name, version, description, category, content_type, course_name,
		       file_path, file_size, blake3_hash, sha256_hash, download_url, platform,
		       thumbnail_path, created_at, updated_at
		FROM packages
		WHERE blake3_hash = $1 OR sha256_hash = $1
		LIMIT 1
	`, hash).Scan(
		&pkg.ID, &pkg.Name, &pkg.Version, &pkg.Description, &pkg.Category,
		&pkg.ContentType, &pkg.CourseName, &pkg.FilePath, &pkg.FileSize,
		&pkg.BLAKE3Hash, &pkg.SHA256Hash, &pkg.DownloadURL, &pkg.Platform,
		&pkg.ThumbnailPath, &pkg.CreatedAt, &pkg.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil // No duplicate found
	}
	return pkg, err
}

// RecordDownload records a package download
func (p *PostgresDB) RecordDownload(packageID int64, ipAddress, userAgent string) error {
	ctx, cancel := p.getContext()
	defer cancel()

	_, err := p.pool.Exec(ctx, `
		INSERT INTO downloads (package_id, ip_address, user_agent)
		VALUES ($1, $2, $3)
	`, packageID, ipAddress, userAgent)

	return err
}

// GetDownloadCount gets the download count for a package
func (p *PostgresDB) GetDownloadCount(packageID int64) (int64, error) {
	ctx, cancel := p.getContext()
	defer cancel()

	var count int64
	err := p.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM downloads WHERE package_id = $1
	`, packageID).Scan(&count)

	return count, err
}

// GetTotalDownloads gets the total number of downloads
func (p *PostgresDB) GetTotalDownloads() (int64, error) {
	ctx, cancel := p.getContext()
	defer cancel()

	var count int64
	err := p.pool.QueryRow(ctx, `SELECT COUNT(*) FROM downloads`).Scan(&count)
	return count, err
}

// GetPackageCount gets the total number of packages
func (p *PostgresDB) GetPackageCount() (int64, error) {
	ctx, cancel := p.getContext()
	defer cancel()

	var count int64
	err := p.pool.QueryRow(ctx, `SELECT COUNT(*) FROM packages`).Scan(&count)
	return count, err
}

// GetTotalSize gets the total size of all packages
func (p *PostgresDB) GetTotalSize() (int64, error) {
	ctx, cancel := p.getContext()
	defer cancel()

	var size int64
	err := p.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(file_size), 0) FROM packages
	`).Scan(&size)
	return size, err
}

// GetRecentPackages gets the most recent packages
func (p *PostgresDB) GetRecentPackages(limit int) ([]*models.Package, error) {
	ctx, cancel := p.getContext()
	defer cancel()

	rows, err := p.pool.Query(ctx, `
		SELECT id, name, version, description, category, content_type, course_name,
		       file_path, file_size, blake3_hash, sha256_hash, download_url, platform,
		       thumbnail_path, created_at, updated_at
		FROM packages
		ORDER BY created_at DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var packages []*models.Package
	for rows.Next() {
		pkg := &models.Package{}
		err := rows.Scan(
			&pkg.ID, &pkg.Name, &pkg.Version, &pkg.Description, &pkg.Category,
			&pkg.ContentType, &pkg.CourseName, &pkg.FilePath, &pkg.FileSize,
			&pkg.BLAKE3Hash, &pkg.SHA256Hash, &pkg.DownloadURL, &pkg.Platform,
			&pkg.ThumbnailPath, &pkg.CreatedAt, &pkg.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		packages = append(packages, pkg)
	}

	return packages, rows.Err()
}

// GetStats retrieves download statistics
func (p *PostgresDB) GetStats() ([]*models.DownloadStats, error) {
	ctx, cancel := p.getContext()
	defer cancel()

	rows, err := p.pool.Query(ctx, `
		SELECT
			p.id,
			p.name,
			COUNT(d.id) as total,
			COALESCE(MAX(d.downloaded_at), '1970-01-01 00:00:00'::timestamp) as last_download
		FROM packages p
		LEFT JOIN downloads d ON p.id = d.package_id
		GROUP BY p.id, p.name
		ORDER BY total DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []*models.DownloadStats
	for rows.Next() {
		s := &models.DownloadStats{}
		err := rows.Scan(&s.PackageID, &s.PackageName, &s.TotalDownloads, &s.LastDownload)
		if err != nil {
			return nil, err
		}
		stats = append(stats, s)
	}

	return stats, rows.Err()
}
