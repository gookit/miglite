--
-- name: add-age-index
-- author: inhere
-- created_at: 20251106-215850
--

-- Migrate:UP
CREATE INDEX idx_users_age ON users(age);

-- Migrate:DOWN
DROP INDEX idx_users_age;