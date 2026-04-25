DROP TRIGGER IF EXISTS trg_items_audit ON items;
DROP TRIGGER IF EXISTS trg_items_set_updated_at ON items;

DROP FUNCTION IF EXISTS log_item_changes();
DROP FUNCTION IF EXISTS set_items_updated_at();

DROP TABLE IF EXISTS item_history;
DROP TABLE IF EXISTS items;
