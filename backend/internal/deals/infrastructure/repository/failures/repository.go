package failures

import (
	"barter-port/internal/deals/domain"
	"barter-port/internal/deals/domain/enums"
	"barter-port/internal/deals/domain/htypes"
	dealsrepo "barter-port/internal/deals/infrastructure/repository/deals"
	"barter-port/pkg/db"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Repository struct {
	*dealsrepo.Repository
}

func NewRepository(base *dealsrepo.Repository) *Repository {
	return &Repository{Repository: base}
}

func (r *Repository) LockDeal(ctx context.Context, tx pgx.Tx, dealID uuid.UUID) (enums.DealStatus, error) {
	var status enums.DealStatus
	err := tx.QueryRow(ctx, `SELECT status FROM deals WHERE id = $1 FOR UPDATE`, dealID).Scan(&status)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, domain.ErrDealNotFound
	}
	if err != nil {
		return 0, fmt.Errorf("sql lock deal: %w", err)
	}

	return status, nil
}

func (r *Repository) HasPendingFailureReview(ctx context.Context, exec db.DB, dealID uuid.UUID) (bool, error) {
	return r.hasFailureReview(ctx, exec, dealID, true)
}

func (r *Repository) HasFailureRecord(ctx context.Context, exec db.DB, dealID uuid.UUID) (bool, error) {
	var exists bool
	err := exec.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM deal_failures WHERE deal_id = $1)`, dealID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("sql has failure record: %w", err)
	}
	return exists, nil
}

func (r *Repository) hasFailureReview(
	ctx context.Context,
	exec db.DB,
	dealID uuid.UUID,
	pendingOnly bool,
) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM deal_failures WHERE deal_id = $1`
	if pendingOnly {
		query += ` AND confirmed_by_admin IS NULL`
	}
	query += `)`

	var exists bool
	err := exec.QueryRow(ctx, query, dealID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("sql has pending failure review: %w", err)
	}

	return exists, nil
}

func (r *Repository) SetFailureVote(
	ctx context.Context,
	tx pgx.Tx,
	dealID, userID, votedFor uuid.UUID,
) error {
	tag, err := tx.Exec(ctx, `
		UPDATE participants
		SET failure_blame_vote_for = $3
		WHERE deal_id = $1 AND user_id = $2`,
		dealID, userID, votedFor,
	)
	if err != nil {
		return fmt.Errorf("sql set failure vote: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrForbidden
	}

	return nil
}

func (r *Repository) ClearFailureVote(ctx context.Context, tx pgx.Tx, dealID, userID uuid.UUID) error {
	tag, err := tx.Exec(ctx, `
		UPDATE participants
		SET failure_blame_vote_for = NULL
		WHERE deal_id = $1 AND user_id = $2`,
		dealID, userID,
	)
	if err != nil {
		return fmt.Errorf("sql clear failure vote: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrForbidden
	}

	return nil
}

func (r *Repository) GetFailureVotes(ctx context.Context, exec db.DB, dealID uuid.UUID) ([]htypes.FailureVote, error) {
	rows, err := exec.Query(ctx, `
		SELECT user_id, failure_blame_vote_for
		FROM participants
		WHERE deal_id = $1
		  AND failure_blame_vote_for IS NOT NULL
		ORDER BY user_id`,
		dealID,
	)
	if err != nil {
		return nil, fmt.Errorf("sql get failure votes: %w", err)
	}
	defer rows.Close()

	result := make([]htypes.FailureVote, 0)
	for rows.Next() {
		var item htypes.FailureVote
		if err = rows.Scan(&item.UserID, &item.Vote); err != nil {
			return nil, fmt.Errorf("scan failure vote: %w", err)
		}
		result = append(result, item)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows failure votes: %w", err)
	}

	return result, nil
}

func (r *Repository) CreateFailureRecord(
	ctx context.Context,
	tx pgx.Tx,
	dealID uuid.UUID,
	userID *uuid.UUID,
) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO deal_failures (deal_id, user_id)
		VALUES ($1, $2)
		ON CONFLICT (deal_id) DO NOTHING`,
		dealID, userID,
	)
	if err != nil {
		return fmt.Errorf("sql create failure record: %w", err)
	}

	return nil
}

func (r *Repository) GetFailureRecord(ctx context.Context, exec db.DB, dealID uuid.UUID) (htypes.FailureRecord, error) {
	return scanFailureRecord(exec.QueryRow(ctx, `
		SELECT deal_id, user_id, confirmed_by_admin, admin_comment, punishment_points
		FROM deal_failures
		WHERE deal_id = $1`,
		dealID,
	))
}

func (r *Repository) ResolveFailureRecord(
	ctx context.Context,
	tx pgx.Tx,
	dealID uuid.UUID,
	confirmed bool,
	userID *uuid.UUID,
	punishmentPoints *int,
	comment *string,
) (htypes.FailureRecord, error) {
	record, err := scanFailureRecord(tx.QueryRow(ctx, `
		UPDATE deal_failures
		SET user_id = $2,
		    confirmed_by_admin = $3,
		    admin_comment = $4,
		    punishment_points = $5
		WHERE deal_id = $1
		  AND confirmed_by_admin IS NULL
		RETURNING deal_id, user_id, confirmed_by_admin, admin_comment, punishment_points`,
		dealID, userID, confirmed, comment, punishmentPoints,
	))
	if err == nil {
		return record, nil
	}
	if !errors.Is(err, domain.ErrFailureNotFound) {
		return htypes.FailureRecord{}, err
	}

	exists, existsErr := r.HasFailureRecord(ctx, tx, dealID)
	if existsErr != nil {
		return htypes.FailureRecord{}, existsErr
	}
	if exists {
		return htypes.FailureRecord{}, domain.ErrFailureAlreadyResolved
	}

	return htypes.FailureRecord{}, domain.ErrFailureNotFound
}

func (r *Repository) GetFailureReviewDeals(
	ctx context.Context,
	exec db.DB,
) ([]htypes.DealIDWithParticipantIDs, error) {
	rows, err := exec.Query(ctx, `
		SELECT d.id,
		       d.status,
		       d.name,
		       COALESCE(array_agg(p.user_id ORDER BY p.user_id) FILTER (WHERE p.user_id IS NOT NULL), '{}')
		FROM deals d
		JOIN deal_failures df
		  ON df.deal_id = d.id
		 AND df.confirmed_by_admin IS NULL
		LEFT JOIN participants p
		  ON p.deal_id = d.id
		GROUP BY d.id, d.status, d.name
		ORDER BY d.created_at DESC, d.id`)
	if err != nil {
		return nil, fmt.Errorf("sql get failure review deals: %w", err)
	}
	defer rows.Close()

	result := make([]htypes.DealIDWithParticipantIDs, 0)
	for rows.Next() {
		var item htypes.DealIDWithParticipantIDs
		var statusStr string
		if err = rows.Scan(&item.ID, &statusStr, &item.Name, &item.ParticipantIDs); err != nil {
			return nil, fmt.Errorf("scan failure review deal: %w", err)
		}

		item.Status, err = enums.DealStatusString(statusStr)
		if err != nil {
			return nil, fmt.Errorf("parse failure review deal status: %w", err)
		}
		result = append(result, item)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows failure review deals: %w", err)
	}

	dealIDs := make([]uuid.UUID, 0, len(result))
	for _, deal := range result {
		dealIDs = append(dealIDs, deal.ID)
	}

	itemsByDealID, err := r.loadItemsByDealIDs(ctx, exec, dealIDs)
	if err != nil {
		return nil, err
	}

	for i := range result {
		result[i].Items = itemsByDealID[result[i].ID]
	}

	return result, nil
}

func (r *Repository) loadItemsByDealIDs(
	ctx context.Context,
	exec db.DB,
	dealIDs []uuid.UUID,
) (map[uuid.UUID][]domain.Item, error) {
	result := make(map[uuid.UUID][]domain.Item, len(dealIDs))
	if len(dealIDs) == 0 {
		return result, nil
	}

	const query = `
		SELECT i.deal_id,
		       i.id,
		       i.offer_id,
		       i.author_id,
		       i.provider_id,
		       i.receiver_id,
		       i.name,
		       i.description,
		       i.type,
		       i.updated_at,
		       i.quantity,
		       COALESCE((SELECT array_agg(ip.id ORDER BY ip.position) FROM items_photos ip WHERE ip.item_id = i.id), '{}'::uuid[]) AS photo_ids,
		       COALESCE((SELECT array_agg(ip.url ORDER BY ip.position) FROM items_photos ip WHERE ip.item_id = i.id), '{}'::text[]) AS photo_urls
		FROM items i
		WHERE i.deal_id = ANY($1)
		ORDER BY i.deal_id, i.id`

	rows, err := exec.Query(ctx, query, dealIDs)
	if err != nil {
		return nil, fmt.Errorf("sql get failure review deal items: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			dealID   uuid.UUID
			item     domain.Item
			itemType string
		)
		if err = rows.Scan(
			&dealID,
			&item.ID,
			&item.OfferID,
			&item.AuthorID,
			&item.ProviderID,
			&item.ReceiverID,
			&item.Name,
			&item.Description,
			&itemType,
			&item.UpdatedAt,
			&item.Quantity,
			&item.PhotoIDs,
			&item.PhotoURLs,
		); err != nil {
			return nil, fmt.Errorf("scan failure review deal item: %w", err)
		}

		item.Type, err = enums.ItemTypeString(itemType)
		if err != nil {
			return nil, fmt.Errorf("failure review item type: %w", err)
		}
		result[dealID] = append(result[dealID], item)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows failure review deal items: %w", err)
	}

	return result, nil
}

func scanFailureRecord(row interface{ Scan(...any) error }) (htypes.FailureRecord, error) {
	var item htypes.FailureRecord
	err := row.Scan(
		&item.DealID,
		&item.UserID,
		&item.ConfirmedByAdmin,
		&item.AdminComment,
		&item.PunishmentPoints,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return htypes.FailureRecord{}, domain.ErrFailureNotFound
	}
	if err != nil {
		return htypes.FailureRecord{}, fmt.Errorf("scan failure record: %w", err)
	}

	return item, nil
}
