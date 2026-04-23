package seed_demo

import (
	dealtypes "barter-port/contracts/openapi/deals/types"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type SeedClient struct {
	BaseURL      string
	SMTP4DevURL  string
	SMTP4DevUser string
	SMTP4DevPass string
	HttpClient   *http.Client
	PollInterval time.Duration
}

type authStatusResponse struct {
	Status string `json:"status"`
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type registerResponse struct {
	UserID uuid.UUID `json:"userId"`
	Email  string    `json:"email"`
}

type loginResponse struct {
	AccessToken string `json:"accessToken"`
}

type verifyEmailRequest struct {
	Token string `json:"token"`
}

type refreshResponse struct {
	AccessToken string `json:"access_token"`
}

type offerGroupRequest struct {
	Name        *string                 `json:"name,omitempty"`
	Description *string                 `json:"description,omitempty"`
	Units       []offerGroupUnitRequest `json:"units"`
}

type offerGroupUnitRequest struct {
	Offers []offerGroupOfferRef `json:"offers"`
}

type offerGroupOfferRef struct {
	OfferID uuid.UUID `json:"offerId"`
}

type offerGroupDraftRequest struct {
	SelectedOfferIDs []uuid.UUID `json:"selectedOfferIds"`
	ResponderOfferID *uuid.UUID  `json:"responderOfferId,omitempty"`
	Name             *string     `json:"name,omitempty"`
	Description      *string     `json:"description,omitempty"`
}

type offerGroupResponse struct {
	ID uuid.UUID `json:"id"`
}

type seededUser struct {
	Key      string
	Name     string
	Bio      string
	Email    string
	Password string
	UserID   uuid.UUID
	Token    string
}

type offerSpec struct {
	Key          string
	Name         string
	Description  string
	Type         dealtypes.ItemType
	Action       dealtypes.OfferAction
	Tags         []dealtypes.TagName
	PhotoAliases []string
	SkipPhoto    bool
}

type SeedSummary struct {
	Users    []seededUserSummary `json:"users"`
	Warnings []string            `json:"warnings,omitempty"`

	OfferGroupID      uuid.UUID `json:"offerGroupId"`
	OfferGroupDraftID uuid.UUID `json:"offerGroupDraftId"`
	MultiUnitGroupID  uuid.UUID `json:"multiUnitGroupId"`

	LookingDealID    uuid.UUID `json:"lookingDealId"`
	DiscussionDealID uuid.UUID `json:"discussionDealId"`
	ConfirmedDealID  uuid.UUID `json:"confirmedDealId"`
	CompletedDealID  uuid.UUID `json:"completedDealId"`
	CompletedDeal2ID uuid.UUID `json:"completedDeal2Id"`
	CompletedDeal3ID uuid.UUID `json:"completedDeal3Id"`
	CancelledDealID  uuid.UUID `json:"cancelledDealId"`
	FailedDealID     uuid.UUID `json:"failedDealId"`
	JoinDealID       uuid.UUID `json:"joinDealId"`

	DirectChatID uuid.UUID `json:"directChatId"`
	DealChatID   uuid.UUID `json:"dealChatId"`

	PendingReportID  uuid.UUID `json:"pendingReportId"`
	AcceptedReportID uuid.UUID `json:"acceptedReportId"`
	RejectedReportID uuid.UUID `json:"rejectedReportId"`
}

type seededUserSummary struct {
	Key      string `json:"key"`
	UserID   string `json:"userId"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type chatMessage struct {
	Token   string
	Content string
}
