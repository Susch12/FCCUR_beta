-- Drop trigger
DROP TRIGGER IF EXISTS update_packages_updated_at ON packages;

-- Drop indexes
DROP INDEX IF EXISTS idx_packages_sha256;
DROP INDEX IF EXISTS idx_packages_blake3;
DROP INDEX IF EXISTS idx_packages_course_name;
DROP INDEX IF EXISTS idx_packages_content_type;
DROP INDEX IF EXISTS idx_packages_platform;
DROP INDEX IF EXISTS idx_packages_category;

-- Drop packages table
DROP TABLE IF EXISTS packages;

-- Note: We don't drop the trigger function here as it might be used by other tables
-- DROP FUNCTION IF EXISTS update_updated_at_column();
