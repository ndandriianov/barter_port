package joins

import (
	"barter-port/internal/deals/domain/htypes"
	"barter-port/pkg/db"
	"context"
	"fmt"

	"github.com/google/uuid"
)

type Repository struct{}

func NewRepository() *Repository {
	return &Repository{}
}

// CreateJoinRequest creates a request to join a deal.
func (r *Repository) CreateJoinRequest(ctx context.Context, exec db.DB, userID, dealID uuid.UUID) error {
	query := `
		INSERT INTO join_requests (user_id, deal_id)
		VALUES ($1, $2);`

	_, err := exec.Exec(ctx, query, userID, dealID)
	if err != nil {
		return fmt.Errorf("sql create join request: %w", err)
	}

	return nil
}

// DeleteJoinRequest deletes a request to join a deal.
func (r *Repository) DeleteJoinRequest(ctx context.Context, exec db.DB, userID, dealID uuid.UUID) error {
	query := `
		DELETE FROM join_requests
		WHERE user_id = $1 AND deal_id = $2;`

	_, err := exec.Exec(ctx, query, userID, dealID)
	if err != nil {
		return fmt.Errorf("sql delete join request: %w", err)
	}

	return nil
}

// GetJoinRequestsByDealID returns join requests for a specific deal.
func (r *Repository) GetJoinRequestsByDealID(ctx context.Context, exec db.DB, dealID uuid.UUID) ([]htypes.JoinRequest, error) {
	query := `
		SELECT user_id, deal_id
		FROM join_requests
		WHERE deal_id = $1
		ORDER BY user_id;`

	rows, err := exec.Query(ctx, query, dealID)
	if err != nil {
		return nil, fmt.Errorf("sql get join requests by deal id: %w", err)
	}
	defer rows.Close()

	result := make([]htypes.JoinRequest, 0)
	for rows.Next() {
		var item htypes.JoinRequest
		if err = rows.Scan(&item.UserID, &item.DealID); err != nil {
			return nil, fmt.Errorf("scan join request: %w", err)
		}
		result = append(result, item)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows join requests by deal id: %w", err)
	}

	return result, nil
}

// UpsertJoinRequestVote creates or updates a vote for a join request.
func (r *Repository) UpsertJoinRequestVote(
	ctx context.Context,
	exec db.DB,
	userID, dealID, voterID uuid.UUID,
) error {
	query := `
		INSERT INTO join_requests_votes (user_id, deal_id, voter_id)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, deal_id, voter_id)
		DO NOTHING `

	_, err := exec.Exec(ctx, query, userID, dealID, voterID)
	if err != nil {
		return fmt.Errorf("sql upsert join request vote: %w", err)
	}

	return nil
}

// GetJoinRequestVotesByDealID returns join request votes for a specific deal.
func (r *Repository) GetJoinRequestVotesByDealID(ctx context.Context, exec db.DB, dealID uuid.UUID) ([]htypes.JoinRequestVote, error) {
	query := `
		SELECT user_id, deal_id, voter_id
		FROM join_requests_votes
		WHERE deal_id = $1
		ORDER BY user_id, voter_id;`

	rows, err := exec.Query(ctx, query, dealID)
	if err != nil {
		return nil, fmt.Errorf("sql get join request votes by deal id: %w", err)
	}
	defer rows.Close()

	result := make([]htypes.JoinRequestVote, 0)
	for rows.Next() {
		var item htypes.JoinRequestVote
		if err = rows.Scan(&item.UserID, &item.DealID, &item.VoterID, &item.Vote); err != nil {
			return nil, fmt.Errorf("scan join request vote: %w", err)
		}
		result = append(result, item)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows join request votes by deal id: %w", err)
	}

	return result, nil
}

// AddParticipant adds a user to deal participants if absent.
func (r *Repository) AddParticipant(ctx context.Context, exec db.DB, dealID, userID uuid.UUID) error {
	query := `
		INSERT INTO participants (deal_id, user_id)
		VALUES ($1, $2)
		ON CONFLICT (deal_id, user_id) DO NOTHING;`

	_, err := exec.Exec(ctx, query, dealID, userID)
	if err != nil {
		return fmt.Errorf("sql add participant: %w", err)
	}

	return nil
}
