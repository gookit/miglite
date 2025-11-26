-- Migrate:UP
ALTER TABLE users ADD COLUMN age INTEGER DEFAULT 0;
ALTER TABLE users ADD COLUMN updated_at DATETIME DEFAULT CURRENT_TIMESTAMP;

-- Migrate:DOWN
-- SQLite 不支持直接删除列，但在实际应用中可以通过创建新表、复制数据、删除旧表、重命名新表的方式实现
-- 这里仅作示例，实际使用时需要更复杂的操作
-- For demonstration purposes only, as SQLite doesn't support DROP COLUMN directly