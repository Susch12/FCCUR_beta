-- Create packages table
CREATE TABLE IF NOT EXISTS packages (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL,
  version TEXT NOT NULL,
  description TEXT,
  category TEXT NOT NULL,
  content_type TEXT NOT NULL DEFAULT 'tool',
  course_name TEXT,
  file_path TEXT NOT NULL UNIQUE,
  file_size INTEGER NOT NULL,
  blake3_hash TEXT NOT NULL,
  sha256_hash TEXT NOT NULL,
  download_url TEXT,
  platform TEXT,
  thumbnail_path TEXT,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for packages
CREATE INDEX IF NOT EXISTS idx_packages_category ON packages(category);
CREATE INDEX IF NOT EXISTS idx_packages_platform ON packages(platform);
CREATE INDEX IF NOT EXISTS idx_packages_content_type ON packages(content_type);
CREATE INDEX IF NOT EXISTS idx_packages_course_name ON packages(course_name);
