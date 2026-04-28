package user

import (
	"barter-port/internal/users/domain"
	"barter-port/pkg/db"
	"barter-port/pkg/repox"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/net/context"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// AddUser adds a new user to the repository.
//
// Errors:
//   - domain.ErrUserAlreadyExists: Occurs if a user with the same ID already exists in the repository.
func (r *Repository) AddUser(ctx context.Context, db db.DB, userID uuid.UUID) error {
	query := `
		INSERT INTO users (id)
		VALUES ($1)
	`

	_, err := db.Exec(ctx, query, userID)
	if err != nil && repox.IsUniqueViolation(err) {
		return domain.ErrUserAlreadyExists
	}
	return err
}

// GetUserById returns user by id.
//
// Errors:
//   - domain.ErrUserNotFound: Occurs if no user is found with the given id.
func (r *Repository) GetUserById(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	query := `
		SELECT id, name, bio, avatar_url, phone_number, current_latitude, current_longitude
		FROM users
		WHERE id = $1
	`

	rows, err := r.db.Query(ctx, query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	user, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.User])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

// DeleteUser should be used only with transaction, with deleting auth-service user and
// transferring userId to deleted users table
func (r *Repository) DeleteUser(ctx context.Context, id uuid.UUID) error {
	//TODO implement me
	panic("implement me")
}

// UpdateName updates the name of a user.
//
// Errors:
//   - domain.ErrUserNotFound: Occurs if no user is found with the given id.
func (r *Repository) UpdateName(ctx context.Context, id uuid.UUID, name string) error {
	query := `
		UPDATE users
		SET name = $2
		WHERE id = $1
	`

	tag, err := r.db.Exec(ctx, query, id, name)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrUserNotFound
	}

	return nil
}

// UpdateBio updates the bio of a user.
//
// Errors:
//   - domain.ErrUserNotFound: Occurs if no user is found with the given id.
func (r *Repository) UpdateBio(ctx context.Context, id uuid.UUID, bio *string) error {
	query := `
		UPDATE users
		SET bio = $2
		WHERE id = $1
	`

	tag, err := r.db.Exec(ctx, query, id, bio)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrUserNotFound
	}

	return nil
}

// UpdateAvatarURL updates the avatar URL of a user.
//
// Errors:
//   - domain.ErrUserNotFound: Occurs if no user is found with the given id.
func (r *Repository) UpdateAvatarURL(ctx context.Context, id uuid.UUID, avatarURL *string) error {
	query := `
		UPDATE users
		SET avatar_url = $2
		WHERE id = $1
	`

	tag, err := r.db.Exec(ctx, query, id, avatarURL)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrUserNotFound
	}

	return nil
}

// UpdatePhoneNumber updates the phone number of a user.
//
// Errors:
//   - domain.ErrUserNotFound: Occurs if no user is found with the given id.
func (r *Repository) UpdatePhoneNumber(ctx context.Context, id uuid.UUID, phoneNumber *string) error {
	query := `
		UPDATE users
		SET phone_number = $2
		WHERE id = $1
	`

	tag, err := r.db.Exec(ctx, query, id, phoneNumber)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrUserNotFound
	}

	return nil
}

// ListUsers returns all users.
//
// No domain Errors
func (r *Repository) ListUsers(ctx context.Context) ([]domain.User, error) {
	query := `
		SELECT id, name, bio, avatar_url, phone_number, current_latitude, current_longitude
		FROM users
		ORDER BY id
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("sql: %w", err)
	}
	defer rows.Close()

	users, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.User])
	if err != nil {
		return nil, fmt.Errorf("collect: %w", err)
	}

	return users, nil
}

// GetReputationPoints returns the reputation_points of a user.
//
// Errors:
//   - domain.ErrUserNotFound
func (r *Repository) GetReputationPoints(ctx context.Context, id uuid.UUID) (int, error) {
	var points int
	err := r.db.QueryRow(ctx, `SELECT reputation_points FROM users WHERE id = $1`, id).Scan(&points)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, domain.ErrUserNotFound
		}
		return 0, err
	}
	return points, nil
}

// GetReputationEvents returns reputation events of a user ordered from newest to oldest.
//
// No domain Errors
func (r *Repository) GetReputationEvents(ctx context.Context, id uuid.UUID) ([]domain.ReputationEvent, error) {
	const query = `
		SELECT id, source_type, source_id, delta, created_at, comment
		FROM user_reputation_events
		WHERE user_id = $1
		ORDER BY created_at DESC, id DESC
	`

	rows, err := r.db.Query(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("sql: %w", err)
	}
	defer rows.Close()

	events, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.ReputationEvent])
	if err != nil {
		return nil, fmt.Errorf("collect: %w", err)
	}

	return events, nil
}

func (r *Repository) GetReputationStats(ctx context.Context) (average float64, median float64, err error) {
	const query = `
		SELECT
			COALESCE(AVG(reputation_points)::float8, 0),
			COALESCE(PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY reputation_points)::float8, 0)
		FROM users
	`

	if err = r.db.QueryRow(ctx, query).Scan(&average, &median); err != nil {
		return 0, 0, err
	}

	return average, median, nil
}

func (r *Repository) GetTopUsersByReputation(ctx context.Context, limit int) ([]domain.User, []int, error) {
	const query = `
		SELECT id, name, reputation_points
		FROM users
		ORDER BY reputation_points DESC, id ASC
		LIMIT $1
	`

	rows, err := r.db.Query(ctx, query, limit)
	if err != nil {
		return nil, nil, fmt.Errorf("sql: %w", err)
	}
	defer rows.Close()

	users := make([]domain.User, 0, limit)
	points := make([]int, 0, limit)
	for rows.Next() {
		var (
			user        domain.User
			pointsValue int
		)
		if err = rows.Scan(&user.Id, &user.Name, &pointsValue); err != nil {
			return nil, nil, fmt.Errorf("scan: %w", err)
		}
		users = append(users, user)
		points = append(points, pointsValue)
	}

	return users, points, nil
}

func (r *Repository) GetFollowersCount(ctx context.Context, id uuid.UUID) (int, error) {
	var count int
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM subscriptions WHERE target_user_id = $1`, id).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (r *Repository) GetSubscriptionsCount(ctx context.Context, id uuid.UUID) (int, error) {
	var count int
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM subscriptions WHERE subscriber_id = $1`, id).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

// Subscribe subscribes subscriberID to targetUserID.
//
// Errors:
//   - domain.ErrAlreadySubscribed: Occurs if the subscription already exists.
//   - domain.ErrUserNotFound: Occurs if targetUserID does not exist.
func (r *Repository) Subscribe(ctx context.Context, subscriberID, targetUserID uuid.UUID) error {
	query := `
		INSERT INTO subscriptions (subscriber_id, target_user_id)
		VALUES ($1, $2)
	`

	_, err := r.db.Exec(ctx, query, subscriberID, targetUserID)
	if err != nil {
		if repox.IsUniqueViolation(err) {
			return domain.ErrAlreadySubscribed
		}
		if isForeignKeyViolation(err) {
			return domain.ErrUserNotFound
		}
		return err
	}
	return nil
}

// Unsubscribe removes the subscription of subscriberID from targetUserID.
//
// Errors:
//   - domain.ErrNotSubscribed: Occurs if the subscription does not exist.
func (r *Repository) Unsubscribe(ctx context.Context, subscriberID, targetUserID uuid.UUID) error {
	query := `
		DELETE FROM subscriptions
		WHERE subscriber_id = $1 AND target_user_id = $2
	`

	tag, err := r.db.Exec(ctx, query, subscriberID, targetUserID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotSubscribed
	}
	return nil
}

// IsSubscribed returns true if subscriberID is subscribed to targetUserID.
//
// No domain Errors
func (r *Repository) IsSubscribed(ctx context.Context, subscriberID, targetUserID uuid.UUID) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx,
		`SELECT EXISTS (SELECT 1 FROM subscriptions WHERE subscriber_id = $1 AND target_user_id = $2)`,
		subscriberID, targetUserID,
	).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

// GetSubscriptions returns the list of users that userID is subscribed to.
//
// No domain Errors
func (r *Repository) GetSubscriptions(ctx context.Context, userID uuid.UUID) ([]domain.User, error) {
	query := `
		SELECT u.id, u.name, u.bio, u.avatar_url, u.phone_number, u.current_latitude, u.current_longitude
		FROM subscriptions s
		JOIN users u ON u.id = s.target_user_id
		WHERE s.subscriber_id = $1
		ORDER BY u.id
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("sql: %w", err)
	}
	defer rows.Close()

	users, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.User])
	if err != nil {
		return nil, fmt.Errorf("collect: %w", err)
	}
	return users, nil
}

// GetSubscribers returns the list of users subscribed to userID.
//
// No domain Errors
func (r *Repository) GetSubscribers(ctx context.Context, userID uuid.UUID) ([]domain.User, error) {
	query := `
		SELECT u.id, u.name, u.bio, u.avatar_url, u.phone_number, u.current_latitude, u.current_longitude
		FROM subscriptions s
		JOIN users u ON u.id = s.subscriber_id
		WHERE s.target_user_id = $1
		ORDER BY u.id
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("sql: %w", err)
	}
	defer rows.Close()

	users, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.User])
	if err != nil {
		return nil, fmt.Errorf("collect: %w", err)
	}
	return users, nil
}

// UpdateLocation updates the current geolocation of a user. Pass nil to clear both fields.
//
// Errors:
//   - domain.ErrUserNotFound: Occurs if no user is found with the given id.
func (r *Repository) UpdateLocation(ctx context.Context, id uuid.UUID, latitude, longitude *float64) error {
	query := `
		UPDATE users
		SET current_latitude = $2, current_longitude = $3
		WHERE id = $1
	`

	tag, err := r.db.Exec(ctx, query, id, latitude, longitude)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}

// GetLocation returns the current geolocation of a user.
//
// Errors:
//   - domain.ErrUserNotFound: Occurs if no user is found with the given id.
func (r *Repository) GetLocation(ctx context.Context, id uuid.UUID) (*float64, *float64, error) {
	var lat, lon *float64
	err := r.db.QueryRow(ctx,
		`SELECT current_latitude, current_longitude FROM users WHERE id = $1`, id,
	).Scan(&lat, &lon)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, domain.ErrUserNotFound
		}
		return nil, nil, err
	}
	return lat, lon, nil
}

func isForeignKeyViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == pgerrcode.ForeignKeyViolation
}

// GetNamesForUserIDs returns a map of user IDs to their corresponding names.
//
// No domain Errors
func (r *Repository) GetNamesForUserIDs(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]*string, error) {
	query := `
		SELECT id, name
		FROM users
		WHERE id = ANY($1)
	`

	rows, err := r.db.Query(ctx, query, ids)
	if err != nil {
		return nil, fmt.Errorf("sql: %w", err)
	}
	defer rows.Close()

	names := make(map[uuid.UUID]*string)
	for rows.Next() {
		var id uuid.UUID
		var name *string

		err = rows.Scan(&id, &name)
		if err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}

		names[id] = name
	}

	return names, nil
}
