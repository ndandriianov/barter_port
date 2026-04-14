package reputation_events

import (
	reputation_inbox "barter-port/internal/users/infrastructure/repository/reputation-inbox"
	"barter-port/pkg/db"
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/net/context"
)

type Repository struct{}

func NewRepository() *Repository {
	return &Repository{}
}

// WriteReputationEvent inserts a reputation event for idempotency tracking.
// Returns (true, nil) if inserted, (false, nil) if already existed (duplicate).
func (r *Repository) WriteReputationEvent(ctx context.Context, exec db.DB, msg reputation_inbox.InboxMessage) (inserted bool, err error) {
	const query = `
		INSERT INTO user_reputation_events (id, user_id, source_type, source_id, delta, created_at, comment)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (source_type, source_id) DO NOTHING`

	tag, err := exec.Exec(ctx, query,
		uuid.New(), msg.UserID, msg.SourceType, msg.SourceID, msg.Delta, time.Now(), msg.Comment,
	)
	if err != nil {
		return false, fmt.Errorf("sql write reputation event: %w", err)
	}
	return tag.RowsAffected() > 0, nil
}

// ApplyReputationDelta adds delta to user's reputation_points atomically.
func (r *Repository) ApplyReputationDelta(ctx context.Context, exec db.DB, userID uuid.UUID, delta int) error {
	_, err := exec.Exec(ctx,
		`UPDATE users SET reputation_points = reputation_points + $2 WHERE id = $1`,
		userID, delta,
	)
	return err
}
