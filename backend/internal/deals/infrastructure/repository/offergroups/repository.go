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

	for _, unit := range units {
		unitID := uuid.New()
		if _, err := exec.Exec(ctx, `
			INSERT INTO offer_group_units (id, offer_group_id)
			VALUES ($1, $2)
		`, unitID, groupID); err != nil {
			return uuid.Nil, fmt.Errorf("insert offer group unit: %w", err)
		}

		for _, offerID := range unit.OfferIDs {
			if _, err := exec.Exec(ctx, `
				INSERT INTO unit_offers (unit_id, offer_id)
				VALUES ($1, $2)
			`, unitID, offerID); err != nil {
				return uuid.Nil, fmt.Errorf("insert unit offer: %w", err)
			}
		}
	}

	return groupID, nil
}

func (r *Repository) GetOfferGroupByID(ctx context.Context, id uuid.UUID) (domain.OfferGroup, error) {
	return r.getOfferGroups(ctx, r.db, &id)
}

func (r *Repository) ListOfferGroups(ctx context.Context) ([]domain.OfferGroup, error) {
	return r.listOfferGroups(ctx, r.db, nil)
}

func (r *Repository) getOfferGroups(ctx context.Context, exec db.DB, filterID *uuid.UUID) (domain.OfferGroup, error) {
	groups, err := r.listOfferGroups(ctx, exec, filterID)
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
) ([]domain.OfferGroup, error) {
	query := `
		SELECT
			og.id,
			og.name,
			og.description,
			ogu.id,
			o.id,
			o.author_id,
			o.name,
			o.type,
			o.action,
			o.description,
			o.created_at,
			o.views
		FROM offer_groups og
		LEFT JOIN offer_group_units ogu ON ogu.offer_group_id = og.id
		LEFT JOIN unit_offers uo ON uo.unit_id = ogu.id
		LEFT JOIN offers o ON o.id = uo.offer_id
	`

	args := make([]any, 0, 1)
	if filterID != nil {
		query += ` WHERE og.id = $1`
		args = append(args, *filterID)
	}

	query += ` ORDER BY og.id, ogu.id, o.id`

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
			unitID           *uuid.UUID
			offerID          *uuid.UUID
			authorID         *uuid.UUID
			offerName        *string
			offerType        *enums.ItemType
			offerAction      *enums.OfferAction
			offerDescription *string
			offerCreatedAt   *time.Time
			offerViews       *int
		)

		if err = rows.Scan(
			&groupID,
			&groupName,
			&groupDescription,
			&unitID,
			&offerID,
			&authorID,
			&offerName,
			&offerType,
			&offerAction,
			&offerDescription,
			&offerCreatedAt,
			&offerViews,
		); err != nil {
			return nil, fmt.Errorf("scan offer group row: %w", err)
		}

		gIdx, ok := groupIndex[groupID]
		if !ok {
			result = append(result, domain.OfferGroup{
				ID:          groupID,
				Name:        groupName,
				Description: groupDescription,
				Units:       make([]domain.OfferGroupUnit, 0),
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
		})
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate offer groups rows: %w", err)
	}

	return result, nil
}

var _ db.DB = (*pgxpool.Pool)(nil)
var _ db.DB = (pgx.Tx)(nil)
