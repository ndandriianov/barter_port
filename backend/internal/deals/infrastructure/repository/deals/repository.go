package deals

import (
	"barter-port/internal/deals/domain"
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
// CreateDraft
// ================================================================================

// CreateDraft inserts a new draft deal into the database and returns its ID.
//
// No domain errors
func (r *Repository) CreateDraft(
	ctx context.Context,
	tx pgx.Tx,
	authorID uuid.UUID,
	name *string,
	description *string,
	offers []domain.OfferIDAndInfo,
) (uuid.UUID, error) {

	dealsQuery := `
		INSERT INTO draft_deals (author_id, name, description)
		VALUES ($1, $2, $3)
		RETURNING id;`

	var id uuid.UUID
	err := tx.QueryRow(ctx, dealsQuery, authorID, name, description).Scan(&id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("sql: deals: %w", err)
	}

	dealsOffersQuery := `
		INSERT INTO draft_deal_offers (draft_deal_id, offer_id, quantity)
		VALUES ($1, $2, $3);`

	for _, offer := range offers {
		_, err = tx.Exec(ctx, dealsOffersQuery, id, offer.ID, offer.Info.Quantity)
		if err != nil {
			return uuid.Nil, fmt.Errorf("sql: deals items: %w, itemID: %s", err, offer.ID)
		}
	}

	return id, nil
}

// ================================================================================
// GetDraftIDsByAuthor
// ================================================================================

// GetDraftIDsByAuthor retrieves the IDs of draft deals created by a specific author.
//
// No domain errors
func (r *Repository) GetDraftIDsByAuthor(ctx context.Context, exec db.DB, authorID uuid.UUID) ([]uuid.UUID, error) {
	query := `
		SELECT id
		FROM draft_deals
		WHERE author_id = $1;`

	rows, err := exec.Query(ctx, query, authorID)
	if err != nil {
		return nil, fmt.Errorf("sql: %w", err)
	}
	defer rows.Close()

	ids, err := pgx.CollectRows(rows, pgx.RowTo[uuid.UUID])
	if err != nil {
		return nil, fmt.Errorf("collect rows: %w", err)
	}

	return ids, nil
}

// ================================================================================
// GetDraftByID
// ================================================================================

// GetDraftByID retrieves a draft deal by its ID, including its associated items.
//
// Errors:
//   - domain.ErrDraftNotFound
//   - SQL errors are wrapped
func (r *Repository) GetDraftByID(ctx context.Context, exec db.DB, id uuid.UUID) (domain.Draft, error) {
	query := `
		SELECT d.id,
		       d.author_id,
		       d.name,
		       d.description,
		       d.created_at,
		       d.updated_at,
		       i.id,
		       i.author_id,
		       i.name,
		       i.type,
		       i.action,
		       i.description,
		       i.created_at,
		       i.views,
		       ddi.quantity,
		       ddi.confirmed
		FROM draft_deals d
		LEFT JOIN draft_deal_offers ddi ON ddi.draft_deal_id = d.id
		LEFT JOIN offers i ON i.id = ddi.offer_id
		WHERE d.id = $1;`

	rows, err := exec.Query(ctx, query, id)
	if err != nil {
		return domain.Draft{}, fmt.Errorf("sql: %w", err)
	}
	defer rows.Close()

	var draft domain.Draft
	found := false

	for rows.Next() {
		var offerID *uuid.UUID
		var offerAuthorID *uuid.UUID
		var offerName *string
		var offerType *string
		var offerAction *string
		var offerDescription *string
		var offerCreatedAt *time.Time
		var offerViews *int
		var offerQuantity *int
		var offerConfirmed *bool

		if err = rows.Scan(
			&draft.ID,
			&draft.AuthorID,
			&draft.Name,
			&draft.Description,
			&draft.CreatedAt,
			&draft.UpdatedAt,
			&offerID,
			&offerAuthorID,
			&offerName,
			&offerType,
			&offerAction,
			&offerDescription,
			&offerCreatedAt,
			&offerViews,
			&offerQuantity,
			&offerConfirmed,
		); err != nil {
			return domain.Draft{}, fmt.Errorf("scan draft item: %w", err)
		}

		found = true

		if offerID == nil {
			continue
		}

		if offerAuthorID == nil || offerName == nil || offerType == nil || offerAction == nil || offerDescription == nil || offerCreatedAt == nil || offerViews == nil || offerQuantity == nil {
			return domain.Draft{}, fmt.Errorf("scan draft item: item has null required fields")
		}

		itemTypeValue, err := domain.ItemTypeString(*offerType)
		if err != nil {
			return domain.Draft{}, fmt.Errorf("item type: %w", err)
		}

		offerActionValue, err := domain.OfferActionString(*offerAction)
		if err != nil {
			return domain.Draft{}, fmt.Errorf("item action: %w", err)
		}

		draft.Items = append(draft.Items, domain.OfferWithInfo{
			Offer: domain.Offer{
				ID:          *offerID,
				AuthorId:    *offerAuthorID,
				Name:        *offerName,
				Type:        itemTypeValue,
				Action:      offerActionValue,
				Description: *offerDescription,
				CreatedAt:   *offerCreatedAt,
				Views:       *offerViews,
			},
			Info: domain.OfferInfo{
				Quantity:  *offerQuantity,
				Confirmed: offerConfirmed,
			},
		})
	}

	if err = rows.Err(); err != nil {
		return domain.Draft{}, fmt.Errorf("rows: %w", err)
	}

	if !found {
		return domain.Draft{}, domain.ErrDraftNotFound
	}

	return draft, nil
}
