-- Drop indexes
DROP INDEX IF EXISTS idx_downloads_date;
DROP INDEX IF EXISTS idx_downloads_package;

-- Drop downloads table
DROP TABLE IF EXISTS downloads;
