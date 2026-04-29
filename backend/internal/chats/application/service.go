package application

import (
	dealspb "barter-port/contracts/grpc/deals/v1"
	userspb "barter-port/contracts/grpc/users/v1"
	"barter-port/internal/chats/domain"
	"barter-port/pkg/authkit"
	"errors"
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
	CountChats(ctx context.Context) (int, error)
	IsParticipant(ctx context.Context, chatID, userID uuid.UUID) (bool, error)
	SendMessage(ctx context.Context, chatID, senderID uuid.UUID, content string) (*domain.Message, error)
	GetMessages(ctx context.Context, chatID uuid.UUID, after *time.Time) ([]domain.Message, error)
}

type Service struct {
	repo         ChatsRepository
	dealsClient  dealspb.DealsServiceClient
	usersClient  userspb.UsersServiceClient
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

func (s *Service) WithUsersClient(client userspb.UsersServiceClient) *Service {
	s.usersClient = client
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

	info, err := s.usersClient.GetUsersWithInfo(ctx, &userspb.GetUsersWithInfoRequest{Ids: chat.GetParticipantIdsToString()})
	if err != nil {
		return nil, fmt.Errorf("usersClient.GetUsersWithInfo: %w", err)
	}
	chat.Participants = chatParticipantsFromUserInfo(info.GetUsers())

	return chat, nil
}

// ListChatsForUser returns all chats where the user participates.
func (s *Service) ListChatsForUser(ctx context.Context, userID uuid.UUID) ([]domain.Chat, error) {
	chats, err := s.repo.ListChatsForUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("repo.ListChatsForUser: %w", err)
	}

	for i, chat := range chats {
		info, err := s.usersClient.GetUsersWithInfo(ctx, &userspb.GetUsersWithInfoRequest{Ids: chat.GetParticipantIdsToString()})
		if err != nil {
			return nil, fmt.Errorf("usersClient.GetUsersWithInfo: %w", err)
		}

		chats[i].Participants = chatParticipantsFromUserInfo(info.GetUsers())
	}

	return chats, nil
}

func (s *Service) GetAdminPlatformChatsCount(ctx context.Context, requesterID uuid.UUID) (int, error) {
	if s.adminChecker == nil {
		return 0, fmt.Errorf("admin checker is not configured")
	}

	isAdmin, err := s.adminChecker.IsAdmin(ctx, requesterID)
	if err != nil {
		return 0, fmt.Errorf("adminChecker.IsAdmin: %w", err)
	}
	if !isAdmin {
		return 0, domain.ErrForbidden
	}

	count, err := s.repo.CountChats(ctx)
	if err != nil {
		return 0, fmt.Errorf("repo.CountChats: %w", err)
	}

	return count, nil
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
// a final status (Completed, Cancelled, Failed) or has a pending admin failure review.
func (s *Service) SendMessage(ctx context.Context, userID, chatID uuid.UUID, content string) (*domain.Message, error) {
	ok, err := s.repo.IsParticipant(ctx, chatID, userID)
	if err != nil {
		return nil, fmt.Errorf("repo.IsParticipant: %w", err)
	}
	if !ok {
		return nil, domain.NewUserMessageError(domain.ErrForbidden, "Нельзя отправить сообщение в чат, в котором вы не участвуете")
	}

	chat, err := s.repo.GetChatByID(ctx, chatID)
	if err != nil {
		return nil, fmt.Errorf("repo.GetChatByID: %w", err)
	}

	if err := s.ensureChatWritable(ctx, chat); err != nil {
		if errors.Is(err, domain.ErrChatPendingFailureReview) || errors.Is(err, domain.ErrChatWriteForbidden) {
			return nil, err
		}
		return nil, fmt.Errorf("ensureChatWritable: %w", err)
	}

	msg, err := s.repo.SendMessage(ctx, chatID, userID, content)
	if err != nil {
		return nil, fmt.Errorf("repo.SendMessage: %w", err)
	}
	return msg, nil
}

func (s *Service) ensureChatWritable(ctx context.Context, chat *domain.Chat) error {
	if chat.DealID == nil {
		return nil
	}

	if s.dealsClient == nil {
		return fmt.Errorf("deals grpc client is not configured")
	}

	resp, err := s.dealsClient.GetDealStatus(ctx, &dealspb.GetDealStatusRequest{DealId: chat.DealID.String()})
	if err != nil {
		return fmt.Errorf("dealsClient.GetDealStatus: %w", err)
	}

	if resp.GetHasPendingFailureReview() {
		return domain.ErrChatPendingFailureReview
	}

	switch resp.GetStatus() {
	case "Completed", "Cancelled", "Failed":
		return domain.NewUserMessageError(
			domain.ErrChatWriteForbidden,
			fmt.Sprintf("Сделка находится в статусе %s", localizeDealStatus(resp.GetStatus())),
		)
	default:
		return nil
	}
}

func localizeDealStatus(status string) string {
	switch status {
	case "Completed":
		return "Completed (завершена)"
	case "Cancelled":
		return "Cancelled (отменена)"
	case "Failed":
		return "Failed (провалена)"
	default:
		return status
	}
}

func chatParticipantsFromUserInfo(userInfo []*userspb.UserInfo) []domain.ChatParticipant {
	participants := make([]domain.ChatParticipant, len(userInfo))
	for i, info := range userInfo {
		var name *string
		if info.GetName() != "" {
			name = &info.Name
		}
		participants[i] = domain.ChatParticipant{
			ID:   uuid.MustParse(info.GetId()),
			Name: name,
		}
	}
	return participants
}
