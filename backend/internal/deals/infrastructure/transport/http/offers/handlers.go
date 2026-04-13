package offers

import (
	"barter-port/contracts/openapi/deals/types"
	offersapp "barter-port/internal/deals/application/offers"
	"barter-port/internal/deals/domain"
	enums "barter-port/internal/deals/domain/enums"
	"barter-port/internal/deals/domain/htypes"
	"barter-port/pkg/authkit"
	"barter-port/pkg/httpx"
	"barter-port/pkg/logger"
	"errors"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

const (
	maxOfferPhotoUploadSize = 5 * 1024 * 1024
	maxOfferPhotoCount      = 10
)

type Handlers struct {
	offerService *offersapp.Service
}

func NewHandlers(offerService *offersapp.Service) *Handlers {
	return &Handlers{offerService: offerService}
}

// ================================================================================
// CREATE OFFER
// ================================================================================

func (h *Handlers) HandleCreateOffer(w http.ResponseWriter, r *http.Request) {
	log := logger.LogFrom(r.Context(), slog.Default())
	log.Info("handling create offer request")

	userID, ok := authkit.UserIDFromContext(r.Context())
	if !ok {
		log.Error("failed to get userID from context")
		httpx.WriteEmptyError(w, http.StatusUnauthorized)
		return
	}

	req, photos, err := decodeCreateOfferRequest(w, r)
	if err != nil {
		log.Error("error decoding request", slog.Any("error", err))
		httpx.WriteErrorStr(w, http.StatusBadRequest, err.Error())
		return
	}
	if r.MultipartForm != nil {
		defer func() { _ = r.MultipartForm.RemoveAll() }()
	}

	log = log.With(slog.Any("request", req))
	log.Debug("decoded create offer request")

	itemType, err := enums.ItemTypeString(string(req.Type))
	if err != nil {
		log.Error("invalid item type", slog.String("type", string(req.Type)), slog.Any("error", err))
		httpx.WriteErrorStr(w, http.StatusBadRequest, "invalid item type")
		return
	}

	action, err := enums.OfferActionString(string(req.Action))
	if err != nil {
		log.Error("invalid offer action", slog.String("action", string(req.Action)), slog.Any("error", err))
		httpx.WriteErrorStr(w, http.StatusBadRequest, "invalid offer action")
		return
	}

	offer, err := h.offerService.CreateOffer(r.Context(), userID, req.Name, itemType, action, req.Description, photos)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidOfferName) {
			log.Warn("invalid offer name", slog.Any("error", err))
			httpx.WriteError(w, http.StatusBadRequest, domain.ErrInvalidOfferName)
			return
		}
		if errors.Is(err, offersapp.ErrOfferPhotoStorageNotConfigured) {
			log.Error("offer photo storage is not configured", slog.Any("error", err))
			httpx.WriteEmptyError(w, http.StatusInternalServerError)
			return
		}
		log.Error("failed to create offer", slog.Any("error", err))
		httpx.WriteEmptyError(w, http.StatusInternalServerError)
		return
	}

	log.Debug(
		"offer created successfully",
		slog.String("offer_id", offer.ID.String()),
		slog.Int("photo_count", len(offer.PhotoUrls)),
	)

	httpx.WriteJSON(w, http.StatusCreated, offer.ToDto())
}

func decodeCreateOfferRequest(w http.ResponseWriter, r *http.Request) (types.CreateOfferRequest, []offersapp.PhotoUpload, error) {
	if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
		return decodeCreateOfferMultipartRequest(w, r)
	}

	var req types.CreateOfferRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		return types.CreateOfferRequest{}, nil, httpx.ErrCannotDecodeRequestBody
	}

	return req, nil, nil
}

func decodeCreateOfferMultipartRequest(w http.ResponseWriter, r *http.Request) (types.CreateOfferRequest, []offersapp.PhotoUpload, error) {
	maxBodySize := int64(maxOfferPhotoCount*maxOfferPhotoUploadSize + (1 << 20))
	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)
	if err := r.ParseMultipartForm(maxBodySize); err != nil {
		return types.CreateOfferRequest{}, nil, errors.New("invalid offer upload")
	}

	req := types.CreateOfferRequest{
		Name:        r.FormValue("name"),
		Description: r.FormValue("description"),
		Type:        types.ItemType(r.FormValue("type")),
		Action:      types.OfferAction(r.FormValue("action")),
	}

	fileHeaders := r.MultipartForm.File["photos"]
	if len(fileHeaders) > maxOfferPhotoCount {
		return types.CreateOfferRequest{}, nil, errors.New("too many offer photos")
	}

	photos := make([]offersapp.PhotoUpload, 0, len(fileHeaders))
	for _, fileHeader := range fileHeaders {
		file, err := fileHeader.Open()
		if err != nil {
			return types.CreateOfferRequest{}, nil, errors.New("failed to read offer photo")
		}

		content, contentType, readErr := readOfferPhotoUpload(file)
		closeErr := file.Close()
		if readErr != nil {
			return types.CreateOfferRequest{}, nil, readErr
		}
		if closeErr != nil {
			return types.CreateOfferRequest{}, nil, errors.New("failed to close offer photo")
		}

		photos = append(photos, offersapp.PhotoUpload{
			ContentType: contentType,
			Content:     content,
		})
	}

	return req, photos, nil
}

func readOfferPhotoUpload(file multipart.File) ([]byte, string, error) {
	content, err := io.ReadAll(io.LimitReader(file, maxOfferPhotoUploadSize+1))
	if err != nil {
		return nil, "", errors.New("failed to read offer photo")
	}
	if len(content) == 0 {
		return nil, "", errors.New("offer photo is empty")
	}
	if len(content) > maxOfferPhotoUploadSize {
		return nil, "", errors.New("offer photo exceeds 5 MB")
	}

	contentType := http.DetectContentType(content)
	if !strings.HasPrefix(contentType, "image/") {
		return nil, "", errors.New("offer photo must be an image")
	}

	return content, contentType, nil
}

func parseOfferID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	idStr := chi.URLParam(r, "offerId")
	id, err := uuid.Parse(idStr)
	if err != nil {
		httpx.WriteEmptyError(w, http.StatusBadRequest)
		return uuid.Nil, false
	}
	return id, true
}

// ================================================================================
// GET OFFER BY ID
// ================================================================================

func (h *Handlers) HandleGetOfferByID(w http.ResponseWriter, r *http.Request) {
	log := logger.LogFrom(r.Context(), slog.Default()).With(slog.String("handler", "GetOfferByID"))
	log.Info("handling get offer by id request")

	id, ok := parseOfferID(w, r)
	if !ok {
		log.Error("error parsing offer id")
		return
	}

	offer, err := h.offerService.GetOfferByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrOfferNotFound) {
			log.Info("offer not found", slog.String("offerId", id.String()))
			httpx.WriteEmptyError(w, http.StatusNotFound)
			return
		}
		log.Error("failed to get offer", slog.Any("error", err))
		httpx.WriteEmptyError(w, http.StatusInternalServerError)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, offer.ToDto())
}

// ================================================================================
// UPDATE OFFER
// ================================================================================

func (h *Handlers) HandleUpdateOffer(w http.ResponseWriter, r *http.Request) {
	log := logger.LogFrom(r.Context(), slog.Default()).With(slog.String("handler", "UpdateOffer"))
	log.Info("handling update offer request")

	offerID, ok := parseOfferID(w, r)
	if !ok {
		return
	}

	userID, ok := authkit.UserIDFromContext(r.Context())
	if !ok {
		log.Error("failed to get userID from context")
		httpx.WriteEmptyError(w, http.StatusUnauthorized)
		return
	}

	var req types.UpdateOfferRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCannotDecodeRequestBody)
		return
	}
	if req.Name == nil && req.Description == nil && req.Type == nil && req.Action == nil {
		httpx.WriteEmptyError(w, http.StatusBadRequest)
		return
	}

	var patch htypes.OfferPatch
	patch.Name = req.Name
	patch.Description = req.Description

	if req.Type != nil {
		itemType, err := enums.ItemTypeString(string(*req.Type))
		if err != nil {
			httpx.WriteErrorStr(w, http.StatusBadRequest, "invalid item type")
			return
		}
		patch.Type = &itemType
	}

	if req.Action != nil {
		action, err := enums.OfferActionString(string(*req.Action))
		if err != nil {
			httpx.WriteErrorStr(w, http.StatusBadRequest, "invalid offer action")
			return
		}
		patch.Action = &action
	}

	offer, err := h.offerService.UpdateOffer(r.Context(), userID, offerID, patch)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidOfferName):
			httpx.WriteError(w, http.StatusBadRequest, domain.ErrInvalidOfferName)
		case errors.Is(err, domain.ErrOfferNotFound):
			httpx.WriteEmptyError(w, http.StatusNotFound)
		case errors.Is(err, domain.ErrForbidden):
			httpx.WriteEmptyError(w, http.StatusForbidden)
		default:
			log.Error("failed to update offer", slog.Any("error", err))
			httpx.WriteEmptyError(w, http.StatusInternalServerError)
		}
		return
	}

	httpx.WriteJSON(w, http.StatusOK, offer.ToDto())
}

// ================================================================================
// DELETE OFFER
// ================================================================================

func (h *Handlers) HandleDeleteOffer(w http.ResponseWriter, r *http.Request) {
	log := logger.LogFrom(r.Context(), slog.Default()).With(slog.String("handler", "DeleteOffer"))
	log.Info("handling delete offer request")

	offerID, ok := parseOfferID(w, r)
	if !ok {
		return
	}

	userID, ok := authkit.UserIDFromContext(r.Context())
	if !ok {
		log.Error("failed to get userID from context")
		httpx.WriteEmptyError(w, http.StatusUnauthorized)
		return
	}

	err := h.offerService.DeleteOffer(r.Context(), userID, offerID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrOfferNotFound):
			httpx.WriteEmptyError(w, http.StatusNotFound)
		case errors.Is(err, domain.ErrForbidden):
			httpx.WriteEmptyError(w, http.StatusForbidden)
		default:
			log.Error("failed to delete offer", slog.Any("error", err))
			httpx.WriteEmptyError(w, http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ================================================================================
// GET OFFERS
// ================================================================================

func (h *Handlers) HandleGetOffers(w http.ResponseWriter, r *http.Request) {
	log := logger.LogFrom(r.Context(), slog.Default())

	// Parse query parameters
	sortTypeStr := r.URL.Query().Get("sort")
	myStr := r.URL.Query().Get("my")
	createdAtStr := r.URL.Query().Get("cursor_created_at")
	viewsStr := r.URL.Query().Get("cursor_views")
	idStr := r.URL.Query().Get("cursor_id")
	limitStr := r.URL.Query().Get("cursor_limit")

	my := false
	if myStr != "" {
		parsedMy, err := strconv.ParseBool(myStr)
		if err != nil {
			log.Error("invalid my filter", slog.String("my", myStr), slog.Any("error", err))
			httpx.WriteErrorStr(w, http.StatusBadRequest, "invalid my filter")
			return
		}
		my = parsedMy
	}

	log = log.With(
		slog.String("sortTypeStr", sortTypeStr),
		slog.String("myStr", myStr),
		slog.Bool("my", my),
		slog.String("createdAtStr", createdAtStr),
		slog.String("viewsStr", viewsStr),
		slog.String("idStr", idStr),
		slog.String("limitStr", limitStr),
	)
	log.Info("handling get offers request")

	sortType, err := enums.SortTypeString(sortTypeStr)
	if err != nil {
		log.Error("invalid sort type", slog.Any("error", err))
		httpx.WriteErrorStr(w, http.StatusBadRequest, "invalid sort type")
		return
	}

	cursor, err := domain.NewUniversalCursor(createdAtStr, viewsStr, idStr)
	if err != nil {
		log.Error("failed to create offers cursor", slog.Any("error", err))
		cursor = nil
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		log.Warn("invalid limit", slog.Any("error", err))
		limit = 10
	}

	log.Debug("parsing finished", slog.Any("cursor", cursor), slog.Int("limit", limit), slog.Bool("my", my))

	var authorID *uuid.UUID
	if my {
		userID, ok := authkit.UserIDFromContext(r.Context())
		if !ok {
			log.Error("failed to get userID from context")
			httpx.WriteEmptyError(w, http.StatusInternalServerError)
			return
		}
		authorID = &userID
	}

	// Fetch offers from the service
	offerList, nextCursor, err := h.offerService.GetOffers(r.Context(), sortType, cursor, limit, authorID)
	if err != nil {
		log.Error("failed to get items", slog.Any("error", err))
		httpx.WriteEmptyError(w, http.StatusInternalServerError)
		return
	}

	respOffers := make([]types.Offer, len(offerList))
	for i, offer := range offerList {
		respOffers[i] = offer.ToDto()
	}

	var respCursor *types.OffersCursor
	if nextCursor != nil {
		respCursor = new(nextCursor.ToDto())
	}

	httpx.WriteJSON(w, http.StatusOK, types.ListOffersResponse{
		Offers:     respOffers,
		NextCursor: respCursor,
	})
}
