CREATE TABLE IF NOT EXISTS comments (
    id VARCHAR(255) PRIMARY KEY,
    parent_id VARCHAR(255),
    text TEXT NOT NULL,
    author VARCHAR(255) NOT NULL,
    date TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (parent_id) REFERENCES comments(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_comments_parent_id ON comments(parent_id);
CREATE INDEX IF NOT EXISTS idx_comments_date ON comments(date);
CREATE INDEX IF NOT EXISTS idx_comments_text_search ON comments USING gin(to_tsvector('russian', text));
