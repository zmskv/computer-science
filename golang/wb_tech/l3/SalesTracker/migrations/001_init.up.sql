CREATE TABLE IF NOT EXISTS items (
	id TEXT PRIMARY KEY,
	item_type TEXT NOT NULL CHECK (item_type IN ('income', 'expense')),
	amount NUMERIC(14, 2) NOT NULL CHECK (amount >= 0),
	category TEXT NOT NULL,
	description TEXT NOT NULL DEFAULT '',
	occurred_at TIMESTAMPTZ NOT NULL,
	created_at TIMESTAMPTZ NOT NULL,
	updated_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_items_occurred_at ON items(occurred_at DESC);
CREATE INDEX IF NOT EXISTS idx_items_type ON items(item_type);
CREATE INDEX IF NOT EXISTS idx_items_category ON items(category);
CREATE INDEX IF NOT EXISTS idx_items_type_occurred_at ON items(item_type, occurred_at DESC);
