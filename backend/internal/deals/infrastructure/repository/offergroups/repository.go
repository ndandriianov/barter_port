package offergroups

import (
	"barter-port/internal/deals/domain"
	"barter-port/internal/deals/domain/enums"
	"barter-port/pkg/db"
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateOfferGroup(
	ctx context.Context,
	exec db.DB,
	name string,
	description *string,
	units []domain.OfferGroupUnitCreateInput,
) (uuid.UUID, error) {
	groupID := uuid.New()

	query := `
		INSERT INTO offer_groups (id, name, description)
		VALUES ($1, $2, $3)
	`
	if _, err := exec.Exec(ctx, query, groupID, name, description); err != nil {
		return uuid.Nil, fmt.Errorf("insert offer group: %w", err)
	}

	for unitIdx, unit := range units {
		unitID := uuid.New()
		if _, err := exec.Exec(ctx, `
			INSERT INTO offer_group_units (id, offer_group_id, position)
			VALUES ($1, $2, $3)
		`, unitID, groupID, unitIdx); err != nil {
			return uuid.Nil, fmt.Errorf("insert offer group unit: %w", err)
		}

		for offerIdx, offerID := range unit.OfferIDs {
			if _, err := exec.Exec(ctx, `
				INSERT INTO unit_offers (unit_id, offer_id, position)
				VALUES ($1, $2, $3)
			`, unitID, offerID, offerIdx); err != nil {
				return uuid.Nil, fmt.Errorf("insert unit offer: %w", err)
			}
		}
	}

	return groupID, nil
}

func (r *Repository) GetOfferGroupByID(ctx context.Context, id uuid.UUID, currentUserID uuid.UUID) (domain.OfferGroup, error) {
	return r.getOfferGroups(ctx, r.db, &id, currentUserID)
}

func (r *Repository) ListOfferGroups(ctx context.Context, currentUserID uuid.UUID) ([]domain.OfferGroup, error) {
	return r.listOfferGroups(ctx, r.db, nil, currentUserID)
}

func (r *Repository) getOfferGroups(
	ctx context.Context,
	exec db.DB,
	filterID *uuid.UUID,
	currentUserID uuid.UUID,
) (domain.OfferGroup, error) {
	groups, err := r.listOfferGroups(ctx, exec, filterID, currentUserID)
	if err != nil {
		return domain.OfferGroup{}, err
	}
	if len(groups) == 0 {
		return domain.OfferGroup{}, domain.ErrOfferGroupNotFound
	}
	return groups[0], nil
}

func (r *Repository) listOfferGroups(
	ctx context.Context,
	exec db.DB,
	filterID *uuid.UUID,
	currentUserID uuid.UUID,
) ([]domain.OfferGroup, error) {
	query := `
		WITH offer_group_owners AS (
			SELECT
				og.id AS offer_group_id,
				CASE
					WHEN COUNT(DISTINCT o.author_id) = 1 THEN (array_agg(DISTINCT o.author_id))[1]
					ELSE NULL
				END AS owner_id
			FROM offer_groups og
			LEFT JOIN offer_group_units ogu ON ogu.offer_group_id = og.id
			LEFT JOIN unit_offers uo ON uo.unit_id = ogu.id
			LEFT JOIN offers o ON o.id = uo.offer_id
			GROUP BY og.id
		),
		offer_group_draft_counts AS (
			SELECT
				dd.offer_group_id,
				COUNT(*)::integer AS draft_deals_count
			FROM draft_deals dd
			WHERE dd.offer_group_id IS NOT NULL
			GROUP BY dd.offer_group_id
		)
		SELECT
			og.id,
			og.name,
			og.description,
			CASE
				WHEN ogo.owner_id = $1 THEN COALESCE(ogdc.draft_deals_count, 0)
				ELSE NULL
			END AS draft_deals_count,
			ogu.id,
			o.id,
			o.author_id,
			o.name,
			o.type,
			o.action,
			o.description,
			o.created_at,
			o.views,
			COALESCE((SELECT array_agg(op.id ORDER BY op.position) FROM offer_photos op WHERE op.offer_id = o.id), '{}'::uuid[]) AS photo_ids,
			COALESCE((SELECT array_agg(op.url ORDER BY op.position) FROM offer_photos op WHERE op.offer_id = o.id), '{}'::text[]) AS photo_urls
		FROM offer_groups og
		LEFT JOIN offer_group_owners ogo ON ogo.offer_group_id = og.id
		LEFT JOIN offer_group_draft_counts ogdc ON ogdc.offer_group_id = og.id
		LEFT JOIN offer_group_units ogu ON ogu.offer_group_id = og.id
		LEFT JOIN unit_offers uo ON uo.unit_id = ogu.id
		LEFT JOIN offers o ON o.id = uo.offer_id
	`

	args := make([]any, 0, 2)
	args = append(args, currentUserID)
	if filterID != nil {
		query += ` WHERE og.id = $2`
		args = append(args, *filterID)
	}

	query += ` ORDER BY og.id, ogu.position NULLS LAST, uo.position NULLS LAST, o.id`

	rows, err := exec.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query offer groups: %w", err)
	}
	defer rows.Close()

	result := make([]domain.OfferGroup, 0)
	groupIndex := make(map[uuid.UUID]int)
	unitIndex := make(map[uuid.UUID]int)

	for rows.Next() {
		var (
			groupID          uuid.UUID
			groupName        string
			groupDescription *string
			draftDealsCount  *int
			unitID           *uuid.UUID
			offerID          *uuid.UUID
			authorID         *uuid.UUID
			offerName        *string
			offerType        *enums.ItemType
			offerAction      *enums.OfferAction
			offerDescription *string
			offerCreatedAt   *time.Time
			offerViews       *int
			offerPhotoIDs    []uuid.UUID
			offerPhotoUrls   []string
		)

		if err = rows.Scan(
			&groupID,
			&groupName,
			&groupDescription,
			&draftDealsCount,
			&unitID,
			&offerID,
			&authorID,
			&offerName,
			&offerType,
			&offerAction,
			&offerDescription,
			&offerCreatedAt,
			&offerViews,
			&offerPhotoIDs,
			&offerPhotoUrls,
		); err != nil {
			return nil, fmt.Errorf("scan offer group row: %w", err)
		}

		gIdx, ok := groupIndex[groupID]
		if !ok {
			result = append(result, domain.OfferGroup{
				ID:              groupID,
				Name:            groupName,
				Description:     groupDescription,
				DraftDealsCount: draftDealsCount,
				Units:           make([]domain.OfferGroupUnit, 0),
			})
			gIdx = len(result) - 1
			groupIndex[groupID] = gIdx
		}

		if unitID == nil {
			continue
		}

		uIdx, ok := unitIndex[*unitID]
		if !ok {
			result[gIdx].Units = append(result[gIdx].Units, domain.OfferGroupUnit{
				ID:     *unitID,
				Offers: make([]domain.Offer, 0),
			})
			uIdx = len(result[gIdx].Units) - 1
			unitIndex[*unitID] = uIdx
		}

		if offerID == nil || authorID == nil || offerName == nil || offerType == nil ||
			offerAction == nil || offerDescription == nil || offerCreatedAt == nil || offerViews == nil {
			continue
		}

		result[gIdx].Units[uIdx].Offers = append(result[gIdx].Units[uIdx].Offers, domain.Offer{
			ID:          *offerID,
			AuthorId:    *authorID,
			Name:        *offerName,
			Type:        *offerType,
			Action:      *offerAction,
			Description: *offerDescription,
			CreatedAt:   *offerCreatedAt,
			Views:       *offerViews,
			PhotoIds:    offerPhotoIDs,
			PhotoUrls:   offerPhotoUrls,
		})
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate offer groups rows: %w", err)
	}

	return result, nil
}

var _ db.DB = (*pgxpool.Pool)(nil)
var _ db.DB = (pgx.Tx)(nil)
