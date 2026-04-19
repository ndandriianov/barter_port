package deals_users

import (
	"time"

	"github.com/google/uuid"
)

const OfferReportPenaltyMessageType = "deals.offer_report.penalty"
const DealFailureResponsibleMessageType = "deals.deal_failure.responsible"

// OfferReportPenaltyMessage is the Kafka message sent from the deals service
// to the users service when an offer report is accepted (penalty applied).
type OfferReportPenaltyMessage struct {
	ID         uuid.UUID `json:"id"          db:"id"`
	ReportID   uuid.UUID `json:"report_id"   db:"report_id"`
	OfferID    uuid.UUID `json:"offer_id"    db:"offer_id"`
	UserID     uuid.UUID `json:"user_id"     db:"user_id"`
	Delta      int       `json:"delta"       db:"delta"`
	ReviewedBy uuid.UUID `json:"reviewed_by" db:"reviewed_by"`
	CreatedAt  time.Time `json:"created_at"  db:"created_at"`
}

func (m OfferReportPenaltyMessage) GetKey() string {
	return m.ReportID.String()
}

func (m OfferReportPenaltyMessage) GetCreatedAt() time.Time {
	return m.CreatedAt
}

func (m OfferReportPenaltyMessage) GetMessageType() string {
	return OfferReportPenaltyMessageType
}

type PenaltyMessage struct {
	ID         uuid.UUID `db:"id"`
	SourceType string    `db:"source_type"`
	SourceID   uuid.UUID `db:"source_id"`
	UserID     uuid.UUID `db:"user_id"`
	Delta      int       `db:"delta"`
	CreatedAt  time.Time `db:"created_at"`
	Comment    *string   `db:"comment"`
}

func (m PenaltyMessage) GetKey() string {
	return m.SourceID.String()
}

func (m PenaltyMessage) GetCreatedAt() time.Time {
	return m.CreatedAt
}

func (m PenaltyMessage) GetMessageType() string {
	return m.SourceType
}
