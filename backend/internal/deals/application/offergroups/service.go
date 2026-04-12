package offergroups

import (
	userspb "barter-port/contracts/grpc/users/v1"
	dealssvc "barter-port/internal/deals/application/deals"
	"barter-port/internal/deals/domain"
	"barter-port/internal/deals/domain/enums"
	offergroupsrepo "barter-port/internal/deals/infrastructure/repository/offergroups"
	offersrepo "barter-port/internal/deals/infrastructure/repository/offers"
	"barter-port/pkg/db"
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	db           *pgxpool.Pool
	repo         *offergroupsrepo.Repository
	offersRepo   *offersrepo.Repository
	dealsService *dealssvc.Service
	usersClient  userspb.UsersServiceClient
}

func NewService(
	db *pgxpool.Pool,
	repo *offergroupsrepo.Repository,
	offersRepo *offersrepo.Repository,
	dealsService *dealssvc.Service,
	usersClient userspb.UsersServiceClient,
) *Service {
	return &Service{
		db:           db,
		repo:         repo,
		offersRepo:   offersRepo,
		dealsService: dealsService,
		usersClient:  usersClient,
	}
}

func (s *Service) CreateOfferGroup(
	ctx context.Context,
	userID uuid.UUID,
	name string,
	description *string,
	units []domain.OfferGroupUnitCreateInput,
) (domain.OfferGroup, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return domain.OfferGroup{}, domain.ErrInvalidOfferName
	}

	offerIDs, err := validateCreateUnits(units)
	if err != nil {
		return domain.OfferGroup{}, err
	}

	var groupID uuid.UUID
	err = db.RunInTx(ctx, s.db, func(ctx context.Context, tx pgx.Tx) error {
		offers, err := s.offersRepo.GetOffersByIDs(ctx, tx, offerIDs)
		if err != nil {
			return fmt.Errorf("get offers for offer group: %w", err)
		}
		if len(offers) != len(offerIDs) {
			return domain.ErrOfferNotFound
		}

		for _, offer := range offers {
			if offer.AuthorId != userID {
				return domain.ErrForbidden
			}
		}

		if err = validateUnitActions(units, offers); err != nil {
			return err
		}

		groupID, err = s.repo.CreateOfferGroup(ctx, tx, name, description, units)
		if err != nil {
			return fmt.Errorf("create offer group: %w", err)
		}
		return nil
	})
	if err != nil {
		return domain.OfferGroup{}, err
	}

	return s.GetOfferGroupByID(ctx, groupID)
}

func (s *Service) ListOfferGroups(ctx context.Context) ([]domain.OfferGroup, error) {
	items, err := s.repo.ListOfferGroups(ctx)
	if err != nil {
		return nil, err
	}

	return s.populateAuthorNames(ctx, items)
}

func (s *Service) GetOfferGroupByID(ctx context.Context, id uuid.UUID) (domain.OfferGroup, error) {
	item, err := s.repo.GetOfferGroupByID(ctx, id)
	if err != nil {
		return domain.OfferGroup{}, err
	}

	items, err := s.populateAuthorNames(ctx, []domain.OfferGroup{item})
	if err != nil {
		return domain.OfferGroup{}, err
	}

	return items[0], nil
}

func (s *Service) CreateDraftFromOfferGroup(
	ctx context.Context,
	offerGroupID uuid.UUID,
	userID uuid.UUID,
	name *string,
	description *string,
	selectedOfferIDs []uuid.UUID,
	responderOfferID *uuid.UUID,
) (uuid.UUID, error) {
	if len(selectedOfferIDs) == 0 {
		return uuid.Nil, domain.ErrInvalidOfferGroupSelect
	}

	selectedSet := make(map[uuid.UUID]struct{}, len(selectedOfferIDs))
	for _, id := range selectedOfferIDs {
		if _, ok := selectedSet[id]; ok {
			return uuid.Nil, domain.ErrInvalidOfferGroupSelect
		}
		selectedSet[id] = struct{}{}
	}

	group, err := s.repo.GetOfferGroupByID(ctx, offerGroupID)
	if err != nil {
		return uuid.Nil, err
	}

	if len(group.Units) == 0 || len(selectedOfferIDs) != len(group.Units) {
		return uuid.Nil, domain.ErrInvalidOfferGroupSelect
	}

	uniformAction, hasUniformAction := getUniformGroupAction(group)

	selectedOffers := make([]domain.OfferIDAndInfo, 0, len(selectedOfferIDs))
	for _, unit := range group.Units {
		matches := 0
		var matchedOfferID uuid.UUID

		for _, offer := range unit.Offers {
			if _, ok := selectedSet[offer.ID]; ok {
				matches++
				matchedOfferID = offer.ID
			}
		}

		if matches != 1 {
			return uuid.Nil, domain.ErrInvalidOfferGroupSelect
		}

		selectedOffers = append(selectedOffers, domain.OfferIDAndInfo{
			ID: matchedOfferID,
			Info: domain.OfferInfo{
				Quantity: 1,
			},
		})
	}

	if responderOfferID != nil {
		if _, alreadySelected := selectedSet[*responderOfferID]; alreadySelected {
			return uuid.Nil, domain.ErrInvalidOfferGroupSelect
		}

		offer, err := s.offersRepo.GetOfferByID(ctx, *responderOfferID)
		if err != nil {
			return uuid.Nil, err
		}
		if offer.AuthorId != userID {
			return uuid.Nil, domain.ErrForbidden
		}
		if hasUniformAction && offer.Action != uniformAction {
			return uuid.Nil, domain.ErrOfferGroupResponderOfferAction
		}

		selectedOffers = append(selectedOffers, domain.OfferIDAndInfo{
			ID: offer.ID,
			Info: domain.OfferInfo{
				Quantity: 1,
			},
		})
	} else if hasUniformAction {
		return uuid.Nil, domain.ErrOfferGroupResponderOfferRequired
	}

	return s.dealsService.CreateDraft(ctx, userID, name, description, selectedOffers)
}

func validateCreateUnits(units []domain.OfferGroupUnitCreateInput) ([]uuid.UUID, error) {
	if len(units) == 0 {
		return nil, domain.ErrNoOfferGroupUnits
	}

	offerIDs := make([]uuid.UUID, 0)
	seen := make(map[uuid.UUID]struct{})

	for _, unit := range units {
		if len(unit.OfferIDs) == 0 {
			return nil, domain.ErrEmptyOfferGroupUnit
		}

		for _, offerID := range unit.OfferIDs {
			if _, ok := seen[offerID]; ok {
				return nil, domain.ErrDuplicateOfferInGroup
			}
			seen[offerID] = struct{}{}
			offerIDs = append(offerIDs, offerID)
		}
	}

	return offerIDs, nil
}

func validateUnitActions(units []domain.OfferGroupUnitCreateInput, offers []domain.Offer) error {
	offersByID := make(map[uuid.UUID]domain.Offer, len(offers))
	for _, offer := range offers {
		offersByID[offer.ID] = offer
	}

	for _, unit := range units {
		var expectedAction *string

		for _, offerID := range unit.OfferIDs {
			offer, ok := offersByID[offerID]
			if !ok {
				return domain.ErrOfferNotFound
			}

			action := offer.Action.String()
			if expectedAction == nil {
				expectedAction = &action
				continue
			}

			if *expectedAction != action {
				return domain.ErrMixedOfferActionsInUnit
			}
		}
	}

	return nil
}

func getUniformGroupAction(group domain.OfferGroup) (enums.OfferAction, bool) {
	actions := make(map[enums.OfferAction]struct{})
	var zeroAction enums.OfferAction

	for _, unit := range group.Units {
		if len(unit.Offers) == 0 {
			continue
		}
		actions[unit.Offers[0].Action] = struct{}{}
	}

	if len(actions) != 1 {
		return zeroAction, false
	}

	for action := range actions {
		return action, true
	}

	return zeroAction, false
}

func (s *Service) populateAuthorNames(ctx context.Context, items []domain.OfferGroup) ([]domain.OfferGroup, error) {
	idsSet := make(map[string]struct{})
	ids := make([]string, 0)

	for _, group := range items {
		for _, unit := range group.Units {
			for _, offer := range unit.Offers {
				id := offer.AuthorId.String()
				if _, ok := idsSet[id]; ok {
					continue
				}
				idsSet[id] = struct{}{}
				ids = append(ids, id)
			}
		}
	}

	if len(ids) == 0 {
		return items, nil
	}

	response, err := s.usersClient.GetUsersWithInfo(ctx, &userspb.GetUsersWithInfoRequest{Ids: ids})
	if err != nil {
		return nil, fmt.Errorf("failed to get author names: %w", err)
	}

	namesByID := make(map[string]string, len(response.Users))
	for _, user := range response.Users {
		if user == nil {
			continue
		}
		namesByID[user.Id] = user.Name
	}

	for gi := range items {
		for ui := range items[gi].Units {
			for oi := range items[gi].Units[ui].Offers {
				offer := &items[gi].Units[ui].Offers[oi]
				if name, ok := namesByID[offer.AuthorId.String()]; ok {
					offer.AuthorName = &name
				}
			}
		}
	}

	return items, nil
}
