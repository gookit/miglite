-- Migrate:UP
ALTER TABLE users ADD COLUMN ext TEXT DEFAULT '';

-- Migrate:DOWN
-- For demonstration purposes only, as SQLite doesn't support DROP COLUMN directly