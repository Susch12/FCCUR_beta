-- Drop indexes
DROP INDEX IF EXISTS idx_packages_course_name;
DROP INDEX IF EXISTS idx_packages_content_type;
DROP INDEX IF EXISTS idx_packages_platform;
DROP INDEX IF EXISTS idx_packages_category;

-- Drop packages table
DROP TABLE IF EXISTS packages;
