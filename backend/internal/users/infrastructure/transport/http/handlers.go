package http

import (
	"barter-port/contracts/openapi/users/types"
	"barter-port/internal/users/application/user"
	"barter-port/internal/users/domain"
	"barter-port/pkg/authkit"
	"barter-port/pkg/httpx"
	httplog "barter-port/pkg/logger"
	"context"
	"errors"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

const maxAvatarUploadSize = 5 * 1024 * 1024

type Handlers struct {
	userService *user.Service
}

func NewHandlers(userService *user.Service) *Handlers {
	return &Handlers{userService: userService}
}

// ================================================================================
// GetUser
// ================================================================================

func (h *Handlers) HandleGetUser(w http.ResponseWriter, r *http.Request) {
	log := httplog.LogFrom(r.Context(), slog.Default())

	userID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		log.Warn("invalid user id", slog.Any("error", err))
		httpx.WriteErrorStr(w, http.StatusBadRequest, "invalid user id")
		return
	}

	u, err := h.userService.GetUser(r.Context(), userID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			log.Info("user not found", slog.String("user_id", userID.String()))
			httpx.WriteEmptyError(w, http.StatusNotFound)
			return
		}

		log.Error("failed to get user", slog.String("user_id", userID.String()), slog.String("error", err.Error()))
		httpx.WriteEmptyError(w, http.StatusInternalServerError)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, types.User{
		Id:          u.Id,
		Name:        u.Name,
		Bio:         u.Bio,
		AvatarUrl:   u.AvatarURL,
		PhoneNumber: u.PhoneNumber,
	})
}

// ================================================================================
// GetMe
// ================================================================================

func (h *Handlers) HandleGetMe(w http.ResponseWriter, r *http.Request) {
	log := httplog.LogFrom(r.Context(), slog.Default())

	userID, ok := authkit.UserIDFromContext(r.Context())
	if !ok {
		log.Warn("failed to get user id from context")
		httpx.WriteEmptyError(w, http.StatusUnauthorized)
		return
	}

	me, err := h.getMe(r.Context(), userID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			log.Error("current user is absent in users storage", slog.String("user_id", userID.String()))
			httpx.WriteEmptyError(w, http.StatusNotFound)
			return
		}

		log.Error("failed to get current user", slog.String("user_id", userID.String()), slog.String("error", err.Error()))
		httpx.WriteEmptyError(w, http.StatusInternalServerError)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, me)
}

// ================================================================================
// GetCurrentUserReputationEvents
// ================================================================================

func (h *Handlers) HandleGetCurrentUserReputationEvents(w http.ResponseWriter, r *http.Request) {
	log := httplog.LogFrom(r.Context(), slog.Default())

	userID, ok := authkit.UserIDFromContext(r.Context())
	if !ok {
		log.Warn("failed to get user id from context")
		httpx.WriteEmptyError(w, http.StatusUnauthorized)
		return
	}

	events, err := h.userService.GetCurrentUserReputationEvents(r.Context(), userID)
	if err != nil {
		log.Error("failed to get current user reputation events", slog.String("user_id", userID.String()), slog.String("error", err.Error()))
		httpx.WriteEmptyError(w, http.StatusInternalServerError)
		return
	}

	response := make(types.GetReputationEventsResponse, 0, len(events))
	for _, event := range events {
		response = append(response, types.ReputationEvent{
			Id:         event.Id,
			SourceType: types.ReputationEventSourceType(event.SourceType),
			SourceId:   event.SourceID,
			Delta:      event.Delta,
			CreatedAt:  event.CreatedAt,
			Comment:    event.Comment,
		})
	}

	httpx.WriteJSON(w, http.StatusOK, response)
}

// ================================================================================
// UploadMeAvatar
// ================================================================================

func (h *Handlers) HandleUploadMeAvatar(w http.ResponseWriter, r *http.Request) {
	log := httplog.LogFrom(r.Context(), slog.Default())

	userID, ok := authkit.UserIDFromContext(r.Context())
	if !ok {
		log.Warn("failed to get user id from context")
		httpx.WriteEmptyError(w, http.StatusUnauthorized)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxAvatarUploadSize+(1<<20))
	if err := r.ParseMultipartForm(maxAvatarUploadSize + (1 << 20)); err != nil {
		log.Warn("failed to parse avatar upload", slog.Any("error", err))
		httpx.WriteErrorStr(w, http.StatusBadRequest, "invalid avatar upload")
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		log.Warn("failed to read avatar file", slog.Any("error", err))
		httpx.WriteErrorStr(w, http.StatusBadRequest, "avatar file is required")
		return
	}
	defer func(file multipart.File) {
		err := file.Close()
		if err != nil {
			log.Warn("failed to close avatar file", slog.Any("error", err))
		}
	}(file)

	content, contentType, err := readAvatarUpload(file)
	if err != nil {
		log.Warn("failed to validate avatar file", slog.Any("error", err))
		httpx.WriteErrorStr(w, http.StatusBadRequest, err.Error())
		return
	}

	avatarURL, err := h.userService.UploadAvatar(r.Context(), userID, contentType, content)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			log.Info("user not found while uploading avatar", slog.String("user_id", userID.String()))
			httpx.WriteEmptyError(w, http.StatusNotFound)
			return
		}

		log.Error("failed to upload avatar", slog.String("user_id", userID.String()), slog.String("error", err.Error()))
		httpx.WriteEmptyError(w, http.StatusInternalServerError)
		return
	}

	log.Debug(
		"avatar uploaded successfully",
		slog.String("user_id", userID.String()),
		slog.String("content_type", contentType),
		slog.Int("size_bytes", len(content)),
		slog.String("avatar_url", avatarURL),
	)

	httpx.WriteJSON(w, http.StatusOK, types.AvatarUploadResponse{AvatarUrl: avatarURL})
}

// ================================================================================
// UpdateMe
// ================================================================================

func (h *Handlers) HandleUpdateMe(w http.ResponseWriter, r *http.Request) {
	log := httplog.LogFrom(r.Context(), slog.Default())

	userID, ok := authkit.UserIDFromContext(r.Context())
	if !ok {
		log.Warn("failed to get user id from context")
		httpx.WriteEmptyError(w, http.StatusUnauthorized)
		return
	}

	var req types.UpdateUserRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		log.Warn("failed to decode", slog.Any("error", err))
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCannotDecodeRequestBody)
		return
	}

	if req.Name == nil && req.Bio == nil && req.AvatarUrl == nil && req.PhoneNumber == nil &&
		req.CurrentLatitude == nil && req.CurrentLongitude == nil {
		httpx.WriteErrorStr(w, http.StatusBadRequest, "empty update payload")
		return
	}

	if req.Name != nil {
		if err := h.userService.UpdateName(r.Context(), userID, *req.Name); err != nil {
			handleUpdateError(w, log, err, userID)
			return
		}
	}

	if req.Bio != nil {
		if err := h.userService.UpdateBio(r.Context(), userID, req.Bio); err != nil {
			handleUpdateError(w, log, err, userID)
			return
		}
	}

	if req.AvatarUrl != nil {
		if err := h.userService.UpdateAvatarURL(r.Context(), userID, req.AvatarUrl); err != nil {
			handleUpdateError(w, log, err, userID)
			return
		}
	}

	if req.PhoneNumber != nil {
		if err := h.userService.UpdatePhoneNumber(r.Context(), userID, req.PhoneNumber); err != nil {
			handleUpdateError(w, log, err, userID)
			return
		}
	}

	if req.CurrentLatitude != nil || req.CurrentLongitude != nil {
		lat := (*float64)(req.CurrentLatitude)
		lon := (*float64)(req.CurrentLongitude)
		if err := h.userService.UpdateLocation(r.Context(), userID, lat, lon); err != nil {
			handleUpdateError(w, log, err, userID)
			return
		}
	}

	me, err := h.getMe(r.Context(), userID)
	if err != nil {
		log.Error("failed to load updated user", slog.String("user_id", userID.String()), slog.String("error", err.Error()))
		httpx.WriteEmptyError(w, http.StatusInternalServerError)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, me)
}

// ================================================================================
// Subscriptions
// ================================================================================

func (h *Handlers) HandleSubscribe(w http.ResponseWriter, r *http.Request) {
	log := httplog.LogFrom(r.Context(), slog.Default())

	userID, ok := authkit.UserIDFromContext(r.Context())
	if !ok {
		log.Warn("failed to get user id from context")
		httpx.WriteEmptyError(w, http.StatusUnauthorized)
		return
	}

	var req types.SubscribeToUserJSONRequestBody
	if err := httpx.DecodeJSON(r, &req); err != nil {
		log.Warn("failed to decode subscribe request", slog.Any("error", err))
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCannotDecodeRequestBody)
		return
	}

	err := h.userService.Subscribe(r.Context(), userID, req.TargetUserId)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrUserNotFound):
			httpx.WriteEmptyError(w, http.StatusNotFound)
		case errors.Is(err, domain.ErrAlreadySubscribed):
			httpx.WriteEmptyError(w, http.StatusConflict)
		case errors.Is(err, domain.ErrCannotSubscribeToYourself):
			httpx.WriteErrorStr(w, http.StatusBadRequest, err.Error())
		default:
			log.Error("failed to subscribe", slog.String("user_id", userID.String()), slog.Any("error", err))
			httpx.WriteEmptyError(w, http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *Handlers) HandleUnsubscribe(w http.ResponseWriter, r *http.Request) {
	log := httplog.LogFrom(r.Context(), slog.Default())

	userID, ok := authkit.UserIDFromContext(r.Context())
	if !ok {
		log.Warn("failed to get user id from context")
		httpx.WriteEmptyError(w, http.StatusUnauthorized)
		return
	}

	var req types.UnsubscribeFromUserJSONRequestBody
	if err := httpx.DecodeJSON(r, &req); err != nil {
		log.Warn("failed to decode unsubscribe request", slog.Any("error", err))
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCannotDecodeRequestBody)
		return
	}

	err := h.userService.Unsubscribe(r.Context(), userID, req.TargetUserId)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrUserNotFound):
			httpx.WriteEmptyError(w, http.StatusNotFound)
		case errors.Is(err, domain.ErrNotSubscribed):
			httpx.WriteEmptyError(w, http.StatusConflict)
		default:
			log.Error("failed to unsubscribe", slog.String("user_id", userID.String()), slog.Any("error", err))
			httpx.WriteEmptyError(w, http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handlers) HandleGetSubscriptions(w http.ResponseWriter, r *http.Request) {
	log := httplog.LogFrom(r.Context(), slog.Default())

	userID, ok := authkit.UserIDFromContext(r.Context())
	if !ok {
		log.Warn("failed to get user id from context")
		httpx.WriteEmptyError(w, http.StatusUnauthorized)
		return
	}

	users, err := h.userService.GetSubscriptions(r.Context(), userID)
	if err != nil {
		log.Error("failed to get subscriptions", slog.String("user_id", userID.String()), slog.Any("error", err))
		httpx.WriteEmptyError(w, http.StatusInternalServerError)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, makeUsersResponse(users))
}

func (h *Handlers) HandleGetSubscriptionsByID(w http.ResponseWriter, r *http.Request) {
	log := httplog.LogFrom(r.Context(), slog.Default())

	if _, ok := authkit.UserIDFromContext(r.Context()); !ok {
		log.Warn("failed to get user id from context")
		httpx.WriteEmptyError(w, http.StatusUnauthorized)
		return
	}

	targetUserID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		log.Warn("invalid target user id", slog.Any("error", err))
		httpx.WriteErrorStr(w, http.StatusBadRequest, "invalid user id")
		return
	}

	if _, err = h.userService.GetUser(r.Context(), targetUserID); err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			log.Info("target user not found", slog.String("target_user_id", targetUserID.String()))
			httpx.WriteEmptyError(w, http.StatusNotFound)
			return
		}

		log.Error("failed to get target user", slog.String("target_user_id", targetUserID.String()), slog.String("error", err.Error()))
		httpx.WriteEmptyError(w, http.StatusInternalServerError)
		return
	}

	users, err := h.userService.GetSubscriptions(r.Context(), targetUserID)
	if err != nil {
		log.Error("failed to get target user subscriptions", slog.String("target_user_id", targetUserID.String()), slog.Any("error", err))
		httpx.WriteEmptyError(w, http.StatusInternalServerError)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, makeUsersResponse(users))
}

func (h *Handlers) HandleGetSubscribers(w http.ResponseWriter, r *http.Request) {
	log := httplog.LogFrom(r.Context(), slog.Default())

	userID, ok := authkit.UserIDFromContext(r.Context())
	if !ok {
		log.Warn("failed to get user id from context")
		httpx.WriteEmptyError(w, http.StatusUnauthorized)
		return
	}

	users, err := h.userService.GetSubscribers(r.Context(), userID)
	if err != nil {
		log.Error("failed to get subscribers", slog.String("user_id", userID.String()), slog.Any("error", err))
		httpx.WriteEmptyError(w, http.StatusInternalServerError)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, makeUsersResponse(users))
}

func (h *Handlers) HandleGetSubscribersByID(w http.ResponseWriter, r *http.Request) {
	log := httplog.LogFrom(r.Context(), slog.Default())

	if _, ok := authkit.UserIDFromContext(r.Context()); !ok {
		log.Warn("failed to get user id from context")
		httpx.WriteEmptyError(w, http.StatusUnauthorized)
		return
	}

	targetUserID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		log.Warn("invalid target user id", slog.Any("error", err))
		httpx.WriteErrorStr(w, http.StatusBadRequest, "invalid user id")
		return
	}

	if _, err = h.userService.GetUser(r.Context(), targetUserID); err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			log.Info("target user not found", slog.String("target_user_id", targetUserID.String()))
			httpx.WriteEmptyError(w, http.StatusNotFound)
			return
		}

		log.Error("failed to get target user", slog.String("target_user_id", targetUserID.String()), slog.String("error", err.Error()))
		httpx.WriteEmptyError(w, http.StatusInternalServerError)
		return
	}

	users, err := h.userService.GetSubscribers(r.Context(), targetUserID)
	if err != nil {
		log.Error("failed to get target user subscribers", slog.String("target_user_id", targetUserID.String()), slog.Any("error", err))
		httpx.WriteEmptyError(w, http.StatusInternalServerError)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, makeUsersResponse(users))
}

func handleUpdateError(w http.ResponseWriter, log *slog.Logger, err error, userID uuid.UUID) {
	updateErrLog := log.With(slog.Any("userID", userID), slog.Any("error", err))

	if errors.Is(err, domain.ErrUserNotFound) {
		updateErrLog.Info("user not found")
		httpx.WriteEmptyError(w, http.StatusNotFound)
	} else if errors.Is(err, domain.ErrInvalidPhoneNumber) {
		updateErrLog.Info("invalid phone number")
		httpx.WriteErrorStr(w, http.StatusBadRequest, err.Error())
	} else {
		updateErrLog.Error("failed to update user")
		httpx.WriteEmptyError(w, http.StatusInternalServerError)
	}
}

// ================================================================================
// helpers
// ================================================================================

func (h *Handlers) getMe(ctx context.Context, userID uuid.UUID) (types.Me, error) {
	me, err := h.userService.GetMe(ctx, userID)
	if err != nil {
		return types.Me{}, err
	}

	return types.Me{
		Id:               me.Id,
		Name:             me.Name,
		Bio:              me.Bio,
		AvatarUrl:        me.AvatarURL,
		PhoneNumber:      me.PhoneNumber,
		CurrentLatitude:  me.CurrentLatitude,
		CurrentLongitude: me.CurrentLongitude,
		Email:            openapi_types.Email(me.Email), // TODO: конвертировать при отключенном bypass
		CreatedAt:        me.CreatedAt,
		IsAdmin:          me.IsAdmin,
		ReputationPoints: me.ReputationPoints,
	}, nil
}

func readAvatarUpload(file multipart.File) ([]byte, string, error) {
	content, err := io.ReadAll(io.LimitReader(file, maxAvatarUploadSize+1))
	if err != nil {
		return nil, "", errors.New("failed to read avatar")
	}
	if len(content) == 0 {
		return nil, "", errors.New("avatar file is empty")
	}
	if len(content) > maxAvatarUploadSize {
		return nil, "", errors.New("avatar file exceeds 5 MB")
	}

	contentType := http.DetectContentType(content)
	if !strings.HasPrefix(contentType, "image/") {
		return nil, "", errors.New("avatar file must be an image")
	}

	return content, contentType, nil
}

func makeUsersResponse(users []domain.User) types.GetSubscriptionsResponse {
	response := make(types.GetSubscriptionsResponse, 0, len(users))
	for _, u := range users {
		response = append(response, types.User{
			Id:          u.Id,
			Name:        u.Name,
			Bio:         u.Bio,
			AvatarUrl:   u.AvatarURL,
			PhoneNumber: u.PhoneNumber,
		})
	}

	return response
}
