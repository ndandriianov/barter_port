package deals

import (
	"barter-port/internal/deals/domain"
	"barter-port/internal/deals/domain/enums"
	"barter-port/internal/deals/domain/htypes"
	"barter-port/pkg/db"
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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
		INSERT INTO items (deal_id, author_id, receiver_id, provider_id, name, description, type, quantity) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	for _, item := range deal.Items {
		_, err = tx.Exec(
			ctx,
			offersQuery,
			id,
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

	return id, nil
}

// ================================================================================
// GET DEAL IDs
// ================================================================================

// GetDealIDs returns deal IDs with participant UUIDs. If userID is non-nil, returns only deals the user participates in.
//
// No domain errors.
func (r *Repository) GetDealIDs(ctx context.Context, exec db.DB, userID *uuid.UUID) ([]htypes.DealIDWithParticipantIDs, error) {
	query := `
		SELECT d.id,
			   COALESCE(array_agg(DISTINCT p) FILTER (WHERE p IS NOT NULL), '{}') AS participant_ids
		FROM deals d
				 LEFT JOIN items i ON i.deal_id = d.id
				 LEFT JOIN LATERAL (VALUES (i.author_id), (i.provider_id), (i.receiver_id)) AS t(p) ON true
		WHERE ($1::uuid IS NULL
			OR EXISTS(SELECT 1
					  FROM items i2
					  WHERE i2.deal_id = d.id
						AND (i2.author_id = $1 OR i2.provider_id = $1 OR i2.receiver_id = $1)))
		GROUP BY d.id`

	rows, err := exec.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("sql get deal ids: %w", err)
	}
	defer rows.Close()

	var result []htypes.DealIDWithParticipantIDs
	for rows.Next() {
		var id uuid.UUID
		var participantIDs []uuid.UUID
		if err = rows.Scan(&id, &participantIDs); err != nil {
			return nil, fmt.Errorf("scan deal id row: %w", err)
		}
		result = append(result, htypes.DealIDWithParticipantIDs{
			ID:             id,
			ParticipantIDs: participantIDs,
		})
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows: %w", err)
	}

	return result, nil
}

// ================================================================================
// GET DEAL BY ID
// ================================================================================

// GetDealByID returns a deal with its items by ID.
//
// Domain errors:
//   - domain.ErrDealNotFound: if no deal with the specified ID exists.
func (r *Repository) GetDealByID(ctx context.Context, exec db.DB, id uuid.UUID) (domain.Deal, error) {
	query := `
		SELECT d.id, d.name, d.description, d.created_at, d.updated_at,
		       i.id, i.author_id, i.provider_id, i.receiver_id,
		       i.name, i.description, i.type, i.updated_at, i.quantity
		FROM deals d
		LEFT JOIN items i ON i.deal_id = d.id
		WHERE d.id = $1`

	rows, err := exec.Query(ctx, query, id)
	if err != nil {
		return domain.Deal{}, fmt.Errorf("sql get deal by id: %w", err)
	}
	defer rows.Close()

	var deal domain.Deal
	found := false

	for rows.Next() {
		var itemID *uuid.UUID
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
			&itemID,
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

	return deal, nil
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
	updateQuery := `
		UPDATE items
		SET name        = COALESCE($4, name),
		    description = COALESCE($5, description),
		    quantity    = COALESCE($6, quantity),
		    updated_at  = NOW()
		WHERE id = $1
		  AND deal_id   = $2
		  AND author_id = $3
		RETURNING id, author_id, provider_id, receiver_id,
		          name, description, type, updated_at, quantity`

	row := exec.QueryRow(ctx, updateQuery, itemID, dealID, userID, patch.Name, patch.Description, patch.Quantity)

	var item domain.Item
	var itemType string
	err := row.Scan(
		&item.ID,
		&item.AuthorID,
		&item.ProviderID,
		&item.ReceiverID,
		&item.Name,
		&item.Description,
		&itemType,
		&item.UpdatedAt,
		&item.Quantity,
	)
	if err == nil {
		item.Type, err = enums.ItemTypeString(itemType)
		if err != nil {
			return domain.Item{}, fmt.Errorf("item type: %w", err)
		}
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
