-- Migrate:UP
CREATE TABLE embed_skipped (id INTEGER PRIMARY KEY);
-- Migrate:DOWN
DROP TABLE embed_skipped;
