package drafts

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
func (r *Repository) GetDraftIDsByAuthor(ctx context.Context, exec db.DB, authorID uuid.UUID, createdByMe bool) ([]uuid.UUID, error) {
	var query string
	if createdByMe {
		query = `
			SELECT id
			FROM draft_deals
			WHERE author_id = $1;`
	} else {
		query = `
			SELECT DISTINCT dd.id
			FROM draft_deals dd
			JOIN draft_deal_offers ddo ON dd.id = ddo.draft_deal_id
			JOIN offers o ON o.id = ddo.offer_id 
			WHERE o.author_id = $1;`
	}

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

		if offerAuthorID == nil || offerName == nil || offerType == nil || offerAction == nil ||
			offerDescription == nil || offerCreatedAt == nil || offerViews == nil || offerQuantity == nil {
			return domain.Draft{}, fmt.Errorf("scan draft item: item has null required fields")
		}

		itemTypeValue, err := enums.ItemTypeString(*offerType)
		if err != nil {
			return domain.Draft{}, fmt.Errorf("item type: %w", err)
		}

		offerActionValue, err := enums.OfferActionString(*offerAction)
		if err != nil {
			return domain.Draft{}, fmt.Errorf("item action: %w", err)
		}

		draft.Offers = append(draft.Offers, domain.OfferWithInfo{
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

// ================================================================================
// ConfirmDraftByID
// ================================================================================

// ConfirmDraftByID marks current user's offer in the draft as confirmed.
//
// Errors:
//   - domain.ErrDraftNotFound: if draft deal with the specified ID does not exist.
//   - domain.ErrUserNotInDraft: if user has no offers in the specified draft deal.
//   - SQL errors are wrapped.
func (r *Repository) ConfirmDraftByID(ctx context.Context, exec db.DB, id uuid.UUID, userID uuid.UUID) error {
	updateQuery := `
		WITH updated_draft_deal_offer AS (
			UPDATE draft_deal_offers ddo
			SET confirmed = true
			FROM offers o
			WHERE o.id = ddo.offer_id
				AND ddo.draft_deal_id = $1
				AND o.author_id = $2
				AND ddo.confirmed = false
			RETURNING ddo.draft_deal_id
		),
		updated_draft_deal AS (
			UPDATE draft_deals dd
			SET updated_at = now()
			FROM updated_draft_deal_offer udo
			WHERE dd.id = udo.draft_deal_id
			RETURNING dd.id
		)
		SELECT EXISTS(
			SELECT 1
			FROM draft_deals dd
			WHERE dd.id = $1
		),
		EXISTS(
			SELECT 1
			FROM draft_deal_offers ddo
			JOIN offers o ON o.id = ddo.offer_id
			WHERE ddo.draft_deal_id = $1
				AND o.author_id = $2
		);`

	var draftExists bool
	var userInDraft bool
	err := exec.QueryRow(ctx, updateQuery, id, userID).Scan(&draftExists, &userInDraft)
	if err != nil {
		return fmt.Errorf("sql confirm draft: %w", err)
	}

	if !draftExists {
		return domain.ErrDraftNotFound
	}

	if !userInDraft {
		return domain.ErrUserNotInDraft
	}

	return nil
}

// ================================================================================
// UnconfirmDraftByID
// ================================================================================

// UnconfirmDraftByID marks current user's offer in the draft as unconfirmed.
//
// Errors:
//   - domain.ErrDraftNotFound: if draft deal with the specified ID does not exist.
//   - domain.ErrUserNotInDraft: if user has no offers in the specified draft deal.
//   - SQL errors are wrapped.
func (r *Repository) UnconfirmDraftByID(ctx context.Context, exec db.DB, id uuid.UUID, userID uuid.UUID) error {
	updateQuery := `
		WITH updated_draft_deal_offer AS (
			UPDATE draft_deal_offers ddo
			SET confirmed = false
			FROM offers o
			WHERE o.id = ddo.offer_id
				AND ddo.draft_deal_id = $1
				AND o.author_id = $2
				AND ddo.confirmed = true
			RETURNING ddo.draft_deal_id
		),
		updated_draft_deal AS (
			UPDATE draft_deals dd
			SET updated_at = now()
			FROM updated_draft_deal_offer udo
			WHERE dd.id = udo.draft_deal_id
			RETURNING dd.id
		)
		SELECT EXISTS(
			SELECT 1
			FROM draft_deals dd
			WHERE dd.id = $1
		),
		EXISTS(
			SELECT 1
			FROM draft_deal_offers ddo
			JOIN offers o ON o.id = ddo.offer_id
			WHERE ddo.draft_deal_id = $1
				AND o.author_id = $2
		);`

	var draftExists bool
	var userInDraft bool
	err := exec.QueryRow(ctx, updateQuery, id, userID).Scan(&draftExists, &userInDraft)
	if err != nil {
		return fmt.Errorf("sql unconfirm draft: %w", err)
	}

	if !draftExists {
		return domain.ErrDraftNotFound
	}

	if !userInDraft {
		return domain.ErrUserNotInDraft
	}

	return nil
}

// ================================================================================
// GetConfirms
// ================================================================================

// GetConfirms returns confirm flags of all participants in a draft deal.
//
// No domain errors
func (r *Repository) GetConfirms(ctx context.Context, exec db.DB, draftID uuid.UUID) ([]htypes.UserConfirmed, error) {
	query := `
		SELECT dd.author_id, ddo.confirmed
		FROM draft_deal_offers ddo
				 JOIN draft_deals dd
					  ON dd.id = ddo.draft_deal_id
		WHERE ddo.draft_deal_id = $1;`

	rows, err := exec.Query(ctx, query, draftID)
	if err != nil {
		return nil, fmt.Errorf("sql check authors: %w", err)
	}
	defer rows.Close()

	var users []htypes.UserConfirmed
	for rows.Next() {
		var uID *uuid.UUID
		var confirmed *bool

		if err = rows.Scan(&uID, &confirmed); err != nil {
			return nil, fmt.Errorf("row scan: %w", err)
		}

		if uID == nil || confirmed == nil {
			return nil, fmt.Errorf("row scan: null fields")
		}

		users = append(users, htypes.UserConfirmed{
			UserID:    *uID,
			Confirmed: *confirmed,
		})
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows: %w", err)
	}

	return users, nil
}

// ================================================================================
// DeleteDraft
// ================================================================================

// DeleteDraft deletes a draft deal by its ID.
//
// Errors:
//   - domain.ErrDraftNotFound
//   - SQL errors are wrapped.
func (r *Repository) DeleteDraft(ctx context.Context, exec db.DB, id uuid.UUID) error {
	query := `
		DELETE FROM draft_deal_offers ddo
		WHERE draft_deal_id = $1;`

	tags, err := exec.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("sql delete draft: %w", err)
	}
	if tags.RowsAffected() == 0 {
		return domain.ErrDraftNotFound
	}

	return nil
}
