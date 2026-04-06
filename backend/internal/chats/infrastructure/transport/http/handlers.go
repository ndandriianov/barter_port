package http

import (
	userspb "barter-port/contracts/grpc/users/v1"
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

type createChatReq struct {
	ParticipantID string `json:"participant_id"`
}

type chatResp struct {
	ID           string   `json:"id"`
	DealID       *string  `json:"deal_id,omitempty"`
	Participants []string `json:"participants"`
	CreatedAt    string   `json:"created_at"`
}

func (h *Handlers) CreateChat(w http.ResponseWriter, r *http.Request) {
	log := h.log.With(slog.String("handler", "CreateChat"))

	userID, ok := authkit.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteEmptyError(w, http.StatusUnauthorized)
		return
	}

	var req createChatReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCannotDecodeRequestBody)
		return
	}

	participantID, err := uuid.Parse(req.ParticipantID)
	if err != nil {
		httpx.WriteEmptyError(w, http.StatusBadRequest)
		return
	}

	chat, err := h.chatsService.CreateChat(r.Context(), nil, []uuid.UUID{userID, participantID})
	if err != nil {
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

	resp := make([]chatResp, len(chats))
	for i := range chats {
		resp[i] = mapChatToResp(&chats[i])
	}
	httpx.WriteJSON(w, http.StatusOK, resp)
}

// ================================================================================
// GET /chats/{chatId}/messages — get messages (polling)
// ================================================================================

type messageResp struct {
	ID        string `json:"id"`
	ChatID    string `json:"chat_id"`
	SenderID  string `json:"sender_id"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
}

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

	resp := make([]messageResp, len(msgs))
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
		if errors.Is(err, domain.ErrForbidden) {
			httpx.WriteEmptyError(w, http.StatusForbidden)
			return
		}
		log.Error("error sending message", slog.Any("error", err))
		httpx.WriteEmptyError(w, http.StatusInternalServerError)
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, mapMessageToResp(msg))
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

	if _, ok := authkit.UserIDFromContext(r.Context()); !ok {
		httpx.WriteEmptyError(w, http.StatusUnauthorized)
		return
	}

	res, err := h.usersClient.ListUsers(r.Context(), &userspb.ListUsersRequest{})
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

// ================================================================================
// helpers
// ================================================================================

func mapChatToResp(c *domain.Chat) chatResp {
	participants := make([]string, len(c.Participants))
	for i, p := range c.Participants {
		participants[i] = p.String()
	}
	resp := chatResp{
		ID:           c.ID.String(),
		Participants: participants,
		CreatedAt:    c.CreatedAt.Format(time.RFC3339Nano),
	}
	if c.DealID != nil {
		s := c.DealID.String()
		resp.DealID = &s
	}
	return resp
}

func mapMessageToResp(m *domain.Message) messageResp {
	return messageResp{
		ID:        m.ID.String(),
		ChatID:    m.ChatID.String(),
		SenderID:  m.SenderID.String(),
		Content:   m.Content,
		CreatedAt: m.CreatedAt.Format(time.RFC3339Nano),
	}
}
