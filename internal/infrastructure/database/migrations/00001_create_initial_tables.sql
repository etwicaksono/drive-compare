-- +goose Up
CREATE TABLE comparisons (
    id SERIAL PRIMARY KEY,
    local_path TEXT NOT NULL,
    drive_path TEXT NOT NULL,
    started_at TIMESTAMP NOT NULL DEFAULT NOW(),
    finished_at TIMESTAMP,
    total_local_only INT NOT NULL DEFAULT 0,
    total_drive_only INT NOT NULL DEFAULT 0,
    total_both INT NOT NULL DEFAULT 0
);

CREATE TABLE comparison_files (
    id SERIAL PRIMARY KEY,
    comparison_id INT NOT NULL REFERENCES comparisons(id) ON DELETE CASCADE,
    filename TEXT NOT NULL,
    extension TEXT NOT NULL DEFAULT '',
    relative_path TEXT NOT NULL,
    size BIGINT NOT NULL DEFAULT 0,
    status TEXT NOT NULL CHECK (status IN ('local_only', 'drive_only', 'both'))
);

CREATE INDEX idx_comparison_files_comparison_id ON comparison_files(comparison_id);
CREATE INDEX idx_comparison_files_status ON comparison_files(status);

CREATE TABLE ignore_patterns (
    id SERIAL PRIMARY KEY,
    pattern TEXT NOT NULL UNIQUE
);

-- +goose Down
DROP TABLE IF EXISTS comparison_files;
DROP TABLE IF EXISTS comparisons;
DROP TABLE IF EXISTS ignore_patterns;
