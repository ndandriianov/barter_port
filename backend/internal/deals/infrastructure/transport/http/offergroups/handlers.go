package offergroups

import (
	"barter-port/contracts/openapi/deals/types"
	offergroupssvc "barter-port/internal/deals/application/offergroups"
	"barter-port/internal/deals/domain"
	"barter-port/pkg/authkit"
	"barter-port/pkg/httpx"
	"barter-port/pkg/logger"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Handlers struct {
	log                *slog.Logger
	offerGroupsService *offergroupssvc.Service
}

func NewHandlers(log *slog.Logger, offerGroupsService *offergroupssvc.Service) *Handlers {
	return &Handlers{
		log:                log,
		offerGroupsService: offerGroupsService,
	}
}

type createOfferGroupRequest struct {
	Name        *string                       `json:"name,omitempty"`
	Description *string                       `json:"description,omitempty"`
	Units       []createOfferGroupUnitRequest `json:"units"`
}

type createOfferGroupUnitRequest struct {
	Offers []offerRef `json:"offers"`
}

type offerRef struct {
	OfferID uuid.UUID `json:"offerId"`
}

type createOfferGroupDraftRequest struct {
	SelectedOfferIDs []uuid.UUID `json:"selectedOfferIds"`
	ResponderOfferID *uuid.UUID  `json:"responderOfferId,omitempty"`
	Name             *string     `json:"name,omitempty"`
	Description      *string     `json:"description,omitempty"`
}

func (h *Handlers) ListOfferGroups(w http.ResponseWriter, r *http.Request) {
	log := logger.LogFrom(r.Context(), h.log).With(slog.String("handler", "ListOfferGroups"))
	log.Info("handling list offer groups request")

	userID, ok := authkit.UserIDFromContext(r.Context())
	if !ok {
		log.Error("failed to get userID from context")
		httpx.WriteEmptyError(w, http.StatusUnauthorized)
		return
	}

	items, err := h.offerGroupsService.ListOfferGroups(r.Context(), userID)
	if err != nil {
		log.Error("error listing offer groups", slog.Any("error", err))
		httpx.WriteEmptyError(w, http.StatusInternalServerError)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, mapOfferGroupsToDTO(items))
}

func (h *Handlers) CreateOfferGroup(w http.ResponseWriter, r *http.Request) {
	log := logger.LogFrom(r.Context(), h.log).With(slog.String("handler", "CreateOfferGroup"))
	log.Info("handling create offer group request")

	var req createOfferGroupRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		log.Error("error decoding request", slog.Any("error", err))
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCannotDecodeRequestBody)
		return
	}

	userID, ok := authkit.UserIDFromContext(r.Context())
	if !ok {
		log.Error("failed to get userID from context")
		httpx.WriteEmptyError(w, http.StatusUnauthorized)
		return
	}

	units := make([]domain.OfferGroupUnitCreateInput, 0, len(req.Units))
	for _, unit := range req.Units {
		offerIDs := make([]uuid.UUID, 0, len(unit.Offers))
		for _, offer := range unit.Offers {
			offerIDs = append(offerIDs, offer.OfferID)
		}
		units = append(units, domain.OfferGroupUnitCreateInput{OfferIDs: offerIDs})
	}

	item, err := h.offerGroupsService.CreateOfferGroup(r.Context(), userID, req.Name, req.Description, units)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidOfferName),
			errors.Is(err, domain.ErrNoOfferGroupUnits),
			errors.Is(err, domain.ErrEmptyOfferGroupUnit),
			errors.Is(err, domain.ErrDuplicateOfferInGroup),
			errors.Is(err, domain.ErrMixedOfferActionsInUnit):
			httpx.WriteError(w, http.StatusBadRequest, err)
		case errors.Is(err, domain.ErrOfferNotFound):
			httpx.WriteError(w, http.StatusNotFound, err)
		case errors.Is(err, domain.ErrForbidden):
			httpx.WriteError(w, http.StatusForbidden, err)
		default:
			log.Error("error creating offer group", slog.Any("error", err))
			httpx.WriteEmptyError(w, http.StatusInternalServerError)
		}
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, mapOfferGroupToDTO(item))
}

func (h *Handlers) GetOfferGroupByID(w http.ResponseWriter, r *http.Request) {
	log := logger.LogFrom(r.Context(), h.log).With(slog.String("handler", "GetOfferGroupByID"))
	log.Info("handling get offer group by id request")

	userID, ok := authkit.UserIDFromContext(r.Context())
	if !ok {
		log.Error("failed to get userID from context")
		httpx.WriteEmptyError(w, http.StatusUnauthorized)
		return
	}

	id, ok := parseOfferGroupID(w, r)
	if !ok {
		return
	}

	item, err := h.offerGroupsService.GetOfferGroupByID(r.Context(), id, userID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrOfferGroupNotFound):
			httpx.WriteError(w, http.StatusNotFound, err)
		default:
			log.Error("error getting offer group", slog.Any("error", err))
			httpx.WriteEmptyError(w, http.StatusInternalServerError)
		}
		return
	}

	httpx.WriteJSON(w, http.StatusOK, mapOfferGroupToDTO(item))
}

func (h *Handlers) CreateDraftFromOfferGroup(w http.ResponseWriter, r *http.Request) {
	log := logger.LogFrom(r.Context(), h.log).With(slog.String("handler", "CreateDraftFromOfferGroup"))
	log.Info("handling create draft from offer group request")

	offerGroupID, ok := parseOfferGroupID(w, r)
	if !ok {
		return
	}

	userID, ok := authkit.UserIDFromContext(r.Context())
	if !ok {
		log.Error("failed to get userID from context")
		httpx.WriteEmptyError(w, http.StatusUnauthorized)
		return
	}

	var req createOfferGroupDraftRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		log.Error("error decoding request", slog.Any("error", err))
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCannotDecodeRequestBody)
		return
	}

	id, err := h.offerGroupsService.CreateDraftFromOfferGroup(
		r.Context(),
		offerGroupID,
		userID,
		req.Name,
		req.Description,
		req.SelectedOfferIDs,
		req.ResponderOfferID,
	)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidOfferGroupSelect),
			errors.Is(err, domain.ErrOfferGroupResponderOfferRequired),
			errors.Is(err, domain.ErrOfferGroupResponderOfferAction):
			httpx.WriteError(w, http.StatusBadRequest, err)
		case errors.Is(err, domain.ErrForbidden):
			httpx.WriteError(w, http.StatusForbidden, err)
		case errors.Is(err, domain.ErrOfferGroupNotFound), errors.Is(err, domain.ErrOfferNotFound):
			httpx.WriteError(w, http.StatusNotFound, err)
		default:
			log.Error("error creating draft from offer group", slog.Any("error", err))
			httpx.WriteEmptyError(w, http.StatusInternalServerError)
		}
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, types.CreateDraftDealResponse{Id: id})
}

func parseOfferGroupID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	idStr := chi.URLParam(r, "offerGroupId")
	id, err := uuid.Parse(idStr)
	if err != nil {
		httpx.WriteEmptyError(w, http.StatusBadRequest)
		return uuid.Nil, false
	}
	return id, true
}
