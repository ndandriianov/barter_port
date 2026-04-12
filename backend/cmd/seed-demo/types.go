package main

import (
	dealtypes "barter-port/contracts/openapi/deals/types"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type seedClient struct {
	baseURL      string
	httpClient   *http.Client
	pollInterval time.Duration
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
	Avatar   string
	UserID   uuid.UUID
	Token    string
}

type offerSpec struct {
	Key         string
	Name        string
	Description string
	Type        dealtypes.ItemType
	Action      dealtypes.OfferAction
}

type seedSummary struct {
	Users             []seededUserSummary `json:"users"`
	OfferGroupID      uuid.UUID           `json:"offerGroupId"`
	OfferGroupDraftID uuid.UUID           `json:"offerGroupDraftId"`
	DiscussionDealID  uuid.UUID           `json:"discussionDealId"`
	CompletedDealID   uuid.UUID           `json:"completedDealId"`
	DirectChatID      uuid.UUID           `json:"directChatId"`
	DealChatID        uuid.UUID           `json:"dealChatId"`
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
