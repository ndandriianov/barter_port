package deals

import (
	"barter-port/contracts/openapi/deals/types"
	"barter-port/internal/deals/domain"
	"barter-port/internal/deals/domain/enums"
	"barter-port/internal/deals/domain/htypes"
	"barter-port/pkg/db"
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// ================================================================================
// CREATE DEAL ITEM REVIEW
// ================================================================================

// CreateDealItemReview creates a review for a deal item. The review target (offer, item, or both)
// is determined automatically by the item's offer_id and updated_at fields.
//
// Domain errors:
//   - domain.ErrDealNotFound
//   - domain.ErrInvalidDealStatus — deal is not Completed
//   - domain.ErrItemNotFound
//   - domain.ErrReceiverMissing — item has no receiver_id
//   - domain.ErrForbidden — current user is not the receiver
//   - domain.ErrProviderMissing — item has no provider_id
//   - domain.ErrSameProviderAndReceiver
//   - domain.ErrReviewAlreadyExists — duplicate review for this context
func (s *Service) CreateDealItemReview(
	ctx context.Context,
	userID, dealID, itemID uuid.UUID,
	rating int,
	comment *string,
) (domain.Review, error) {
	var review domain.Review

	err := db.RunInTx(ctx, s.db, func(ctx context.Context, tx pgx.Tx) error {
		deal, err := s.dealsRepository.GetDealByID(ctx, tx, dealID)
		if err != nil {
			return err
		}
		if deal.Status != enums.DealStatusCompleted {
			return domain.ErrInvalidDealStatus
		}

		item, err := s.dealsRepository.GetItemForReview(ctx, tx, dealID, itemID)
		if err != nil {
			return err
		}

		if item.ReceiverID == nil {
			return domain.ErrReceiverMissing
		}
		if *item.ReceiverID != userID {
			return domain.ErrForbidden
		}
		if item.ProviderID == nil {
			return domain.ErrProviderMissing
		}
		if *item.ProviderID == *item.ReceiverID {
			return domain.ErrSameProviderAndReceiver
		}

		offerID, storedItemID := determineReviewContext(item)

		review, err = s.dealsRepository.CreateReview(
			ctx, tx,
			dealID, userID, *item.ProviderID,
			storedItemID, offerID,
			rating, comment,
		)
		return err
	})
	if err != nil {
		return domain.Review{}, err
	}
	return review, nil
}

// ================================================================================
// GET REVIEW BY ID
// ================================================================================

// GetReviewByID returns a review by its ID.
//
// Domain errors:
//   - domain.ErrReviewNotFound
func (s *Service) GetReviewByID(ctx context.Context, reviewID uuid.UUID) (domain.Review, error) {
	return s.dealsRepository.GetReviewByID(ctx, s.db, reviewID)
}

// ================================================================================
// UPDATE REVIEW
// ================================================================================

// UpdateReview updates the rating and/or comment of a review.
// Only the author can update, and only while the deal is in Completed status.
//
// Domain errors:
//   - domain.ErrReviewNotFound
//   - domain.ErrForbidden — not the author
//   - domain.ErrDealNotFound
//   - domain.ErrInvalidDealStatus — deal is not Completed
func (s *Service) UpdateReview(
	ctx context.Context,
	userID, reviewID uuid.UUID,
	rating *int,
	comment *string,
) (domain.Review, error) {
	var review domain.Review

	err := db.RunInTx(ctx, s.db, func(ctx context.Context, tx pgx.Tx) error {
		existing, err := s.dealsRepository.GetReviewByID(ctx, tx, reviewID)
		if err != nil {
			return err
		}
		if existing.AuthorID != userID {
			return domain.ErrForbidden
		}

		deal, err := s.dealsRepository.GetDealByID(ctx, tx, existing.DealID)
		if err != nil {
			return err
		}
		if deal.Status != enums.DealStatusCompleted {
			return domain.ErrInvalidDealStatus
		}

		review, err = s.dealsRepository.UpdateReview(ctx, tx, reviewID, rating, comment)
		return err
	})
	if err != nil {
		return domain.Review{}, err
	}
	return review, nil
}

// ================================================================================
// DELETE REVIEW
// ================================================================================

// DeleteReview deletes a review.
// Only the author can delete, and only while the deal is in Completed status.
//
// Domain errors:
//   - domain.ErrReviewNotFound
//   - domain.ErrForbidden — not the author
//   - domain.ErrDealNotFound
//   - domain.ErrInvalidDealStatus — deal is not Completed
func (s *Service) DeleteReview(ctx context.Context, userID, reviewID uuid.UUID) error {
	return db.RunInTx(ctx, s.db, func(ctx context.Context, tx pgx.Tx) error {
		existing, err := s.dealsRepository.GetReviewByID(ctx, tx, reviewID)
		if err != nil {
			return err
		}
		if existing.AuthorID != userID {
			return domain.ErrForbidden
		}

		deal, err := s.dealsRepository.GetDealByID(ctx, tx, existing.DealID)
		if err != nil {
			return err
		}
		if deal.Status != enums.DealStatusCompleted {
			return domain.ErrInvalidDealStatus
		}

		return s.dealsRepository.DeleteReview(ctx, tx, reviewID)
	})
}

// ================================================================================
// GET DEAL REVIEWS
// ================================================================================

// GetDealReviews returns all reviews for a deal, ordered newest first.
//
// Domain errors:
//   - domain.ErrDealNotFound
func (s *Service) GetDealReviews(ctx context.Context, dealID uuid.UUID) ([]domain.Review, error) {
	var deal domain.Deal
	err := db.RunInTx(ctx, s.db, func(ctx context.Context, tx pgx.Tx) error {
		var err error
		deal, err = s.dealsRepository.GetDealByID(ctx, tx, dealID)
		return err
	})
	if err != nil {
		return nil, err
	}
	_ = deal
	return s.dealsRepository.GetReviewsByDealID(ctx, s.db, dealID)
}

// ================================================================================
// GET DEAL PENDING REVIEWS
// ================================================================================

// GetDealPendingReviews returns all pending review contexts for the current user in a deal.
// Only items where the user is the receiver are included.
//
// Domain errors:
//   - domain.ErrDealNotFound
//   - domain.ErrForbidden — user is not a participant of the deal
//   - domain.ErrInvalidDealStatus — deal is not Completed
func (s *Service) GetDealPendingReviews(
	ctx context.Context,
	userID, dealID uuid.UUID,
) ([]htypes.PendingReview, error) {
	var deal domain.Deal
	err := db.RunInTx(ctx, s.db, func(ctx context.Context, tx pgx.Tx) error {
		var err error
		deal, err = s.dealsRepository.GetDealByID(ctx, tx, dealID)
		return err
	})
	if err != nil {
		return nil, err
	}

	if !containsUserID(deal.Participants, userID) {
		return nil, domain.ErrForbidden
	}
	if deal.Status != enums.DealStatusCompleted {
		return nil, domain.ErrInvalidDealStatus
	}

	return s.dealsRepository.GetPendingReviewsForParticipant(ctx, s.db, dealID, userID)
}

// ================================================================================
// GET DEAL ITEM REVIEW ELIGIBILITY
// ================================================================================

// GetDealItemReviewEligibility returns the review eligibility for a specific item in a deal
// for the current user. Always returns a result (never 404 for eligibility itself),
// but returns 404 if the deal or item does not exist.
//
// Domain errors:
//   - domain.ErrDealNotFound
//   - domain.ErrItemNotFound
func (s *Service) GetDealItemReviewEligibility(
	ctx context.Context,
	userID, dealID, itemID uuid.UUID,
) (htypes.ReviewEligibility, error) {
	var deal domain.Deal
	err := db.RunInTx(ctx, s.db, func(ctx context.Context, tx pgx.Tx) error {
		var err error
		deal, err = s.dealsRepository.GetDealByID(ctx, tx, dealID)
		return err
	})
	if err != nil {
		return htypes.ReviewEligibility{}, err
	}

	item, err := s.dealsRepository.GetItemForReview(ctx, s.db, dealID, itemID)
	if err != nil {
		return htypes.ReviewEligibility{}, err
	}

	offerID, storedItemID := determineReviewContext(item)

	e := htypes.ReviewEligibility{
		ContextType: reviewContextType(item),
		ProviderID:  item.ProviderID,
		OfferID:     offerID,
		ItemID:      storedItemID,
		DealID:      dealID,
	}

	// Evaluate eligibility conditions in order per spec
	if deal.Status != enums.DealStatusCompleted {
		r := types.DealNotCompleted
		e.Reason = &r
		return e, nil
	}
	if item.ReceiverID == nil {
		r := types.ReceiverMissing
		e.Reason = &r
		return e, nil
	}
	if *item.ReceiverID != userID {
		r := types.ForbiddenNotReceiver
		e.Reason = &r
		return e, nil
	}
	if item.ProviderID == nil {
		r := types.ProviderMissing
		e.Reason = &r
		return e, nil
	}
	if *item.ProviderID == *item.ReceiverID {
		r := types.SameProviderAndReceiver
		e.Reason = &r
		return e, nil
	}

	exists, err := s.dealsRepository.ReviewExistsForUser(
		ctx, s.db, dealID, userID, storedItemID, offerID, *item.ProviderID,
	)
	if err != nil {
		return htypes.ReviewEligibility{}, fmt.Errorf("check review exists: %w", err)
	}
	if exists {
		r := types.AlreadyReviewed
		e.Reason = &r
		return e, nil
	}

	e.CanCreate = true
	return e, nil
}

// ================================================================================
// GET DEAL ITEM REVIEWS
// ================================================================================

// GetDealItemReviews returns all reviews for a specific item in a deal, ordered newest first.
//
// Domain errors:
//   - domain.ErrDealNotFound
//   - domain.ErrItemNotFound
func (s *Service) GetDealItemReviews(
	ctx context.Context,
	dealID, itemID uuid.UUID,
) ([]domain.Review, error) {
	var deal domain.Deal
	err := db.RunInTx(ctx, s.db, func(ctx context.Context, tx pgx.Tx) error {
		var err error
		deal, err = s.dealsRepository.GetDealByID(ctx, tx, dealID)
		return err
	})
	if err != nil {
		return nil, err
	}
	_ = deal

	if _, err = s.dealsRepository.GetItemForReview(ctx, s.db, dealID, itemID); err != nil {
		return nil, err
	}

	return s.dealsRepository.GetReviewsByDealItemID(ctx, s.db, dealID, itemID)
}

// ================================================================================
// GET OFFER REVIEWS
// ================================================================================

// GetOfferReviews returns all reviews for a specific offer, ordered newest first.
// Only reviews where offer_id is set are included.
//
// Domain errors:
//   - domain.ErrOfferNotFound
func (s *Service) GetOfferReviews(ctx context.Context, offerID uuid.UUID) ([]domain.Review, error) {
	if _, err := s.offersRepository.GetOffer(ctx, s.db, offerID); err != nil {
		return nil, err
	}
	return s.dealsRepository.GetReviewsByOfferID(ctx, s.db, offerID)
}

// ================================================================================
// GET OFFER REVIEWS SUMMARY
// ================================================================================

// GetOfferReviewsSummary returns aggregated review statistics for a specific offer.
// Returns zero-value summary when no reviews exist.
//
// Domain errors:
//   - domain.ErrOfferNotFound
func (s *Service) GetOfferReviewsSummary(
	ctx context.Context,
	offerID uuid.UUID,
) (htypes.ReviewSummary, error) {
	if _, err := s.offersRepository.GetOffer(ctx, s.db, offerID); err != nil {
		return htypes.ReviewSummary{}, err
	}
	return s.dealsRepository.GetReviewsSummaryByOfferID(ctx, s.db, offerID)
}

// ================================================================================
// GET PROVIDER REVIEWS
// ================================================================================

// GetProviderReviews returns all reviews for a specific provider, ordered newest first.
// No existence check is performed on the provider.
func (s *Service) GetProviderReviews(ctx context.Context, providerID uuid.UUID) ([]domain.Review, error) {
	return s.dealsRepository.GetReviewsByProviderID(ctx, s.db, providerID)
}

// ================================================================================
// GET PROVIDER REVIEWS SUMMARY
// ================================================================================

// GetProviderReviewsSummary returns aggregated review statistics for a specific provider.
// No existence check is performed on the provider.
func (s *Service) GetProviderReviewsSummary(
	ctx context.Context,
	providerID uuid.UUID,
) (htypes.ReviewSummary, error) {
	return s.dealsRepository.GetReviewsSummaryByProviderID(ctx, s.db, providerID)
}

// ================================================================================
// GET AUTHOR REVIEWS
// ================================================================================

// GetAuthorReviews returns all reviews written by a specific author, ordered newest first.
// No existence check is performed on the author.
func (s *Service) GetAuthorReviews(ctx context.Context, authorID uuid.UUID) ([]domain.Review, error) {
	return s.dealsRepository.GetReviewsByAuthorID(ctx, s.db, authorID)
}

// ================================================================================
// HELPERS
// ================================================================================

// determineReviewContext derives which IDs to store based on item fields:
//   - offer_id IS NULL → item-only (store item_id only)
//   - offer_id NOT NULL AND updated_at IS NULL → offer-only (store offer_id only)
//   - offer_id NOT NULL AND updated_at NOT NULL → offer+item (store both)
func determineReviewContext(item domain.Item) (offerID *uuid.UUID, storedItemID *uuid.UUID) {
	if item.OfferID == nil {
		return nil, &item.ID // item-only
	}
	if item.UpdatedAt == nil {
		return item.OfferID, nil // offer-only
	}
	return item.OfferID, &item.ID // offer+item
}

// reviewContextType returns the ReviewContextType for an item.
func reviewContextType(item domain.Item) types.ReviewContextType {
	if item.OfferID == nil {
		return types.ItemOnly
	}
	if item.UpdatedAt == nil {
		return types.OfferOnly
	}
	return types.OfferItem
}
