DROP INDEX IF EXISTS idx_users_email;
DROP INDEX IF EXISTS idx_bookings_user_id;
ALTER TABLE bookings DROP COLUMN IF EXISTS user_id;
DROP TABLE IF EXISTS users;
