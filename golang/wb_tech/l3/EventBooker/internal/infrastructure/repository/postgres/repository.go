package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/wb-go/wbf/dbpg"
	"github.com/zmskv/computer-science/golang/wb_tech/l3/EventBooker/internal/application"
	"github.com/zmskv/computer-science/golang/wb_tech/l3/EventBooker/internal/domain/entity"
)

type Repository struct {
	db *dbpg.DB
}

func NewRepository(ctx context.Context, masterDSN string, slaveDSNs []string, opts *dbpg.Options) (*Repository, error) {
	db, err := dbpg.New(masterDSN, slaveDSNs, opts)
	if err != nil {
		return nil, fmt.Errorf("create postgres connection: %w", err)
	}

	if err := db.Master.PingContext(ctx); err != nil {
		_ = db.Master.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return &Repository{db: db}, nil
}

func (r *Repository) Close() {
	if r.db == nil {
		return
	}

	if r.db.Master != nil {
		_ = r.db.Master.Close()
	}

	for _, slave := range r.db.Slaves {
		if slave != nil {
			_ = slave.Close()
		}
	}
}

func (r *Repository) CreateUser(ctx context.Context, user entity.User) error {
	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO users (id, name, email, telegram_chat_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		user.ID,
		user.Name,
		user.Email,
		user.TelegramChatID,
		user.CreatedAt.UTC(),
		user.UpdatedAt.UTC(),
	)
	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}

	return nil
}

func (r *Repository) ListUsers(ctx context.Context) ([]entity.User, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT id, name, email, telegram_chat_id, created_at, updated_at
		FROM users
		ORDER BY created_at ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	users := make([]entity.User, 0)
	for rows.Next() {
		user, err := scanUser(rows)
		if err != nil {
			return nil, err
		}

		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate users: %w", err)
	}

	return users, nil
}

func (r *Repository) GetUser(ctx context.Context, userID string) (entity.User, error) {
	row := r.db.QueryRowContext(
		ctx,
		`SELECT id, name, email, telegram_chat_id, created_at, updated_at
		FROM users
		WHERE id = $1`,
		userID,
	)

	user, err := scanUser(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.User{}, application.ErrUserNotFound
		}
		return entity.User{}, err
	}

	return user, nil
}

func (r *Repository) CreateEvent(ctx context.Context, event entity.Event) error {
	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO events (
			id, name, starts_at, capacity, requires_confirmation, booking_ttl_minutes, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		event.ID,
		event.Name,
		event.StartsAt.UTC(),
		event.Capacity,
		event.RequiresConfirmation,
		event.BookingTTLMinutes,
		event.CreatedAt.UTC(),
		event.UpdatedAt.UTC(),
	)
	if err != nil {
		return fmt.Errorf("create event: %w", err)
	}

	return nil
}

func (r *Repository) ListEvents(ctx context.Context) ([]entity.EventDetails, error) {
	now := time.Now().UTC()

	rows, err := r.db.QueryContext(
		ctx,
		`SELECT
			e.id,
			e.name,
			e.starts_at,
			e.capacity,
			e.requires_confirmation,
			e.booking_ttl_minutes,
			e.created_at,
			e.updated_at,
			COALESCE(COUNT(*) FILTER (WHERE b.status = 'pending' AND b.expires_at > $1), 0) AS pending_bookings,
			COALESCE(COUNT(*) FILTER (WHERE b.status = 'confirmed'), 0) AS confirmed_bookings
		FROM events e
		LEFT JOIN bookings b ON b.event_id = e.id
		GROUP BY e.id
		ORDER BY e.starts_at ASC, e.created_at ASC`,
		now,
	)
	if err != nil {
		return nil, fmt.Errorf("list events: %w", err)
	}
	defer rows.Close()

	events := make([]entity.EventDetails, 0)
	byID := make(map[string]int)

	for rows.Next() {
		eventDetails, err := scanEventDetails(rows)
		if err != nil {
			return nil, err
		}

		eventDetails.Bookings = []entity.Booking{}
		eventDetails.AvailableSeats = max(0, eventDetails.Capacity-eventDetails.PendingBookings-eventDetails.ConfirmedBookings)

		byID[eventDetails.ID] = len(events)
		events = append(events, eventDetails)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate events: %w", err)
	}

	if len(events) == 0 {
		return events, nil
	}

	bookingRows, err := r.db.QueryContext(
		ctx,
		`SELECT
			b.id,
			b.event_id,
			b.user_id,
			b.customer_name,
			b.customer_email,
			b.status,
			b.expires_at,
			b.confirmed_at,
			b.created_at,
			b.updated_at,
			u.id,
			u.name,
			u.email,
			u.telegram_chat_id,
			u.created_at,
			u.updated_at
		FROM bookings b
		LEFT JOIN users u ON u.id = b.user_id
		WHERE b.status = 'confirmed' OR (b.status = 'pending' AND b.expires_at > $1)
		ORDER BY b.created_at ASC`,
		now,
	)
	if err != nil {
		return nil, fmt.Errorf("list active bookings: %w", err)
	}
	defer bookingRows.Close()

	for bookingRows.Next() {
		booking, err := scanBookingWithUser(bookingRows)
		if err != nil {
			return nil, err
		}

		index, ok := byID[booking.EventID]
		if !ok {
			continue
		}

		events[index].Bookings = append(events[index].Bookings, booking)
	}

	if err := bookingRows.Err(); err != nil {
		return nil, fmt.Errorf("iterate active bookings: %w", err)
	}

	return events, nil
}

func (r *Repository) GetEvent(ctx context.Context, eventID string) (entity.EventDetails, error) {
	now := time.Now().UTC()

	row := r.db.QueryRowContext(
		ctx,
		`SELECT
			e.id,
			e.name,
			e.starts_at,
			e.capacity,
			e.requires_confirmation,
			e.booking_ttl_minutes,
			e.created_at,
			e.updated_at,
			COALESCE(COUNT(*) FILTER (WHERE b.status = 'pending' AND b.expires_at > $1), 0) AS pending_bookings,
			COALESCE(COUNT(*) FILTER (WHERE b.status = 'confirmed'), 0) AS confirmed_bookings
		FROM events e
		LEFT JOIN bookings b ON b.event_id = e.id
		WHERE e.id = $2
		GROUP BY e.id`,
		now,
		eventID,
	)

	eventDetails, err := scanEventDetails(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.EventDetails{}, application.ErrEventNotFound
		}
		return entity.EventDetails{}, err
	}

	eventDetails.Bookings = []entity.Booking{}
	eventDetails.AvailableSeats = max(0, eventDetails.Capacity-eventDetails.PendingBookings-eventDetails.ConfirmedBookings)

	rows, err := r.db.QueryContext(
		ctx,
		`SELECT
			b.id,
			b.event_id,
			b.user_id,
			b.customer_name,
			b.customer_email,
			b.status,
			b.expires_at,
			b.confirmed_at,
			b.created_at,
			b.updated_at,
			u.id,
			u.name,
			u.email,
			u.telegram_chat_id,
			u.created_at,
			u.updated_at
		FROM bookings b
		LEFT JOIN users u ON u.id = b.user_id
		WHERE b.event_id = $1
		  AND (b.status = 'confirmed' OR (b.status = 'pending' AND b.expires_at > $2))
		ORDER BY b.created_at ASC`,
		eventID,
		now,
	)
	if err != nil {
		return entity.EventDetails{}, fmt.Errorf("list event bookings: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		booking, err := scanBookingWithUser(rows)
		if err != nil {
			return entity.EventDetails{}, err
		}

		eventDetails.Bookings = append(eventDetails.Bookings, booking)
	}

	if err := rows.Err(); err != nil {
		return entity.EventDetails{}, fmt.Errorf("iterate event bookings: %w", err)
	}

	return eventDetails, nil
}

func (r *Repository) CreateBooking(ctx context.Context, booking entity.Booking, now time.Time) (entity.Booking, error) {
	tx, err := r.db.Master.BeginTx(ctx, nil)
	if err != nil {
		return entity.Booking{}, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	if err := expirePendingTx(ctx, tx, now, booking.EventID); err != nil {
		return entity.Booking{}, err
	}

	var capacity int
	if err := tx.QueryRowContext(
		ctx,
		`SELECT capacity
		FROM events
		WHERE id = $1
		FOR UPDATE`,
		booking.EventID,
	).Scan(&capacity); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.Booking{}, application.ErrEventNotFound
		}
		return entity.Booking{}, fmt.Errorf("load event capacity: %w", err)
	}

	var activeBookings int
	if err := tx.QueryRowContext(
		ctx,
		`SELECT COUNT(*)
		FROM bookings
		WHERE event_id = $1
		  AND (status = 'confirmed' OR (status = 'pending' AND expires_at > $2))`,
		booking.EventID,
		now,
	).Scan(&activeBookings); err != nil {
		return entity.Booking{}, fmt.Errorf("count active bookings: %w", err)
	}

	if activeBookings >= capacity {
		return entity.Booking{}, application.ErrNoSeatsAvailable
	}

	_, err = tx.ExecContext(
		ctx,
		`INSERT INTO bookings (
			id, event_id, user_id, customer_name, customer_email, status, expires_at, confirmed_at, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		booking.ID,
		booking.EventID,
		booking.UserID,
		booking.CustomerName,
		booking.CustomerEmail,
		booking.Status,
		booking.ExpiresAt,
		booking.ConfirmedAt,
		booking.CreatedAt.UTC(),
		booking.UpdatedAt.UTC(),
	)
	if err != nil {
		return entity.Booking{}, fmt.Errorf("create booking: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return entity.Booking{}, fmt.Errorf("commit booking transaction: %w", err)
	}

	return booking, nil
}

func (r *Repository) ConfirmBooking(ctx context.Context, eventID, bookingID string, now time.Time) (entity.Booking, error) {
	tx, err := r.db.Master.BeginTx(ctx, nil)
	if err != nil {
		return entity.Booking{}, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	if err := expirePendingTx(ctx, tx, now, eventID); err != nil {
		return entity.Booking{}, err
	}

	var requiresConfirmation bool
	if err := tx.QueryRowContext(
		ctx,
		`SELECT requires_confirmation
		FROM events
		WHERE id = $1
		FOR UPDATE`,
		eventID,
	).Scan(&requiresConfirmation); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.Booking{}, application.ErrEventNotFound
		}
		return entity.Booking{}, fmt.Errorf("load event confirmation mode: %w", err)
	}

	if !requiresConfirmation {
		return entity.Booking{}, application.ErrConfirmationNotRequired
	}

	row := tx.QueryRowContext(
		ctx,
		`SELECT id, event_id, user_id, customer_name, customer_email, status, expires_at, confirmed_at, created_at, updated_at
		FROM bookings
		WHERE id = $1 AND event_id = $2
		FOR UPDATE`,
		bookingID,
		eventID,
	)

	booking, err := scanBooking(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.Booking{}, application.ErrBookingNotFound
		}
		return entity.Booking{}, err
	}

	switch booking.Status {
	case entity.BookingStatusConfirmed:
		return entity.Booking{}, application.ErrBookingAlreadyConfirmed
	case entity.BookingStatusExpired:
		return entity.Booking{}, application.ErrBookingExpired
	case entity.BookingStatusPending:
	default:
		return entity.Booking{}, application.ErrInvalidInput
	}

	if booking.ExpiresAt != nil && !booking.ExpiresAt.After(now) {
		if _, err := tx.ExecContext(
			ctx,
			`UPDATE bookings SET status = $1, updated_at = $2 WHERE id = $3`,
			entity.BookingStatusExpired,
			now,
			bookingID,
		); err != nil {
			return entity.Booking{}, fmt.Errorf("expire booking during confirm: %w", err)
		}

		return entity.Booking{}, application.ErrBookingExpired
	}

	confirmedAt := now.UTC()

	if _, err := tx.ExecContext(
		ctx,
		`UPDATE bookings
		SET status = $1, confirmed_at = $2, updated_at = $3
		WHERE id = $4 AND event_id = $5`,
		entity.BookingStatusConfirmed,
		confirmedAt,
		now,
		bookingID,
		eventID,
	); err != nil {
		return entity.Booking{}, fmt.Errorf("confirm booking: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return entity.Booking{}, fmt.Errorf("commit confirm transaction: %w", err)
	}

	booking.Status = entity.BookingStatusConfirmed
	booking.ConfirmedAt = &confirmedAt
	booking.UpdatedAt = now.UTC()

	return booking, nil
}

func (r *Repository) ExpireBookings(ctx context.Context, now time.Time) ([]entity.ExpiredBookingNotice, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`WITH expired AS (
			UPDATE bookings
			SET status = $1, updated_at = $2
			WHERE status = $3
			  AND expires_at IS NOT NULL
			  AND expires_at <= $2
			RETURNING id, event_id, user_id, customer_name, customer_email, status, expires_at, confirmed_at, created_at, updated_at
		)
		SELECT
			expired.id,
			expired.event_id,
			expired.user_id,
			expired.customer_name,
			expired.customer_email,
			expired.status,
			expired.expires_at,
			expired.confirmed_at,
			expired.created_at,
			expired.updated_at,
			u.id,
			u.name,
			u.email,
			u.telegram_chat_id,
			u.created_at,
			u.updated_at,
			e.id,
			e.name,
			e.starts_at,
			e.capacity,
			e.requires_confirmation,
			e.booking_ttl_minutes,
			e.created_at,
			e.updated_at
		FROM expired
		LEFT JOIN users u ON u.id = expired.user_id
		JOIN events e ON e.id = expired.event_id
		ORDER BY expired.created_at ASC`,
		entity.BookingStatusExpired,
		now,
		entity.BookingStatusPending,
	)
	if err != nil {
		return nil, fmt.Errorf("expire bookings: %w", err)
	}
	defer rows.Close()

	notices := make([]entity.ExpiredBookingNotice, 0)
	for rows.Next() {
		notice, err := scanExpiredBookingNotice(rows)
		if err != nil {
			return nil, err
		}

		notices = append(notices, notice)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate expired bookings: %w", err)
	}

	return notices, nil
}

func expirePendingTx(ctx context.Context, tx *sql.Tx, now time.Time, eventID string) error {
	if _, err := tx.ExecContext(
		ctx,
		`UPDATE bookings
		SET status = $1, updated_at = $2
		WHERE event_id = $3
		  AND status = $4
		  AND expires_at IS NOT NULL
		  AND expires_at <= $5`,
		entity.BookingStatusExpired,
		now,
		eventID,
		entity.BookingStatusPending,
		now,
	); err != nil {
		return fmt.Errorf("expire pending bookings: %w", err)
	}

	return nil
}

func scanEventDetails(scanner interface {
	Scan(dest ...any) error
}) (entity.EventDetails, error) {
	var event entity.EventDetails

	if err := scanner.Scan(
		&event.ID,
		&event.Name,
		&event.StartsAt,
		&event.Capacity,
		&event.RequiresConfirmation,
		&event.BookingTTLMinutes,
		&event.CreatedAt,
		&event.UpdatedAt,
		&event.PendingBookings,
		&event.ConfirmedBookings,
	); err != nil {
		return entity.EventDetails{}, err
	}

	return event, nil
}

func scanUser(scanner interface {
	Scan(dest ...any) error
}) (entity.User, error) {
	var user entity.User

	if err := scanner.Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.TelegramChatID,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		return entity.User{}, err
	}

	return user, nil
}

func scanBooking(scanner interface {
	Scan(dest ...any) error
}) (entity.Booking, error) {
	var booking entity.Booking
	var expiresAt sql.NullTime
	var confirmedAt sql.NullTime

	if err := scanner.Scan(
		&booking.ID,
		&booking.EventID,
		&booking.UserID,
		&booking.CustomerName,
		&booking.CustomerEmail,
		&booking.Status,
		&expiresAt,
		&confirmedAt,
		&booking.CreatedAt,
		&booking.UpdatedAt,
	); err != nil {
		return entity.Booking{}, err
	}

	if expiresAt.Valid {
		value := expiresAt.Time.UTC()
		booking.ExpiresAt = &value
	}

	if confirmedAt.Valid {
		value := confirmedAt.Time.UTC()
		booking.ConfirmedAt = &value
	}

	return booking, nil
}

func scanBookingWithUser(scanner interface {
	Scan(dest ...any) error
}) (entity.Booking, error) {
	var booking entity.Booking
	var user entity.User
	var expiresAt sql.NullTime
	var confirmedAt sql.NullTime
	var userID sql.NullString
	var userName sql.NullString
	var userEmail sql.NullString
	var userTelegramChatID sql.NullString
	var userCreatedAt sql.NullTime
	var userUpdatedAt sql.NullTime

	if err := scanner.Scan(
		&booking.ID,
		&booking.EventID,
		&booking.UserID,
		&booking.CustomerName,
		&booking.CustomerEmail,
		&booking.Status,
		&expiresAt,
		&confirmedAt,
		&booking.CreatedAt,
		&booking.UpdatedAt,
		&userID,
		&userName,
		&userEmail,
		&userTelegramChatID,
		&userCreatedAt,
		&userUpdatedAt,
	); err != nil {
		return entity.Booking{}, err
	}

	if expiresAt.Valid {
		value := expiresAt.Time.UTC()
		booking.ExpiresAt = &value
	}

	if confirmedAt.Valid {
		value := confirmedAt.Time.UTC()
		booking.ConfirmedAt = &value
	}

	if userID.Valid {
		user = entity.User{
			ID:             userID.String,
			Name:           userName.String,
			Email:          userEmail.String,
			TelegramChatID: userTelegramChatID.String,
		}
		if userCreatedAt.Valid {
			user.CreatedAt = userCreatedAt.Time.UTC()
		}
		if userUpdatedAt.Valid {
			user.UpdatedAt = userUpdatedAt.Time.UTC()
		}
		booking.User = &user
	}

	return booking, nil
}

func scanExpiredBookingNotice(scanner interface {
	Scan(dest ...any) error
}) (entity.ExpiredBookingNotice, error) {
	var notice entity.ExpiredBookingNotice
	var expiresAt sql.NullTime
	var confirmedAt sql.NullTime
	var userID sql.NullString
	var userName sql.NullString
	var userEmail sql.NullString
	var userTelegramChatID sql.NullString
	var userCreatedAt sql.NullTime
	var userUpdatedAt sql.NullTime

	if err := scanner.Scan(
		&notice.Booking.ID,
		&notice.Booking.EventID,
		&notice.Booking.UserID,
		&notice.Booking.CustomerName,
		&notice.Booking.CustomerEmail,
		&notice.Booking.Status,
		&expiresAt,
		&confirmedAt,
		&notice.Booking.CreatedAt,
		&notice.Booking.UpdatedAt,
		&userID,
		&userName,
		&userEmail,
		&userTelegramChatID,
		&userCreatedAt,
		&userUpdatedAt,
		&notice.Event.ID,
		&notice.Event.Name,
		&notice.Event.StartsAt,
		&notice.Event.Capacity,
		&notice.Event.RequiresConfirmation,
		&notice.Event.BookingTTLMinutes,
		&notice.Event.CreatedAt,
		&notice.Event.UpdatedAt,
	); err != nil {
		return entity.ExpiredBookingNotice{}, err
	}

	if expiresAt.Valid {
		value := expiresAt.Time.UTC()
		notice.Booking.ExpiresAt = &value
	}

	if confirmedAt.Valid {
		value := confirmedAt.Time.UTC()
		notice.Booking.ConfirmedAt = &value
	}

	if userID.Valid {
		notice.User = entity.User{
			ID:             userID.String,
			Name:           userName.String,
			Email:          userEmail.String,
			TelegramChatID: userTelegramChatID.String,
		}
		if userCreatedAt.Valid {
			notice.User.CreatedAt = userCreatedAt.Time.UTC()
		}
		if userUpdatedAt.Valid {
			notice.User.UpdatedAt = userUpdatedAt.Time.UTC()
		}
		notice.Booking.User = &notice.User
	}

	return notice, nil
}