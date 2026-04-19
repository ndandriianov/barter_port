package deals_users

import (
	"time"

	"github.com/google/uuid"
)

const OfferReportPenaltyMessageType = "deals.offer_report.penalty"
const DealFailureResponsibleMessageType = "deals.deal_failure.responsible"

type ReputationMessage struct {
	ID         uuid.UUID `db:"id"`
	SourceType string    `db:"source_type"`
	SourceID   uuid.UUID `db:"source_id"`
	UserID     uuid.UUID `db:"user_id"`
	Delta      int       `db:"delta"`
	CreatedAt  time.Time `db:"created_at"`
	Comment    *string   `db:"comment"`
}

func (m ReputationMessage) GetKey() string {
	return m.SourceID.String()
}

func (m ReputationMessage) GetCreatedAt() time.Time {
	return m.CreatedAt
}

func (m ReputationMessage) GetMessageType() string {
	return m.SourceType
}
