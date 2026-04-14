package offer_reports

import (
	dealsusers "barter-port/contracts/kafka/messages/deals-users"
	"barter-port/internal/deals/domain"
	offer_report_outbox "barter-port/internal/deals/infrastructure/repository/offer-report-outbox"
	offer_reports "barter-port/internal/deals/infrastructure/repository/offer-reports"
	offersrepo "barter-port/internal/deals/infrastructure/repository/offers"
	"barter-port/pkg/authkit"
	"barter-port/pkg/db"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const penaltyDelta = -10

type Service struct {
	db          *pgxpool.Pool
	offersRepo  *offersrepo.Repository
	reportsRepo *offer_reports.Repository
	outboxRepo  *offer_report_outbox.Repository
	admin       *authkit.AdminChecker
	logger      *slog.Logger
}

func NewService(
	dbPool *pgxpool.Pool,
	offersRepo *offersrepo.Repository,
	reportsRepo *offer_reports.Repository,
	outboxRepo *offer_report_outbox.Repository,
	admin *authkit.AdminChecker,
	logger *slog.Logger,
) *Service {
	return &Service{
		db:          dbPool,
		offersRepo:  offersRepo,
		reportsRepo: reportsRepo,
		outboxRepo:  outboxRepo,
		admin:       admin,
		logger:      logger,
	}
}

// CreateReport creates a new report or adds a reporter message to an existing pending report.
// Returns (true, report, nil) if a new report was created, (false, report, nil) if added to existing.
//
// Domain errors:
//   - domain.ErrOfferNotFound
//   - domain.ErrSelfReport
//   - domain.ErrReporterAlreadyAttached
func (s *Service) CreateReport(
	ctx context.Context,
	reporterID uuid.UUID,
	offerID uuid.UUID,
	message string,
) (created bool, report domain.OfferReport, err error) {
	err = db.RunInTx(ctx, s.db, func(ctx context.Context, tx pgx.Tx) error {
		offer, err := s.offersRepo.GetOffer(ctx, tx, offerID)
		if err != nil {
			return err
		}

		if offer.AuthorId == reporterID {
			return domain.ErrSelfReport
		}

		pendingReport, err := s.reportsRepo.GetPendingReportForOffer(ctx, tx, offerID)
		if err != nil {
			return fmt.Errorf("get pending report: %w", err)
		}

		now := time.Now()

		if pendingReport == nil {
			// Create new report
			newReport := domain.OfferReport{
				ID:            uuid.New(),
				OfferID:       offerID,
				OfferAuthorID: offer.AuthorId,
				Status:        domain.OfferReportStatusPending,
				CreatedAt:     now,
			}
			if err = s.reportsRepo.CreateReport(ctx, tx, newReport); err != nil {
				return fmt.Errorf("create report: %w", err)
			}

			if err = s.reportsRepo.AddReporterMessage(ctx, tx, domain.OfferReportMessage{
				OfferReportID: newReport.ID,
				AuthorID:      reporterID,
				Message:       message,
			}); err != nil {
				return err
			}

			if err = s.reportsRepo.BlockOfferModification(ctx, tx, offerID, now); err != nil {
				return fmt.Errorf("block offer modification: %w", err)
			}

			report = newReport
			created = true
		} else {
			// Add to existing pending report
			if err = s.reportsRepo.AddReporterMessage(ctx, tx, domain.OfferReportMessage{
				OfferReportID: pendingReport.ID,
				AuthorID:      reporterID,
				Message:       message,
			}); err != nil {
				return err
			}

			report = *pendingReport
			created = false
		}

		return nil
	})
	return created, report, err
}

// GetOfferReports returns all reports for an offer. Only the offer author or admin may call this.
//
// Domain errors:
//   - domain.ErrOfferNotFound
//   - domain.ErrForbidden
func (s *Service) GetOfferReports(ctx context.Context, userID uuid.UUID, offerID uuid.UUID) (domain.Offer, []domain.OfferReport, map[uuid.UUID][]domain.OfferReportMessage, error) {
	offer, err := s.offersRepo.GetOfferByID(ctx, offerID)
	if err != nil {
		return domain.Offer{}, nil, nil, err
	}

	if offer.AuthorId != userID {
		isAdmin, adminErr := s.admin.IsAdmin(ctx, userID)
		if adminErr != nil {
			return domain.Offer{}, nil, nil, fmt.Errorf("check admin: %w", adminErr)
		}
		if !isAdmin {
			return domain.Offer{}, nil, nil, domain.ErrForbidden
		}
	}

	reports, err := s.reportsRepo.GetReportsForOffer(ctx, s.db, offerID)
	if err != nil {
		return domain.Offer{}, nil, nil, err
	}

	reportIDs := make([]uuid.UUID, len(reports))
	for i, r := range reports {
		reportIDs[i] = r.ID
	}

	messagesByReport, err := s.reportsRepo.GetReportMessagesForOfferReports(ctx, s.db, reportIDs)
	if err != nil {
		return domain.Offer{}, nil, nil, err
	}

	return *offer, reports, messagesByReport, nil
}

// ListAdminReports returns all reports, optionally filtered by status. Admin only.
//
// Domain errors:
//   - domain.ErrAdminOnly
func (s *Service) ListAdminReports(ctx context.Context, adminID uuid.UUID, status *domain.OfferReportStatus) ([]domain.OfferReport, error) {
	isAdmin, err := s.admin.IsAdmin(ctx, adminID)
	if err != nil {
		return nil, fmt.Errorf("check admin: %w", err)
	}
	if !isAdmin {
		return nil, domain.ErrAdminOnly
	}

	return s.reportsRepo.ListReports(ctx, s.db, status)
}

// GetAdminReportDetails returns a report with its offer and messages. Admin only.
//
// Domain errors:
//   - domain.ErrAdminOnly
//   - domain.ErrReportNotFound
func (s *Service) GetAdminReportDetails(ctx context.Context, adminID uuid.UUID, reportID uuid.UUID) (domain.OfferReport, *domain.Offer, []domain.OfferReportMessage, error) {
	isAdmin, err := s.admin.IsAdmin(ctx, adminID)
	if err != nil {
		return domain.OfferReport{}, nil, nil, fmt.Errorf("check admin: %w", err)
	}
	if !isAdmin {
		return domain.OfferReport{}, nil, nil, domain.ErrAdminOnly
	}

	report, err := s.reportsRepo.GetReportByID(ctx, s.db, reportID)
	if err != nil {
		return domain.OfferReport{}, nil, nil, err
	}

	// Get offer directly (may be hidden — admin can see it)
	offer, err := s.offersRepo.GetOffer(ctx, s.db, report.OfferID)
	if err != nil && !errors.Is(err, domain.ErrOfferNotFound) {
		return domain.OfferReport{}, nil, nil, err
	}

	messages, err := s.reportsRepo.GetReportMessages(ctx, s.db, reportID)
	if err != nil {
		return domain.OfferReport{}, nil, nil, err
	}

	return *report, offer, messages, nil
}

// ResolveReport resolves a report. Admin only.
// If accepted: hides the offer and queues a reputation penalty via outbox.
// Always: unblocks offer modification and sets review fields.
//
// Domain errors:
//   - domain.ErrAdminOnly
//   - domain.ErrReportNotFound
//   - domain.ErrAlreadyReviewed
func (s *Service) ResolveReport(
	ctx context.Context,
	adminID uuid.UUID,
	reportID uuid.UUID,
	accepted bool,
	comment *string,
) (domain.OfferReport, error) {
	isAdmin, err := s.admin.IsAdmin(ctx, adminID)
	if err != nil {
		return domain.OfferReport{}, fmt.Errorf("check admin: %w", err)
	}
	if !isAdmin {
		return domain.OfferReport{}, domain.ErrAdminOnly
	}

	var resolved domain.OfferReport

	err = db.RunInTx(ctx, s.db, func(ctx context.Context, tx pgx.Tx) error {
		report, err := s.reportsRepo.GetReportByID(ctx, tx, reportID)
		if err != nil {
			return err
		}
		if report.Status != domain.OfferReportStatusPending {
			return domain.ErrAlreadyReviewed
		}

		now := time.Now()
		newStatus := domain.OfferReportStatusRejected
		var appliedPenaltyDelta *int

		if accepted {
			newStatus = domain.OfferReportStatusAccepted
			appliedPenaltyDelta = new(penaltyDelta)

			if err = s.reportsRepo.HideOffer(ctx, tx, report.OfferID, "Accepted report: "+reportID.String(), now); err != nil {
				return fmt.Errorf("hide offer: %w", err)
			}

			outboxMsg := dealsusers.OfferReportPenaltyMessage{
				ID:         uuid.New(),
				ReportID:   reportID,
				OfferID:    report.OfferID,
				UserID:     report.OfferAuthorID,
				Delta:      penaltyDelta,
				ReviewedBy: adminID,
				CreatedAt:  now,
			}
			if err = s.outboxRepo.WriteOutboxMessage(ctx, tx, outboxMsg); err != nil {
				return fmt.Errorf("write outbox message: %w", err)
			}
		}

		if err = s.reportsRepo.UnblockOfferModification(ctx, tx, report.OfferID); err != nil {
			return fmt.Errorf("unblock offer modification: %w", err)
		}

		if err = s.reportsRepo.UpdateReportResolution(ctx, tx, reportID, adminID, newStatus, comment, appliedPenaltyDelta, now); err != nil {
			return err
		}

		resolved = *report
		resolved.Status = newStatus
		resolved.ReviewedAt = &now
		resolved.ReviewedBy = &adminID
		resolved.ResolutionComment = comment
		resolved.AppliedPenaltyDelta = appliedPenaltyDelta

		return nil
	})
	if err != nil {
		return domain.OfferReport{}, err
	}

	s.logger.Debug("offer report resolved",
		slog.String("report_id", reportID.String()),
		slog.Bool("accepted", accepted),
	)

	return resolved, nil
}
