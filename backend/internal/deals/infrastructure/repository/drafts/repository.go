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
	offerGroupID *uuid.UUID,
) (uuid.UUID, error) {

	dealsQuery := `
		INSERT INTO draft_deals (author_id, name, description, offer_group_id)
		VALUES ($1, $2, $3, $4)
		RETURNING id;`

	var id uuid.UUID
	err := tx.QueryRow(ctx, dealsQuery, authorID, name, description, offerGroupID).Scan(&id)
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
// GetDraftsByAuthor
// ================================================================================

// GetDraftsByAuthor retrieves the IDs of draft deals created by a specific author.
//
// No domain errors
func (r *Repository) GetDraftsByAuthor(
	ctx context.Context,
	exec db.DB,
	authorID uuid.UUID,
	createdByMe bool,
) ([]htypes.DraftIDWithAuthorIDs, error) {
	var query string
	if createdByMe {
		query = `
			SELECT dd.id, dd.name, array_agg(DISTINCT o.author_id) AS participant_ids
			FROM draft_deals dd
			JOIN draft_deal_offers ddo ON dd.id = ddo.draft_deal_id
			JOIN offers o ON o.id = ddo.offer_id
			WHERE dd.author_id = $1
			GROUP BY dd.id;`
	} else {
		query = `
			SELECT dd.id, dd.name, array_agg(DISTINCT o.author_id) AS participant_ids
			FROM draft_deals dd
			JOIN draft_deal_offers ddo ON dd.id = ddo.draft_deal_id
			JOIN offers o ON o.id = ddo.offer_id
			WHERE EXISTS(
			    SELECT 1
			    FROM draft_deal_offers ddo2
			    JOIN offers o2 ON o2.id = ddo2.offer_id
			    WHERE o2.author_id = $1 AND ddo2.draft_deal_id = dd.id
			)
			GROUP BY dd.id;`
	}

	rows, err := exec.Query(ctx, query, authorID)
	if err != nil {
		return nil, fmt.Errorf("sql: %w", err)
	}
	defer rows.Close()

	var drafts []htypes.DraftIDWithAuthorIDs
	for rows.Next() {
		var id uuid.UUID
		var name *string
		var participantIDs []uuid.UUID

		err = rows.Scan(&id, &name, &participantIDs)
		if err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}

		drafts = append(drafts, htypes.DraftIDWithAuthorIDs{
			ID:             id,
			Name:           name,
			ParticipantIDs: participantIDs,
		})
	}

	return drafts, nil
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
		       d.offer_group_id,
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
		       COALESCE((SELECT array_agg(op.id ORDER BY op.position) FROM offer_photos op WHERE op.offer_id = i.id), '{}'::uuid[]) AS photo_ids,
		       COALESCE((SELECT array_agg(op.url ORDER BY op.position) FROM offer_photos op WHERE op.offer_id = i.id), '{}'::text[]) AS photo_urls,
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
		var offerPhotoIDs []uuid.UUID
		var offerPhotoUrls []string
		var offerQuantity *int
		var offerConfirmed *bool
		var offerGroupID *uuid.UUID

		if err = rows.Scan(
			&draft.ID,
			&draft.AuthorID,
			&draft.Name,
			&draft.Description,
			&offerGroupID,
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
			&offerPhotoIDs,
			&offerPhotoUrls,
			&offerQuantity,
			&offerConfirmed,
		); err != nil {
			return domain.Draft{}, fmt.Errorf("scan draft item: %w", err)
		}

		found = true
		draft.OfferGroupID = offerGroupID

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
				PhotoIds:    offerPhotoIDs,
				PhotoUrls:   offerPhotoUrls,
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
// GetParticipants
// ================================================================================

// GetParticipants returns the IDs of all users participating in a draft deal.
//
// No domain errors
func (r *Repository) GetParticipants(ctx context.Context, exec db.DB, draftID uuid.UUID) ([]uuid.UUID, error) {
	query := `
		SELECT array_agg(DISTINCT o.author_id)
		FROM draft_deal_offers ddo
		JOIN offers o ON o.id = ddo.offer_id
		WHERE ddo.draft_deal_id = $1;`

	var participants []uuid.UUID
	err := exec.QueryRow(ctx, query, draftID).Scan(&participants)
	if err != nil {
		return nil, fmt.Errorf("sql get participants: %w", err)
	}

	return participants, nil
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
		DELETE FROM draft_deals
		WHERE id = $1;`

	tags, err := exec.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("sql delete draft: %w", err)
	}
	if tags.RowsAffected() == 0 {
		return domain.ErrDraftNotFound
	}

	return nil
}

// ================================================================================
// GetDraftsCountForOfferID
// ================================================================================

func (r *Repository) GetDraftsCountForOfferID(ctx context.Context, exec db.DB, offerID uuid.UUID) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM draft_deal_offers
		WHERE offer_id = $1;`

	var count int
	err := exec.QueryRow(ctx, query, offerID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("sql get drafts count for offer: %w", err)
	}

	return count, nil
}
