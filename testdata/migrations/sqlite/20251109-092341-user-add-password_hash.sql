-- Migrate:UP --
ALTER TABLE users
    ADD COLUMN password_hash TEXT;

-- Migrate:DOWN --
-- ALTER TABLE users DROP COLUMN password_hash;

