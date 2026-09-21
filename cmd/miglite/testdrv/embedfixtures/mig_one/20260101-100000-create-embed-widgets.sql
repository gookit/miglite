-- Migrate:UP
CREATE TABLE embed_widgets (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL
);
-- Migrate:DOWN
DROP TABLE embed_widgets;
