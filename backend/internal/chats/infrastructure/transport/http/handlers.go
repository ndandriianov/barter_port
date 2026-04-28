package http

import (
	userspb "barter-port/contracts/grpc/users/v1"
	"barter-port/contracts/openapi/chats/types"
	"barter-port/internal/chats/application"
	"barter-port/internal/chats/domain"
	"barter-port/pkg/authkit"
	"barter-port/pkg/httpx"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Handlers struct {
	log          *slog.Logger
	chatsService *application.Service
	usersClient  userspb.UsersServiceClient
}

func NewHandlers(log *slog.Logger, chatsService *application.Service, usersClient userspb.UsersServiceClient) *Handlers {
	return &Handlers{
		log:          log,
		chatsService: chatsService,
		usersClient:  usersClient,
	}
}

// ================================================================================
// POST /chats — create direct chat
// ================================================================================

func (h *Handlers) CreateChat(w http.ResponseWriter, r *http.Request) {
	log := h.log.With(slog.String("handler", "CreateChat"))

	userID, ok := authkit.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteEmptyError(w, http.StatusUnauthorized)
		return
	}

	var req types.CreateChatRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCannotDecodeRequestBody)
		return
	}

	checkResp, err := h.usersClient.CheckSubscription(r.Context(), &userspb.CheckSubscriptionRequest{
		RequesterUserId: userID.String(),
		TargetUserId:    req.ParticipantId.String(),
	})
	if err != nil {
		log.Error("error checking subscription in users service", slog.Any("error", err))
		httpx.WriteEmptyError(w, http.StatusInternalServerError)
		return
	}
	if !checkResp.GetIsSubscribed() {
		httpx.WriteEmptyError(w, http.StatusForbidden)
		return
	}

	chat, err := h.chatsService.CreateChat(r.Context(), nil, []uuid.UUID{userID, req.ParticipantId})
	if err != nil {
		if errors.Is(err, domain.ErrChatAlreadyExists) {
			httpx.WriteEmptyError(w, http.StatusConflict)
			return
		}
		log.Error("error creating chat", slog.Any("error", err))
		httpx.WriteEmptyError(w, http.StatusInternalServerError)
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, mapChatToResp(chat))
}

// ================================================================================
// GET /chats — list user's chats
// ================================================================================

func (h *Handlers) ListChats(w http.ResponseWriter, r *http.Request) {
	log := h.log.With(slog.String("handler", "ListChats"))

	userID, ok := authkit.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteEmptyError(w, http.StatusUnauthorized)
		return
	}

	chats, err := h.chatsService.ListChatsForUser(r.Context(), userID)
	if err != nil {
		log.Error("error listing chats", slog.Any("error", err))
		httpx.WriteEmptyError(w, http.StatusInternalServerError)
		return
	}

	resp := make([]types.Chat, len(chats))
	for i := range chats {
		resp[i] = mapChatToResp(&chats[i])
	}
	httpx.WriteJSON(w, http.StatusOK, resp)
}

// ================================================================================
// GET /chats/deals/{dealId} — get deal chat metadata
// ================================================================================

func (h *Handlers) GetDealChat(w http.ResponseWriter, r *http.Request) {
	log := h.log.With(slog.String("handler", "GetDealChat"))

	userID, ok := authkit.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteEmptyError(w, http.StatusUnauthorized)
		return
	}

	dealID, err := uuid.Parse(chi.URLParam(r, "dealId"))
	if err != nil {
		httpx.WriteEmptyError(w, http.StatusBadRequest)
		return
	}

	chat, err := h.chatsService.GetDealChat(r.Context(), userID, dealID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrForbidden):
			httpx.WriteEmptyError(w, http.StatusForbidden)
			return
		case errors.Is(err, domain.ErrChatNotFound):
			httpx.WriteEmptyError(w, http.StatusNotFound)
			return
		default:
			log.Error("error getting deal chat", slog.Any("error", err))
			httpx.WriteEmptyError(w, http.StatusInternalServerError)
			return
		}
	}

	httpx.WriteJSON(w, http.StatusOK, mapChatToResp(chat))
}

// ================================================================================
// GET /chats/{chatId}/messages — get messages (polling)
// ================================================================================

func (h *Handlers) GetMessages(w http.ResponseWriter, r *http.Request) {
	log := h.log.With(slog.String("handler", "GetMessages"))

	userID, ok := authkit.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteEmptyError(w, http.StatusUnauthorized)
		return
	}

	chatID, err := uuid.Parse(chi.URLParam(r, "chatId"))
	if err != nil {
		httpx.WriteEmptyError(w, http.StatusBadRequest)
		return
	}

	var after *time.Time
	if afterStr := r.URL.Query().Get("after"); afterStr != "" {
		t, err := time.Parse(time.RFC3339Nano, afterStr)
		if err != nil {
			httpx.WriteEmptyError(w, http.StatusBadRequest)
			return
		}
		after = &t
	}

	msgs, err := h.chatsService.GetMessages(r.Context(), userID, chatID, after)
	if err != nil {
		if errors.Is(err, domain.ErrForbidden) {
			httpx.WriteEmptyError(w, http.StatusForbidden)
			return
		}
		log.Error("error getting messages", slog.Any("error", err))
		httpx.WriteEmptyError(w, http.StatusInternalServerError)
		return
	}

	resp := make([]types.Message, len(msgs))
	for i := range msgs {
		resp[i] = mapMessageToResp(&msgs[i])
	}
	httpx.WriteJSON(w, http.StatusOK, resp)
}

// ================================================================================
// POST /chats/{chatId}/messages — send message
// ================================================================================

type sendMessageReq struct {
	Content string `json:"content"`
}

func (h *Handlers) SendMessage(w http.ResponseWriter, r *http.Request) {
	log := h.log.With(slog.String("handler", "SendMessage"))

	userID, ok := authkit.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteEmptyError(w, http.StatusUnauthorized)
		return
	}

	chatID, err := uuid.Parse(chi.URLParam(r, "chatId"))
	if err != nil {
		httpx.WriteEmptyError(w, http.StatusBadRequest)
		return
	}

	var req sendMessageReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCannotDecodeRequestBody)
		return
	}

	if req.Content == "" {
		httpx.WriteEmptyError(w, http.StatusBadRequest)
		return
	}

	msg, err := h.chatsService.SendMessage(r.Context(), userID, chatID, req.Content)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrChatPendingFailureReview), errors.Is(err, domain.ErrChatWriteForbidden), errors.Is(err, domain.ErrForbidden):
			httpx.WriteError(w, http.StatusForbidden, err)
			return
		}
		log.Error("error sending message", slog.Any("error", err))
		httpx.WriteEmptyError(w, http.StatusInternalServerError)
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, mapMessageToResp(msg))
}

func (h *Handlers) GetAdminPlatformStatistics(w http.ResponseWriter, r *http.Request) {
	userID, ok := authkit.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteEmptyError(w, http.StatusUnauthorized)
		return
	}

	total, err := h.chatsService.GetAdminPlatformChatsCount(r.Context(), userID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrForbidden):
			httpx.WriteEmptyError(w, http.StatusForbidden)
		default:
			h.log.Error("error getting admin chats platform statistics", slog.Any("error", err))
			httpx.WriteEmptyError(w, http.StatusInternalServerError)
		}
		return
	}

	httpx.WriteJSON(w, http.StatusOK, types.AdminPlatformStatistics{
		Chats: types.AdminChatStatistics{Total: total},
	})
}

// ================================================================================
// GET /users — list users (for starting a new chat)
// ================================================================================

type userResp struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (h *Handlers) ListUsers(w http.ResponseWriter, r *http.Request) {
	log := h.log.With(slog.String("handler", "ListUsers"))

	id, ok := authkit.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteEmptyError(w, http.StatusUnauthorized)
		return
	}

	req := userspb.ListUsersForChatCreationRequest{RequesterUserId: id.String()}

	res, err := h.usersClient.ListUsersForChatCreation(r.Context(), &req)
	if err != nil {
		log.Error("error listing users from users service", slog.Any("error", err))
		httpx.WriteEmptyError(w, http.StatusInternalServerError)
		return
	}

	resp := make([]userResp, len(res.Users))
	for i, u := range res.Users {
		resp[i] = userResp{ID: u.Id, Name: u.Name}
	}
	httpx.WriteJSON(w, http.StatusOK, resp)
}
