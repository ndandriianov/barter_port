package deals_users

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

const OfferReportPenaltyMessageType = "deals.offer_report.penalty"
const DealFailureResponsibleMessageType = "deals.deal_failure.responsible"
const DealCompletionRewardMessageType = "deals.deal_completion.reward"
const ReviewCreationRewardMessageType = "deals.review_creation.reward"

var reputationSourceNamespace = uuid.MustParse("2df40a4b-1846-43af-bdbb-61af8dcb23f8")

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

func BuildDealCompletionRewardSourceID(dealID, userID uuid.UUID) uuid.UUID {
	return uuid.NewSHA1(reputationSourceNamespace, []byte(fmt.Sprintf("deal-completion:%s:%s", dealID, userID)))
}

func BuildReviewCreationRewardSourceID(
	dealID uuid.UUID,
	itemID, offerID *uuid.UUID,
	authorID, providerID uuid.UUID,
) uuid.UUID {
	return uuid.NewSHA1(reputationSourceNamespace, []byte(fmt.Sprintf(
		"review-creation:%s:%s:%s:%s:%s",
		dealID,
		uuidOrNil(offerID),
		uuidOrNil(itemID),
		authorID,
		providerID,
	)))
}

func uuidOrNil(id *uuid.UUID) uuid.UUID {
	if id == nil {
		return uuid.Nil
	}
	return *id
}
