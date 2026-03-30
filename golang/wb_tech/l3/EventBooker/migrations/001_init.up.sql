CREATE TABLE IF NOT EXISTS events (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	starts_at TIMESTAMPTZ NOT NULL,
	capacity INTEGER NOT NULL CHECK (capacity > 0),
	requires_confirmation BOOLEAN NOT NULL,
	booking_ttl_minutes INTEGER NOT NULL CHECK (booking_ttl_minutes >= 0),
	created_at TIMESTAMPTZ NOT NULL,
	updated_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS bookings (
	id TEXT PRIMARY KEY,
	event_id TEXT NOT NULL REFERENCES events(id) ON DELETE CASCADE,
	customer_name TEXT NOT NULL,
	customer_email TEXT NOT NULL DEFAULT '',
	status TEXT NOT NULL,
	expires_at TIMESTAMPTZ NULL,
	confirmed_at TIMESTAMPTZ NULL,
	created_at TIMESTAMPTZ NOT NULL,
	updated_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_bookings_event_status ON bookings(event_id, status);
CREATE INDEX IF NOT EXISTS idx_bookings_expires_at ON bookings(expires_at);

