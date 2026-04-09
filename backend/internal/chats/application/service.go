package application

import (
	"barter-port/internal/chats/domain"
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/net/context"
)

type ChatsRepository interface {
	CreateChat(ctx context.Context, dealID *uuid.UUID, participantIDs []uuid.UUID) (*domain.Chat, error)
	GetDealChatID(ctx context.Context, dealID uuid.UUID) (uuid.UUID, error)
	GetChatByID(ctx context.Context, chatID uuid.UUID) (*domain.Chat, error)
	ListChatsForUser(ctx context.Context, userID uuid.UUID) ([]domain.Chat, error)
	IsParticipant(ctx context.Context, chatID, userID uuid.UUID) (bool, error)
	SendMessage(ctx context.Context, chatID, senderID uuid.UUID, content string) (*domain.Message, error)
	GetMessages(ctx context.Context, chatID uuid.UUID, after *time.Time) ([]domain.Message, error)
}

type Service struct {
	repo ChatsRepository
}

func NewService(repo ChatsRepository) *Service {
	return &Service{repo: repo}
}

// CreateChat creates a new chat between participants. For deal chats, dealID is non-nil.
func (s *Service) CreateChat(ctx context.Context, dealID *uuid.UUID, participantIDs []uuid.UUID) (*domain.Chat, error) {
	chat, err := s.repo.CreateChat(ctx, dealID, participantIDs)
	if err != nil {
		return nil, fmt.Errorf("repo.CreateChat: %w", err)
	}
	return chat, nil
}

// GetDealChatID returns the chat ID associated with a deal.
func (s *Service) GetDealChatID(ctx context.Context, dealID uuid.UUID) (uuid.UUID, error) {
	chatID, err := s.repo.GetDealChatID(ctx, dealID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("repo.GetDealChatID: %w", err)
	}
	return chatID, nil
}

// ListChatsForUser returns all chats where the user participates.
func (s *Service) ListChatsForUser(ctx context.Context, userID uuid.UUID) ([]domain.Chat, error) {
	chats, err := s.repo.ListChatsForUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("repo.ListChatsForUser: %w", err)
	}
	return chats, nil
}

// GetMessages returns messages in a chat for an authenticated participant.
func (s *Service) GetMessages(ctx context.Context, userID, chatID uuid.UUID, after *time.Time) ([]domain.Message, error) {
	ok, err := s.repo.IsParticipant(ctx, chatID, userID)
	if err != nil {
		return nil, fmt.Errorf("repo.IsParticipant: %w", err)
	}
	if !ok {
		return nil, domain.ErrForbidden
	}

	msgs, err := s.repo.GetMessages(ctx, chatID, after)
	if err != nil {
		return nil, fmt.Errorf("repo.GetMessages: %w", err)
	}
	return msgs, nil
}

// SendMessage sends a message in a chat if the user is a participant.
func (s *Service) SendMessage(ctx context.Context, userID, chatID uuid.UUID, content string) (*domain.Message, error) {
	ok, err := s.repo.IsParticipant(ctx, chatID, userID)
	if err != nil {
		return nil, fmt.Errorf("repo.IsParticipant: %w", err)
	}
	if !ok {
		return nil, domain.ErrForbidden
	}

	msg, err := s.repo.SendMessage(ctx, chatID, userID, content)
	if err != nil {
		return nil, fmt.Errorf("repo.SendMessage: %w", err)
	}
	return msg, nil
}
