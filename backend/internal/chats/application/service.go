package application

import (
	dealspb "barter-port/contracts/grpc/deals/v1"
	"barter-port/internal/chats/domain"
	"barter-port/pkg/authkit"
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
	repo         ChatsRepository
	dealsClient  dealspb.DealsServiceClient
	adminChecker *authkit.AdminChecker
}

func NewService(repo ChatsRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) WithAdminChecker(checker *authkit.AdminChecker) *Service {
	s.adminChecker = checker
	return s
}

func (s *Service) WithDealsClient(client dealspb.DealsServiceClient) *Service {
	s.dealsClient = client
	return s
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

// GetDealChat returns the deal chat if the user is a participant or an admin.
func (s *Service) GetDealChat(ctx context.Context, userID, dealID uuid.UUID) (*domain.Chat, error) {
	chatID, err := s.repo.GetDealChatID(ctx, dealID)
	if err != nil {
		return nil, fmt.Errorf("repo.GetDealChatID: %w", err)
	}

	ok, err := s.repo.IsParticipant(ctx, chatID, userID)
	if err != nil {
		return nil, fmt.Errorf("repo.IsParticipant: %w", err)
	}
	if !ok {
		if s.adminChecker == nil {
			return nil, domain.ErrForbidden
		}

		isAdmin, adminErr := s.adminChecker.IsAdmin(ctx, userID)
		if adminErr != nil {
			return nil, fmt.Errorf("adminChecker.IsAdmin: %w", adminErr)
		}
		if !isAdmin {
			return nil, domain.ErrForbidden
		}
	}

	chat, err := s.repo.GetChatByID(ctx, chatID)
	if err != nil {
		return nil, fmt.Errorf("repo.GetChatByID: %w", err)
	}

	return chat, nil
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
		if s.adminChecker == nil {
			return nil, domain.ErrForbidden
		}

		isAdmin, adminErr := s.adminChecker.IsAdmin(ctx, userID)
		if adminErr != nil {
			return nil, fmt.Errorf("adminChecker.IsAdmin: %w", adminErr)
		}
		if !isAdmin {
			return nil, domain.ErrForbidden
		}
	}

	msgs, err := s.repo.GetMessages(ctx, chatID, after)
	if err != nil {
		return nil, fmt.Errorf("repo.GetMessages: %w", err)
	}
	return msgs, nil
}

// SendMessage sends a message in a chat if the user is a participant and the chat is writable.
// Personal chats are always writable. Deal chats become read-only when the deal reaches
// a final status (Completed, Cancelled, Failed).
func (s *Service) SendMessage(ctx context.Context, userID, chatID uuid.UUID, content string) (*domain.Message, error) {
	ok, err := s.repo.IsParticipant(ctx, chatID, userID)
	if err != nil {
		return nil, fmt.Errorf("repo.IsParticipant: %w", err)
	}
	if !ok {
		return nil, domain.ErrForbidden
	}

	chat, err := s.repo.GetChatByID(ctx, chatID)
	if err != nil {
		return nil, fmt.Errorf("repo.GetChatByID: %w", err)
	}

	writable, err := s.isChatWritable(ctx, chat)
	if err != nil {
		return nil, fmt.Errorf("isChatWritable: %w", err)
	}
	if !writable {
		return nil, domain.ErrForbidden
	}

	msg, err := s.repo.SendMessage(ctx, chatID, userID, content)
	if err != nil {
		return nil, fmt.Errorf("repo.SendMessage: %w", err)
	}
	return msg, nil
}

func (s *Service) isChatWritable(ctx context.Context, chat *domain.Chat) (bool, error) {
	if chat.DealID == nil {
		return true, nil
	}

	if s.dealsClient == nil {
		return false, fmt.Errorf("deals grpc client is not configured")
	}

	resp, err := s.dealsClient.GetDealStatus(ctx, &dealspb.GetDealStatusRequest{DealId: chat.DealID.String()})
	if err != nil {
		return false, fmt.Errorf("dealsClient.GetDealStatus: %w", err)
	}

	switch resp.GetStatus() {
	case "Completed", "Cancelled", "Failed":
		return false, nil
	default:
		return true, nil
	}
}
