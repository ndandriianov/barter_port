package deals

import (
	"barter-port/internal/deals/domain"
	"barter-port/internal/deals/domain/enums"
	"barter-port/internal/deals/domain/htypes"
	"barter-port/pkg/db"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const (
	ReceiverID = "receiver_id"
	ProviderID = "provider_id"
)

type Repository struct{}

func NewRepository() *Repository {
	return &Repository{}
}

// ================================================================================
// CREATE DEAL
// ================================================================================

// CreateDeal creates a new deal and its associated items in the database within a transaction.
//
// No domain Errors.
func (r *Repository) CreateDeal(ctx context.Context, tx pgx.Tx, deal domain.Deal) (uuid.UUID, error) {
	dealQuery := `
		INSERT INTO deals (name, description)
		VALUES ($1, $2)
		RETURNING id;`

	var id uuid.UUID
	err := tx.QueryRow(ctx, dealQuery, deal.Name, deal.Description).Scan(&id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("sql deal: %w", err)
	}

	offersQuery := `
		INSERT INTO items (deal_id, offer_id, author_id, receiver_id, provider_id, name, description, type, quantity) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);`

	for _, item := range deal.Items {
		_, err = tx.Exec(
			ctx,
			offersQuery,
			id,
			item.OfferID,
			item.AuthorID,
			item.ReceiverID,
			item.ProviderID,
			item.Name,
			item.Description,
			item.Type,
			item.Quantity,
		)
		if err != nil {
			return uuid.Nil, fmt.Errorf("sql items: %w", err)
		}
	}

	participantsQuery := `
		INSERT INTO participants (deal_id, user_id)
		VALUES ($1, $2);`

	participants := make([]uuid.UUID, 0)
	seen := make(map[uuid.UUID]struct{})
	for _, item := range deal.Items {
		if _, ok := seen[item.AuthorID]; !ok {
			seen[item.AuthorID] = struct{}{}
			participants = append(participants, item.AuthorID)
		}
	}

	for _, participant := range participants {
		_, err = tx.Exec(ctx, participantsQuery, id, participant)
		if err != nil {
			return uuid.Nil, fmt.Errorf("sql participants: %w", err)
		}
	}

	return id, nil
}

// ================================================================================
// GET DEAL IDs
// ================================================================================

// GetDealIDs returns deal IDs with participant UUIDs.
// If userID is non-nil, returns only deals the user participates in.
// If open is true, returns only deals that are not in a final status.
//
// No domain errors.
func (r *Repository) GetDealIDs(ctx context.Context, exec db.DB, userID *uuid.UUID, open bool) ([]htypes.DealIDWithParticipantIDs, error) {
	query := `
		SELECT d.id,
		       d.status,
		       d.name,
			   COALESCE(array_agg(DISTINCT p) FILTER (WHERE p IS NOT NULL), '{}') AS participant_ids
		FROM deals d
				 LEFT JOIN items i ON i.deal_id = d.id
				 LEFT JOIN LATERAL (VALUES (i.author_id), (i.provider_id), (i.receiver_id)) AS t(p) ON true
		WHERE ($1::uuid IS NULL
			OR EXISTS(SELECT 1
					  FROM items i2
					  WHERE i2.deal_id = d.id
						AND (i2.author_id = $1 OR i2.provider_id = $1 OR i2.receiver_id = $1)))
		  AND (NOT $2::boolean OR d.status::text NOT IN ('Completed', 'Cancelled', 'Failed'))
		GROUP BY d.id`

	rows, err := exec.Query(ctx, query, userID, open)
	if err != nil {
		return nil, fmt.Errorf("sql get deal ids: %w", err)
	}
	defer rows.Close()

	var result []htypes.DealIDWithParticipantIDs
	for rows.Next() {
		var id uuid.UUID
		var statusStr string
		var name *string
		var participantIDs []uuid.UUID
		if err = rows.Scan(&id, &statusStr, &name, &participantIDs); err != nil {
			return nil, fmt.Errorf("scan deal id row: %w", err)
		}

		status, err := enums.DealStatusString(statusStr)
		if err != nil {
			return nil, fmt.Errorf("parse deal status: %w", err)
		}

		result = append(result, htypes.DealIDWithParticipantIDs{
			ID:             id,
			Status:         status,
			Name:           name,
			ParticipantIDs: participantIDs,
		})
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows: %w", err)
	}

	return result, nil
}

// ================================================================================
// UPDATE DEAL NAME
// ================================================================================

// UpdateDealName sets a new name for the deal.
//
// No domain errors.
func (r *Repository) UpdateDealName(ctx context.Context, tx pgx.Tx, dealID uuid.UUID, name string) error {
	query := `UPDATE deals SET name = $1, updated_at = NOW() WHERE id = $2`
	_, err := tx.Exec(ctx, query, name, dealID)
	if err != nil {
		return fmt.Errorf("sql update deal name: %w", err)
	}
	return nil
}

// ================================================================================
// GET DEAL BY ID
// ================================================================================

// GetDealByID returns a deal with its items by ID.
//
// Domain errors:
//   - domain.ErrDealNotFound: if no deal with the specified ID exists.
func (r *Repository) GetDealByID(ctx context.Context, tx pgx.Tx, id uuid.UUID) (domain.Deal, error) {
	query := `
		SELECT d.id, d.name, d.description, d.created_at, d.updated_at, d.status,
		       i.id, i.offer_id, i.author_id, i.provider_id, i.receiver_id,
		       i.name, i.description, i.type, i.updated_at, i.quantity
		FROM deals d
		LEFT JOIN items i ON i.deal_id = d.id
		WHERE d.id = $1`

	rows, err := tx.Query(ctx, query, id)
	if err != nil {
		return domain.Deal{}, fmt.Errorf("sql get deal by id: %w", err)
	}
	defer rows.Close()

	var deal domain.Deal
	found := false

	for rows.Next() {
		var itemID *uuid.UUID
		var itemOfferID *uuid.UUID
		var itemAuthorID *uuid.UUID
		var itemProviderID *uuid.UUID
		var itemReceiverID *uuid.UUID
		var itemName *string
		var itemDescription *string
		var itemType *string
		var itemUpdatedAt *time.Time
		var itemQuantity *int64

		if err = rows.Scan(
			&deal.ID,
			&deal.Name,
			&deal.Description,
			&deal.CreatedAt,
			&deal.UpdatedAt,
			&deal.Status,
			&itemID,
			&itemOfferID,
			&itemAuthorID,
			&itemProviderID,
			&itemReceiverID,
			&itemName,
			&itemDescription,
			&itemType,
			&itemUpdatedAt,
			&itemQuantity,
		); err != nil {
			return domain.Deal{}, fmt.Errorf("scan deal row: %w", err)
		}

		found = true

		if itemID == nil {
			continue
		}

		if itemAuthorID == nil || itemName == nil || itemDescription == nil || itemType == nil {
			return domain.Deal{}, fmt.Errorf("scan deal item: null required fields")
		}

		itemTypeValue, err := enums.ItemTypeString(*itemType)
		if err != nil {
			return domain.Deal{}, fmt.Errorf("item type: %w", err)
		}

		deal.Items = append(deal.Items, domain.Item{
			ID:          *itemID,
			OfferID:     itemOfferID,
			AuthorID:    *itemAuthorID,
			ProviderID:  itemProviderID,
			ReceiverID:  itemReceiverID,
			Name:        *itemName,
			Description: *itemDescription,
			Type:        itemTypeValue,
			UpdatedAt:   itemUpdatedAt,
			Quantity:    int(*itemQuantity),
		})
	}

	if err = rows.Err(); err != nil {
		return domain.Deal{}, fmt.Errorf("rows: %w", err)
	}

	if !found {
		return domain.Deal{}, domain.ErrDealNotFound
	}

	participantsQuery := `
		SELECT user_id
		FROM participants
		where deal_id = $1`

	rows, err = tx.Query(ctx, participantsQuery, id)
	if err != nil {
		return domain.Deal{}, fmt.Errorf("sql get participants: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var userID uuid.UUID
		if err = rows.Scan(&userID); err != nil {
			return domain.Deal{}, fmt.Errorf("scan deal row: %w", err)
		}
		deal.Participants = append(deal.Participants, userID)
	}

	if err = rows.Err(); err != nil {
		return domain.Deal{}, fmt.Errorf("rows: %w", err)
	}

	return deal, nil
}

// ================================================================================
// ADD ITEM
// ================================================================================

// AddItem inserts an item into the deal and returns created row.
//
// Domain errors:
//   - domain.ErrDealNotFound: if no deal with the specified ID exists.
func (r *Repository) AddItem(ctx context.Context, exec db.DB, dealID uuid.UUID, item domain.Item) (domain.Item, error) {
	if err := r.ensureDealMutable(ctx, exec, dealID); err != nil {
		return domain.Item{}, err
	}

	query := `
		INSERT INTO items (deal_id, offer_id, author_id, provider_id, receiver_id, name, description, type, quantity)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, offer_id, author_id, provider_id, receiver_id,
		          name, description, type, updated_at, quantity`

	itemCreated, err := scanItem(exec.QueryRow(
		ctx,
		query,
		dealID,
		item.OfferID,
		item.AuthorID,
		item.ProviderID,
		item.ReceiverID,
		item.Name,
		item.Description,
		item.Type,
		item.Quantity,
	))
	if err != nil {
		return domain.Item{}, fmt.Errorf("sql add item: %w", err)
	}

	return itemCreated, nil
}

// ================================================================================
// GET ITEM ROLE IDS
// ================================================================================

// GetItemReceiverID returns receiver_id for the specified item in the deal.
//
// Domain errors:
//   - domain.ErrItemNotFound: if no item with the specified ID exists in the deal.
func (r *Repository) GetItemReceiverID(ctx context.Context, exec db.DB, dealID, itemID uuid.UUID) (*uuid.UUID, error) {
	return r.getItemRoleID(ctx, exec, dealID, itemID, ReceiverID)
}

// GetItemProviderID returns provider_id for the specified item in the deal.
//
// Domain errors:
//   - domain.ErrItemNotFound: if no item with the specified ID exists in the deal.
func (r *Repository) GetItemProviderID(ctx context.Context, exec db.DB, dealID, itemID uuid.UUID) (*uuid.UUID, error) {
	return r.getItemRoleID(ctx, exec, dealID, itemID, ProviderID)
}

func (r *Repository) getItemRoleID(
	ctx context.Context,
	exec db.DB,
	dealID, itemID uuid.UUID,
	column string,
) (*uuid.UUID, error) {
	var query string
	switch column {
	case ReceiverID:
		query = `SELECT receiver_id FROM items WHERE id = $1 AND deal_id = $2 FOR UPDATE`
	case ProviderID:
		query = `SELECT provider_id FROM items WHERE id = $1 AND deal_id = $2 FOR UPDATE`
	default:
		return nil, fmt.Errorf("unsupported role column: %s", column)
	}

	var roleID *uuid.UUID
	err := exec.QueryRow(ctx, query, itemID, dealID).Scan(&roleID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrItemNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("sql get item %s: %w", column, err)
	}

	return roleID, nil
}

// ================================================================================
// UPDATE ITEM
// ================================================================================

// UpdateItem applies a partial update to an item. Only fields set in patch are updated.
//
// Domain errors:
//   - domain.ErrItemNotFound: if no item with the specified ID exists in the deal.
//   - domain.ErrForbidden: if the item exists but userID is not its author.
func (r *Repository) UpdateItem(
	ctx context.Context,
	exec db.DB,
	dealID uuid.UUID,
	itemID uuid.UUID,
	userID uuid.UUID,
	patch htypes.ItemPatch,
) (domain.Item, error) {
	if err := r.ensureDealMutable(ctx, exec, dealID); err != nil {
		return domain.Item{}, err
	}

	updateQuery := `
		UPDATE items
		SET name        = COALESCE($4, name),
		    description = COALESCE($5, description),
		    quantity    = COALESCE($6, quantity),
		    updated_at  = NOW()
		WHERE id = $1
		  AND deal_id   = $2
		  AND author_id = $3
		RETURNING id, offer_id, author_id, provider_id, receiver_id,
		          name, description, type, updated_at, quantity`

	row := exec.QueryRow(ctx, updateQuery, itemID, dealID, userID, patch.Name, patch.Description, patch.Quantity)

	item, err := scanItem(row)
	if err == nil {
		return item, nil
	}
	if err != pgx.ErrNoRows {
		return domain.Item{}, fmt.Errorf("sql update item: %w", err)
	}

	// No rows updated — determine why: item doesn't exist or user is not author
	var authorID uuid.UUID
	checkErr := exec.QueryRow(ctx,
		`SELECT author_id FROM items WHERE id = $1 AND deal_id = $2`,
		itemID, dealID,
	).Scan(&authorID)

	if checkErr == pgx.ErrNoRows {
		return domain.Item{}, domain.ErrItemNotFound
	}
	if checkErr != nil {
		return domain.Item{}, fmt.Errorf("sql check item: %w", checkErr)
	}

	return domain.Item{}, domain.ErrForbidden
}

// ================================================================================
// CLAIM / RELEASE PROVIDER & RECEIVER
// ================================================================================

// ClaimItemProvider sets provider_id = userID if the slot is currently empty.
//
// Domain errors:
//   - domain.ErrItemNotFound: item does not exist in this deal.
//   - domain.ErrRoleAlreadyTaken: provider_id is already set to another user.
func (r *Repository) ClaimItemProvider(ctx context.Context, exec db.DB, dealID, itemID, userID uuid.UUID) (domain.Item, error) {
	const query = `
		UPDATE items
		SET provider_id = $3, updated_at = NOW()
		WHERE id = $1 AND deal_id = $2 AND provider_id IS NULL
		RETURNING id, offer_id, author_id, provider_id, receiver_id,
				  name, description, type, updated_at, quantity`

	const check = `SELECT provider_id FROM items WHERE id = $1 AND deal_id = $2`

	return r.updateItemRole(ctx, exec, dealID, itemID, userID, query, check, domain.ErrRoleAlreadyTaken)
}

// ReleaseItemProvider sets provider_id = NULL if it currently equals userID.
//
// Domain errors:
//   - domain.ErrItemNotFound: item does not exist in this deal.
//   - domain.ErrNotRoleHolder: provider_id is not set to this user.
func (r *Repository) ReleaseItemProvider(ctx context.Context, exec db.DB, dealID, itemID, userID uuid.UUID) (domain.Item, error) {
	const query = `
		UPDATE items
		SET provider_id = NULL, updated_at = NOW()
		WHERE id = $1 AND deal_id = $2 AND provider_id = $3
		RETURNING id, offer_id, author_id, provider_id, receiver_id,
				  name, description, type, updated_at, quantity`

	const check = `SELECT provider_id FROM items WHERE id = $1 AND deal_id = $2`

	return r.updateItemRole(ctx, exec, dealID, itemID, userID, query, check, domain.ErrNotRoleHolder)
}

// ClaimItemReceiver sets receiver_id = userID if the slot is currently empty.
//
// Domain errors:
//   - domain.ErrItemNotFound: item does not exist in this deal.
//   - domain.ErrRoleAlreadyTaken: receiver_id is already set to another user.
func (r *Repository) ClaimItemReceiver(ctx context.Context, exec db.DB, dealID, itemID, userID uuid.UUID) (domain.Item, error) {
	const query = `
		UPDATE items
		SET receiver_id = $3, updated_at = NOW()
		WHERE id = $1 AND deal_id = $2 AND receiver_id IS NULL
		RETURNING id, offer_id, author_id, provider_id, receiver_id,
				  name, description, type, updated_at, quantity`

	const check = `SELECT receiver_id FROM items WHERE id = $1 AND deal_id = $2`

	return r.updateItemRole(ctx, exec, dealID, itemID, userID, query, check, domain.ErrRoleAlreadyTaken)
}

// ReleaseItemReceiver sets receiver_id = NULL if it currently equals userID.
//
// Domain errors:
//   - domain.ErrItemNotFound: item does not exist in this deal.
//   - domain.ErrNotRoleHolder: receiver_id is not set to this user.
func (r *Repository) ReleaseItemReceiver(ctx context.Context, exec db.DB, dealID, itemID, userID uuid.UUID) (domain.Item, error) {
	const query = `
		UPDATE items
		SET receiver_id = NULL, updated_at = NOW()
		WHERE id = $1 AND deal_id = $2 AND receiver_id = $3
		RETURNING id, offer_id, author_id, provider_id, receiver_id,
				  name, description, type, updated_at, quantity`

	const check = `SELECT receiver_id FROM items WHERE id = $1 AND deal_id = $2`

	return r.updateItemRole(ctx, exec, dealID, itemID, userID, query, check, domain.ErrNotRoleHolder)
}

// updateItemRole is a helper that runs an UPDATE on items and falls back to a
// diagnostic SELECT when no rows are updated, returning the appropriate error.
func (r *Repository) updateItemRole(
	ctx context.Context,
	exec db.DB,
	dealID, itemID, userID uuid.UUID,
	query string,
	checkQuery string,
	conflictErr error,
) (domain.Item, error) {
	if err := r.ensureDealMutable(ctx, exec, dealID); err != nil {
		return domain.Item{}, err
	}

	row := exec.QueryRow(ctx, query, itemID, dealID, userID)

	item, err := scanItem(row)
	if err == nil {
		return item, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return domain.Item{}, fmt.Errorf("sql update item role: %w", err)
	}

	// диагностика причины
	var placeholder *uuid.UUID
	checkErr := exec.QueryRow(ctx, checkQuery, itemID, dealID).Scan(&placeholder)
	if errors.Is(checkErr, pgx.ErrNoRows) {
		return domain.Item{}, domain.ErrItemNotFound
	}
	if checkErr != nil {
		return domain.Item{}, fmt.Errorf("sql check item role: %w", checkErr)
	}

	return domain.Item{}, conflictErr
}

// ================================================================================
// DEAL STATUS VOTES
// ================================================================================

// SetStatusVote records or updates the user's requested status vote for a deal.
func (r *Repository) SetStatusVote(ctx context.Context, tx pgx.Tx, dealID, userID uuid.UUID, status enums.DealStatus) error {
	query := `
		UPDATE participants
		SET requested_status = $3
		WHERE deal_id = $1
		  AND user_id = $2`

	_, err := tx.Exec(ctx, query, dealID, userID, status)
	if err != nil {
		return fmt.Errorf("sql: %w", err)
	}
	return nil
}

// GetStatusVotes returns all votes for a deal as a map of userID → requestedStatus.
func (r *Repository) GetStatusVotes(ctx context.Context, exec db.DB, dealID uuid.UUID) (map[uuid.UUID]enums.DealStatus, error) {
	query := `
		SELECT user_id, requested_status
		FROM participants
		WHERE deal_id = $1
		  AND requested_status IS NOT NULL`

	rows, err := exec.Query(ctx, query, dealID)
	if err != nil {
		return nil, fmt.Errorf("sql get status votes: %w", err)
	}
	defer rows.Close()

	votes := make(map[uuid.UUID]enums.DealStatus)
	for rows.Next() {
		var userID uuid.UUID
		var statusStr string

		if err = rows.Scan(&userID, &statusStr); err != nil {
			return nil, fmt.Errorf("scan status vote: %w", err)
		}

		s, err := enums.DealStatusString(statusStr)
		if err != nil {
			return nil, fmt.Errorf("parse deal status: %w", err)
		}
		votes[userID] = s
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows status votes: %w", err)
	}
	return votes, nil
}

// ================================================================================
// PARTICIPANTS
// ================================================================================

func (r *Repository) DeleteParticipant(ctx context.Context, exec db.DB, dealID, userID uuid.UUID) error {
	query := `
		DELETE
		FROM participants
		WHERE deal_id = $1
		  AND user_id = $2;`

	_, err := exec.Exec(ctx, query, dealID, userID)
	if err != nil {
		return fmt.Errorf("sql delete participant: %w", err)
	}
	return nil
}

// ================================================================================
// DEAL STATUS
// ================================================================================

// UpdateDealStatus sets the deal's status and updated_at.
func (r *Repository) UpdateDealStatus(ctx context.Context, tx pgx.Tx, dealID uuid.UUID, status enums.DealStatus) error {
	if err := r.ensureDealMutable(ctx, tx, dealID); err != nil {
		return err
	}

	_, err := tx.Exec(ctx, `
		UPDATE deals SET status = $2, updated_at = NOW() WHERE id = $1`,
		dealID, status,
	)
	if err != nil {
		return fmt.Errorf("sql update deal status: %w", err)
	}
	return nil
}

func (r *Repository) ensureDealMutable(ctx context.Context, exec db.DB, dealID uuid.UUID) error {
	var status enums.DealStatus
	err := exec.QueryRow(ctx, `SELECT status FROM deals WHERE id = $1`, dealID).Scan(&status)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrDealNotFound
	}
	if err != nil {
		return fmt.Errorf("sql get deal status: %w", err)
	}

	switch status {
	case enums.DealStatusCompleted, enums.DealStatusCancelled, enums.DealStatusFailed:
		return domain.ErrInvalidDealStatus
	default:
		return nil
	}
}

// DeleteStatusVotes removes all votes for a deal (called after a status transition).
func (r *Repository) DeleteStatusVotes(ctx context.Context, tx pgx.Tx, dealID uuid.UUID) error {
	query := `
		UPDATE participants
		SET requested_status = NULL
		WHERE deal_id = $1`

	_, err := tx.Exec(ctx, query, dealID)
	if err != nil {
		return fmt.Errorf("sql delete status votes: %w", err)
	}
	return nil
}

// scanItem scans an item row returned from an UPDATE … RETURNING or SELECT query.
func scanItem(row interface{ Scan(...any) error }) (domain.Item, error) {
	var item domain.Item
	var itemType string
	err := row.Scan(
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
	)
	if err != nil {
		return domain.Item{}, err
	}
	item.Type, err = enums.ItemTypeString(itemType)
	if err != nil {
		return domain.Item{}, fmt.Errorf("item type: %w", err)
	}
	return item, nil
}
