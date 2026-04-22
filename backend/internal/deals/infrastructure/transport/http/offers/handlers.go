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
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"time"

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

	tags, err := normalizeTagNames(req.Tags)
	if err != nil {
		httpx.WriteErrorStr(w, http.StatusBadRequest, err.Error())
		return
	}

	offer, err := h.offerService.CreateOffer(r.Context(), userID, req.Name, itemType, action, req.Description, tags, photos)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidOfferName) {
			log.Warn("invalid offer name", slog.Any("error", err))
			httpx.WriteError(w, http.StatusBadRequest, domain.ErrInvalidOfferName)
			return
		}
		if errors.Is(err, domain.ErrInvalidTagName) || errors.Is(err, domain.ErrTooManyTags) {
			httpx.WriteErrorStr(w, http.StatusBadRequest, err.Error())
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

func decodeCreateOfferRequest(w http.ResponseWriter, r *http.Request) (types.CreateOffersJSONRequestBody, []offersapp.PhotoUpload, error) {
	if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
		return decodeCreateOfferMultipartRequest(w, r)
	}

	var req types.CreateOffersJSONRequestBody
	if err := httpx.DecodeJSON(r, &req); err != nil {
		return types.CreateOffersJSONRequestBody{}, nil, httpx.ErrCannotDecodeRequestBody
	}

	return req, nil, nil
}

func decodeCreateOfferMultipartRequest(w http.ResponseWriter, r *http.Request) (types.CreateOffersMultipartRequestBody, []offersapp.PhotoUpload, error) {
	maxBodySize := int64(maxOfferPhotoCount*maxOfferPhotoUploadSize + (1 << 20))
	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)
	if err := r.ParseMultipartForm(maxBodySize); err != nil {
		return types.CreateOffersMultipartRequestBody{}, nil, errors.New("invalid offer upload")
	}

	req := types.CreateOffersMultipartRequestBody{
		Name:        r.FormValue("name"),
		Description: r.FormValue("description"),
		Type:        types.ItemType(r.FormValue("type")),
		Action:      types.OfferAction(r.FormValue("action")),
	}
	if rawTags, ok := r.MultipartForm.Value["tags"]; ok {
		tags := make([]types.TagName, 0, len(rawTags))
		for _, tag := range rawTags {
			tags = append(tags, types.TagName(tag))
		}
		req.Tags = &tags
	}

	fileHeaders := r.MultipartForm.File["photos"]
	if len(fileHeaders) > maxOfferPhotoCount {
		return types.CreateOffersMultipartRequestBody{}, nil, errors.New("too many offer photos")
	}

	photos := make([]offersapp.PhotoUpload, 0, len(fileHeaders))
	for _, fileHeader := range fileHeaders {
		file, err := fileHeader.Open()
		if err != nil {
			return types.CreateOffersMultipartRequestBody{}, nil, errors.New("failed to read offer photo")
		}

		content, contentType, readErr := readOfferPhotoUpload(file)
		closeErr := file.Close()
		if readErr != nil {
			return types.CreateOffersMultipartRequestBody{}, nil, readErr
		}
		if closeErr != nil {
			return types.CreateOffersMultipartRequestBody{}, nil, errors.New("failed to close offer photo")
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

	requesterID, ok := authkit.UserIDFromContext(r.Context())
	var requesterIDPtr *uuid.UUID
	if ok {
		requesterIDPtr = &requesterID
	}

	offer, err := h.offerService.GetOfferByID(r.Context(), id, requesterIDPtr)
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
// VIEW OFFER BY ID
// ================================================================================

func (h *Handlers) HandleViewOfferByID(w http.ResponseWriter, r *http.Request) {
	log := logger.LogFrom(r.Context(), slog.Default()).With(slog.String("handler", "ViewOfferByID"))
	log.Info("handling view offer by id request")

	id, ok := parseOfferID(w, r)
	if !ok {
		log.Error("error parsing offer id")
		return
	}

	err := h.offerService.ViewOfferByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrOfferNotFound) {
			log.Info("offer not found", slog.String("offerId", id.String()))
			httpx.WriteEmptyError(w, http.StatusNotFound)
			return
		}
		log.Error("failed to register offer view", slog.Any("error", err))
		httpx.WriteEmptyError(w, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
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

	req, photos, err := decodeUpdateOfferRequest(w, r)
	if err != nil {
		httpx.WriteErrorStr(w, http.StatusBadRequest, err.Error())
		return
	}
	if r.MultipartForm != nil {
		defer func() { _ = r.MultipartForm.RemoveAll() }()
	}
	if req.Name == nil && req.Description == nil && req.Type == nil && req.Action == nil && req.Tags == nil &&
		(req.DeletePhotoIds == nil || len(*req.DeletePhotoIds) == 0) && len(photos) == 0 {
		httpx.WriteEmptyError(w, http.StatusBadRequest)
		return
	}

	var patch htypes.OfferPatch
	patch.Name = req.Name
	patch.Description = req.Description
	if req.DeletePhotoIds != nil {
		patch.DeletePhotoIds = make([]uuid.UUID, 0, len(*req.DeletePhotoIds))
		for _, photoID := range *req.DeletePhotoIds {
			patch.DeletePhotoIds = append(patch.DeletePhotoIds, photoID)
		}
	}
	if req.Tags != nil {
		tags, normalizeErr := normalizeTagNames(req.Tags)
		if normalizeErr != nil {
			httpx.WriteErrorStr(w, http.StatusBadRequest, normalizeErr.Error())
			return
		}
		patch.Tags = &tags
	}

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

	offer, err := h.offerService.UpdateOffer(r.Context(), userID, offerID, patch, photos)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidOfferName):
			httpx.WriteError(w, http.StatusBadRequest, domain.ErrInvalidOfferName)
		case errors.Is(err, offersapp.ErrOfferPhotoLimitExceeded), errors.Is(err, offersapp.ErrOfferPhotoNotFound):
			httpx.WriteErrorStr(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, domain.ErrOfferNotFound):
			httpx.WriteEmptyError(w, http.StatusNotFound)
		case errors.Is(err, domain.ErrForbidden):
			httpx.WriteEmptyError(w, http.StatusForbidden)
		case errors.Is(err, domain.ErrModificationBlocked):
			httpx.WriteEmptyError(w, http.StatusConflict)
		case errors.Is(err, offersapp.ErrOfferPhotoStorageNotConfigured):
			log.Error("offer photo storage is not configured", slog.Any("error", err))
			httpx.WriteEmptyError(w, http.StatusInternalServerError)
		default:
			log.Error("failed to update offer", slog.Any("error", err))
			httpx.WriteEmptyError(w, http.StatusInternalServerError)
		}
		return
	}

	httpx.WriteJSON(w, http.StatusOK, offer.ToDto())
}

func decodeUpdateOfferRequest(w http.ResponseWriter, r *http.Request) (types.UpdateOfferByIdJSONRequestBody, []offersapp.PhotoUpload, error) {
	if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
		return decodeUpdateOfferMultipartRequest(w, r)
	}

	var req types.UpdateOfferByIdJSONRequestBody
	if err := httpx.DecodeJSON(r, &req); err != nil {
		return types.UpdateOfferByIdJSONRequestBody{}, nil, httpx.ErrCannotDecodeRequestBody
	}

	return req, nil, nil
}

func decodeUpdateOfferMultipartRequest(w http.ResponseWriter, r *http.Request) (types.UpdateOfferByIdMultipartRequestBody, []offersapp.PhotoUpload, error) {
	maxBodySize := int64(maxOfferPhotoCount*maxOfferPhotoUploadSize + (1 << 20))
	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)
	if err := r.ParseMultipartForm(maxBodySize); err != nil {
		return types.UpdateOfferByIdMultipartRequestBody{}, nil, errors.New("invalid offer upload")
	}

	values := r.MultipartForm.Value
	req := types.UpdateOfferByIdMultipartRequestBody{}

	if value, ok := firstMultipartValue(values, "name"); ok {
		req.Name = &value
	}
	if value, ok := firstMultipartValue(values, "description"); ok {
		req.Description = &value
	}
	if value, ok := firstMultipartValue(values, "type"); ok {
		itemType := types.ItemType(value)
		req.Type = &itemType
	}
	if value, ok := firstMultipartValue(values, "action"); ok {
		action := types.OfferAction(value)
		req.Action = &action
	}
	if rawTags, ok := values["tags"]; ok {
		tags := make([]types.TagName, 0, len(rawTags))
		for _, tag := range rawTags {
			tags = append(tags, types.TagName(tag))
		}
		req.Tags = &tags
	}
	if rawIDs, ok := values["deletePhotoIds"]; ok {
		deletePhotoIDs := make([]uuid.UUID, 0, len(rawIDs))
		for _, rawID := range rawIDs {
			photoID, err := uuid.Parse(rawID)
			if err != nil {
				return types.UpdateOfferByIdMultipartRequestBody{}, nil, errors.New("invalid deletePhotoIds")
			}
			deletePhotoIDs = append(deletePhotoIDs, photoID)
		}
		req.DeletePhotoIds = &deletePhotoIDs
	}

	fileHeaders := r.MultipartForm.File["photos"]
	if len(fileHeaders) > maxOfferPhotoCount {
		return types.UpdateOfferByIdMultipartRequestBody{}, nil, errors.New("too many offer photos")
	}

	photos := make([]offersapp.PhotoUpload, 0, len(fileHeaders))
	for _, fileHeader := range fileHeaders {
		file, err := fileHeader.Open()
		if err != nil {
			return types.UpdateOfferByIdMultipartRequestBody{}, nil, errors.New("failed to read offer photo")
		}

		content, contentType, readErr := readOfferPhotoUpload(file)
		closeErr := file.Close()
		if readErr != nil {
			return types.UpdateOfferByIdMultipartRequestBody{}, nil, readErr
		}
		if closeErr != nil {
			return types.UpdateOfferByIdMultipartRequestBody{}, nil, errors.New("failed to close offer photo")
		}

		photos = append(photos, offersapp.PhotoUpload{
			ContentType: contentType,
			Content:     content,
		})
	}

	return req, photos, nil
}

func firstMultipartValue(values map[string][]string, key string) (string, bool) {
	raw, ok := values[key]
	if !ok || len(raw) == 0 {
		return "", false
	}

	return raw[0], true
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
		case errors.Is(err, domain.ErrModificationBlocked):
			httpx.WriteEmptyError(w, http.StatusConflict)
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

	sortType, cursor, limit, my, tagFilter, ok := parseGetOffersRequest(w, r, log)
	if !ok {
		return
	}

	requesterID, requesterOK := authkit.UserIDFromContext(r.Context())
	var requesterIDPtr *uuid.UUID
	if requesterOK {
		requesterIDPtr = &requesterID
	}

	var authorID *uuid.UUID
	if my {
		if !requesterOK {
			log.Error("failed to get userID from context")
			httpx.WriteEmptyError(w, http.StatusUnauthorized)
			return
		}
		authorID = &requesterID
	}

	offerList, nextCursor, err := h.offerService.GetOffers(r.Context(), sortType, cursor, limit, authorID, requesterIDPtr, tagFilter)
	if err != nil {
		log.Error("failed to get items", slog.Any("error", err))
		httpx.WriteEmptyError(w, http.StatusInternalServerError)
		return
	}

	writeListOffersResponse(w, offerList, nextCursor)
}

func (h *Handlers) HandleGetSubscribedOffers(w http.ResponseWriter, r *http.Request) {
	log := logger.LogFrom(r.Context(), slog.Default())

	sortType, cursor, limit, _, _, ok := parseGetOffersRequest(w, r, log)
	if !ok {
		return
	}

	requesterID, ok := authkit.UserIDFromContext(r.Context())
	if !ok {
		log.Error("failed to get userID from context")
		httpx.WriteEmptyError(w, http.StatusUnauthorized)
		return
	}

	offerList, nextCursor, err := h.offerService.GetSubscribedOffers(r.Context(), requesterID, sortType, cursor, limit)
	if err != nil {
		log.Error("failed to get subscribed offers", slog.Any("error", err))
		httpx.WriteEmptyError(w, http.StatusInternalServerError)
		return
	}

	writeListOffersResponse(w, offerList, nextCursor)
}

func parseGetOffersRequest(
	w http.ResponseWriter,
	r *http.Request,
	log *slog.Logger,
) (enums.SortType, *domain.UniversalCursor, int, bool, *[]string, bool) {
	params, tagsFilter, err := decodeListOffersRequest(r)
	if err != nil {
		log.Error("failed to decode list offers request", slog.Any("error", err))
		httpx.WriteErrorStr(w, http.StatusBadRequest, err.Error())
		return enums.SortType(0), nil, 0, false, nil, false
	}

	my := params.My != nil && *params.My
	log.Info("handling get offers request")

	sortType, err := enums.SortTypeString(string(params.Sort))
	if err != nil {
		log.Error("invalid sort type", slog.Any("error", err))
		httpx.WriteErrorStr(w, http.StatusBadRequest, "invalid sort type")
		return enums.SortType(0), nil, 0, false, nil, false
	}

	cursor, err := newUniversalCursorFromParams(params.CursorCreatedAt, params.CursorViews, params.CursorId)
	if err != nil {
		log.Error("failed to create offers cursor", slog.Any("error", err))
		httpx.WriteErrorStr(w, http.StatusBadRequest, "invalid cursor")
		return enums.SortType(0), nil, 0, false, nil, false
	}

	limit := 10
	if params.CursorLimit != nil {
		limit = int(*params.CursorLimit)
		if limit <= 0 {
			httpx.WriteErrorStr(w, http.StatusBadRequest, "invalid limit")
			return enums.SortType(0), nil, 0, false, nil, false
		}
	}

	log.Debug("parsing finished", slog.Any("cursor", cursor), slog.Int("limit", limit), slog.Bool("my", my))

	return sortType, cursor, limit, my, tagsFilter, true
}

func writeListOffersResponse(w http.ResponseWriter, offerList []domain.Offer, nextCursor *domain.UniversalCursor) {
	respOffers := make([]types.Offer, len(offerList))
	for i, offer := range offerList {
		respOffers[i] = offer.ToDto()
	}

	var respCursor *types.OffersCursor
	if nextCursor != nil {
		cursorDTO := nextCursor.ToDto()
		respCursor = &cursorDTO
	}

	httpx.WriteJSON(w, http.StatusOK, types.ListOffersResponse{
		Offers:     respOffers,
		NextCursor: respCursor,
	})
}

type listOffersRequestBody struct {
	Tags *[]types.TagName `json:"tags,omitempty"`
}

func decodeListOffersRequest(r *http.Request) (types.ListOffersParams, *[]string, error) {
	params := types.ListOffersParams{}
	query := r.URL.Query()

	params.Sort = types.ListOffersParamsSort(query.Get("sort"))
	if rawMy := query.Get("my"); rawMy != "" {
		value, err := strconv.ParseBool(rawMy)
		if err != nil {
			return types.ListOffersParams{}, nil, errors.New("invalid my filter")
		}
		params.My = &value
	}
	if rawCreatedAt := query.Get("cursor_created_at"); rawCreatedAt != "" {
		value, err := time.Parse(time.RFC3339, rawCreatedAt)
		if err != nil {
			return types.ListOffersParams{}, nil, errors.New("invalid cursor_created_at")
		}
		createdAt := types.CursorCreatedAt(value)
		params.CursorCreatedAt = &createdAt
	}
	if rawViews := query.Get("cursor_views"); rawViews != "" {
		value, err := strconv.ParseInt(rawViews, 10, 64)
		if err != nil {
			return types.ListOffersParams{}, nil, errors.New("invalid cursor_views")
		}
		cursorViews := types.CursorViews(value)
		params.CursorViews = &cursorViews
	}
	if rawID := query.Get("cursor_id"); rawID != "" {
		value, err := uuid.Parse(rawID)
		if err != nil {
			return types.ListOffersParams{}, nil, errors.New("invalid cursor_id")
		}
		cursorID := types.CursorId(value)
		params.CursorId = &cursorID
	}
	if rawLimit := query.Get("cursor_limit"); rawLimit != "" {
		value, err := strconv.Atoi(rawLimit)
		if err != nil {
			return types.ListOffersParams{}, nil, errors.New("invalid cursor_limit")
		}
		limit := types.Limit(value)
		params.CursorLimit = &limit
	}

	tagFilter, err := decodeListOffersTagFilter(r)
	if err != nil {
		return types.ListOffersParams{}, nil, err
	}

	return params, tagFilter, nil
}

func decodeListOffersTagFilter(r *http.Request) (*[]string, error) {
	if r.Body == nil {
		return nil, nil
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, httpx.ErrCannotDecodeRequestBody
	}
	if len(strings.TrimSpace(string(body))) == 0 {
		return nil, nil
	}

	var req listOffersRequestBody
	if err := json.Unmarshal(body, &req); err != nil {
		return nil, httpx.ErrCannotDecodeRequestBody
	}
	if req.Tags == nil {
		return nil, nil
	}

	normalized, err := normalizeTagNames(req.Tags)
	if err != nil {
		return nil, err
	}

	return &normalized, nil
}

func normalizeTagNames(tags *[]types.TagName) ([]string, error) {
	if tags == nil {
		return nil, nil
	}

	raw := make([]string, 0, len(*tags))
	for _, tag := range *tags {
		raw = append(raw, string(tag))
	}

	return domain.NormalizeTags(raw)
}

func newUniversalCursorFromParams(createdAt *types.CursorCreatedAt, views *types.CursorViews, id *types.CursorId) (*domain.UniversalCursor, error) {
	if createdAt == nil && views == nil && id == nil {
		return nil, nil
	}

	var createdAtStr string
	if createdAt != nil {
		createdAtStr = time.Time(*createdAt).Format(time.RFC3339)
	}

	var viewsStr string
	if views != nil {
		viewsStr = strconv.FormatInt(int64(*views), 10)
	}

	var idStr string
	if id != nil {
		idStr = uuid.UUID(*id).String()
	}

	return domain.NewUniversalCursor(createdAtStr, viewsStr, idStr)
}
