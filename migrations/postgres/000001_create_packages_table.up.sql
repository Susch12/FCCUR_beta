-- Enable extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create packages table
CREATE TABLE IF NOT EXISTS packages (
  id BIGSERIAL PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  version VARCHAR(100) NOT NULL,
  description TEXT,
  category VARCHAR(100) NOT NULL,
  content_type VARCHAR(50) NOT NULL DEFAULT 'tool',
  course_name VARCHAR(255),
  file_path VARCHAR(500) NOT NULL UNIQUE,
  file_size BIGINT NOT NULL,
  blake3_hash VARCHAR(64) NOT NULL,
  sha256_hash VARCHAR(64) NOT NULL,
  download_url VARCHAR(500),
  platform VARCHAR(100),
  thumbnail_path VARCHAR(500),
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for packages
CREATE INDEX IF NOT EXISTS idx_packages_category ON packages(category);
CREATE INDEX IF NOT EXISTS idx_packages_platform ON packages(platform);
CREATE INDEX IF NOT EXISTS idx_packages_content_type ON packages(content_type);
CREATE INDEX IF NOT EXISTS idx_packages_course_name ON packages(course_name);
CREATE INDEX IF NOT EXISTS idx_packages_blake3 ON packages(blake3_hash);
CREATE INDEX IF NOT EXISTS idx_packages_sha256 ON packages(sha256_hash);

-- Create trigger function for updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
   NEW.updated_at = CURRENT_TIMESTAMP;
   RETURN NEW;
END;
$$ language 'plpgsql';

-- Create trigger for packages
CREATE TRIGGER update_packages_updated_at
  BEFORE UPDATE ON packages
  FOR EACH ROW
  EXECUTE FUNCTION update_updated_at_column();
