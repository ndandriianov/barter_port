package domain

import (
	"barter-port/contracts/openapi/deals/types"
	"time"

	"github.com/google/uuid"
)

type OfferReportStatus string

const (
	OfferReportStatusPending  OfferReportStatus = "Pending"
	OfferReportStatusAccepted OfferReportStatus = "Accepted"
	OfferReportStatusRejected OfferReportStatus = "Rejected"
)

type OfferReport struct {
	ID                  uuid.UUID         `db:"id"`
	OfferID             uuid.UUID         `db:"offer_id"`
	OfferAuthorID       uuid.UUID         `db:"offer_author_id"`
	Status              OfferReportStatus `db:"status"`
	CreatedAt           time.Time         `db:"created_at"`
	ReviewedAt          *time.Time        `db:"reviewed_at"`
	ReviewedBy          *uuid.UUID        `db:"reviewed_by"`
	ResolutionComment   *string           `db:"resolution_comment"`
	AppliedPenaltyDelta *int              `db:"applied_penalty_delta"`
}

func (r *OfferReport) ToDto() types.OfferReport {
	return types.OfferReport{
		Id:                  r.ID,
		OfferId:             r.OfferID,
		OfferAuthorId:       r.OfferAuthorID,
		Status:              types.OfferReportStatus(r.Status),
		CreatedAt:           r.CreatedAt,
		ReviewedAt:          r.ReviewedAt,
		ReviewedBy:          r.ReviewedBy,
		ResolutionComment:   r.ResolutionComment,
		AppliedPenaltyDelta: r.AppliedPenaltyDelta,
	}
}

type OfferReportMessage struct {
	OfferReportID uuid.UUID `db:"offer_report_id"`
	AuthorID      uuid.UUID `db:"author_id"`
	MessageID     uuid.UUID `db:"message_id"`
}

func (m *OfferReportMessage) ToDto() types.OfferReportMessage {
	return types.OfferReportMessage{
		OfferReportId: m.OfferReportID,
		AuthorId:      m.AuthorID,
		MessageId:     m.MessageID,
	}
}
