CREATE TABLE IF NOT EXISTS items (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    sku TEXT NOT NULL UNIQUE,
    quantity INTEGER NOT NULL CHECK (quantity >= 0),
    location TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS item_history (
    id BIGSERIAL PRIMARY KEY,
    item_id TEXT NOT NULL,
    action TEXT NOT NULL CHECK (action IN ('insert', 'update', 'delete')),
    changed_by TEXT NOT NULL,
    changed_role TEXT NOT NULL,
    changed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    old_data JSONB,
    new_data JSONB
);

CREATE INDEX IF NOT EXISTS idx_items_sku ON items (sku);
CREATE INDEX IF NOT EXISTS idx_items_updated_at ON items (updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_item_history_item_id ON item_history (item_id);
CREATE INDEX IF NOT EXISTS idx_item_history_changed_at ON item_history (changed_at DESC);
CREATE INDEX IF NOT EXISTS idx_item_history_changed_by ON item_history (changed_by);

CREATE OR REPLACE FUNCTION set_items_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION log_item_changes()
RETURNS TRIGGER AS $$
DECLARE
    actor_name TEXT := COALESCE(NULLIF(current_setting('app.current_user', true), ''), 'system');
    actor_role TEXT := COALESCE(NULLIF(current_setting('app.current_role', true), ''), 'system');
BEGIN
    IF TG_OP = 'INSERT' THEN
        INSERT INTO item_history (item_id, action, changed_by, changed_role, changed_at, old_data, new_data)
        VALUES (NEW.id, 'insert', actor_name, actor_role, NOW(), NULL, to_jsonb(NEW));
        RETURN NEW;
    ELSIF TG_OP = 'UPDATE' THEN
        INSERT INTO item_history (item_id, action, changed_by, changed_role, changed_at, old_data, new_data)
        VALUES (NEW.id, 'update', actor_name, actor_role, NOW(), to_jsonb(OLD), to_jsonb(NEW));
        RETURN NEW;
    ELSIF TG_OP = 'DELETE' THEN
        INSERT INTO item_history (item_id, action, changed_by, changed_role, changed_at, old_data, new_data)
        VALUES (OLD.id, 'delete', actor_name, actor_role, NOW(), to_jsonb(OLD), NULL);
        RETURN OLD;
    END IF;

    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_items_set_updated_at ON items;
CREATE TRIGGER trg_items_set_updated_at
BEFORE UPDATE ON items
FOR EACH ROW
EXECUTE FUNCTION set_items_updated_at();

DROP TRIGGER IF EXISTS trg_items_audit ON items;
CREATE TRIGGER trg_items_audit
AFTER INSERT OR UPDATE OR DELETE ON items
FOR EACH ROW
EXECUTE FUNCTION log_item_changes();
