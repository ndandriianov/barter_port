package user

import (
	authpb "barter-port/contracts/grpc/auth/v1"
	"barter-port/internal/users/domain"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/google/uuid"
	"golang.org/x/net/context"
)

type UsersRepository interface {
	GetUserById(ctx context.Context, id uuid.UUID) (*domain.User, error)
	GetReputationPoints(ctx context.Context, id uuid.UUID) (int, error)
	GetReputationEvents(ctx context.Context, id uuid.UUID) ([]domain.ReputationEvent, error)
	UpdateName(ctx context.Context, id uuid.UUID, name string) error
	UpdateBio(ctx context.Context, id uuid.UUID, bio *string) error
	UpdateAvatarURL(ctx context.Context, id uuid.UUID, avatarURL *string) error
	UpdatePhoneNumber(ctx context.Context, id uuid.UUID, phoneNumber *string) error
	GetNamesForUserIDs(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]*string, error)
	ListUsers(ctx context.Context) ([]domain.User, error)
	Subscribe(ctx context.Context, subscriberID, targetUserID uuid.UUID) error
	Unsubscribe(ctx context.Context, subscriberID, targetUserID uuid.UUID) error
	IsSubscribed(ctx context.Context, subscriberID, targetUserID uuid.UUID) (bool, error)
	GetSubscriptions(ctx context.Context, userID uuid.UUID) ([]domain.User, error)
	GetSubscribers(ctx context.Context, userID uuid.UUID) ([]domain.User, error)
}

var ErrAuthClientNotConfigured = errors.New("auth grpc client is not configured")
var ErrAvatarStorageNotConfigured = errors.New("avatar storage is not configured")

var phoneNumberRegexp = regexp.MustCompile(`^\+7 \(\d{3}\) \d{3}-\d{2}-\d{2}$`)

type AvatarStorage interface {
	UploadAvatar(ctx context.Context, userID uuid.UUID, contentType string, content []byte) (string, error)
	DeleteAvatar(ctx context.Context, avatarURL string) error
	IsManagedAvatarURL(avatarURL string) bool
}

type Me struct {
	Id               uuid.UUID
	Name             *string
	Bio              *string
	AvatarURL        *string
	PhoneNumber      *string
	Email            string
	CreatedAt        time.Time
	IsAdmin          bool
	ReputationPoints int
}

type Service struct {
	repository    UsersRepository
	authClient    authpb.AuthServiceClient
	avatarStorage AvatarStorage
}

func NewService(repository UsersRepository, authClient authpb.AuthServiceClient, avatarStorage AvatarStorage) *Service {
	return &Service{repository: repository, authClient: authClient, avatarStorage: avatarStorage}
}

func (s *Service) GetUser(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return s.repository.GetUserById(ctx, id)
}

func (s *Service) GetMe(ctx context.Context, id uuid.UUID) (Me, error) {
	u, err := s.repository.GetUserById(ctx, id)
	if err != nil {
		return Me{}, err
	}

	if s.authClient == nil {
		return Me{}, ErrAuthClientNotConfigured
	}

	authMe, err := s.authClient.GetMe(ctx, &authpb.GetMeRequest{Id: id.String()})
	if err != nil {
		return Me{}, err
	}

	reputationPoints, err := s.repository.GetReputationPoints(ctx, id)
	if err != nil {
		return Me{}, err
	}

	var createdAt time.Time
	if ts := authMe.GetCreatedAt(); ts != nil {
		createdAt = ts.AsTime()
	}

	return Me{
		Id:               u.Id,
		Name:             u.Name,
		Bio:              u.Bio,
		AvatarURL:        u.AvatarURL,
		PhoneNumber:      u.PhoneNumber,
		Email:            authMe.GetEmail(),
		CreatedAt:        createdAt,
		IsAdmin:          authMe.GetIsAdmin(),
		ReputationPoints: reputationPoints,
	}, nil
}

func (s *Service) GetCurrentUserReputationEvents(ctx context.Context, id uuid.UUID) ([]domain.ReputationEvent, error) {
	return s.repository.GetReputationEvents(ctx, id)
}

// UpdateName updates users name by id.
//
// Errors:
//   - domain.ErrUserNotFound: Occurs if no user is found with the given id.
func (s *Service) UpdateName(ctx context.Context, id uuid.UUID, name string) error {
	return s.repository.UpdateName(ctx, id, name)
}

// UpdateBio updates users bio by id. Bio can be null.
//
// Errors:
//   - domain.ErrUserNotFound: Occurs if no user is found with the given id.
func (s *Service) UpdateBio(ctx context.Context, id uuid.UUID, bio *string) error {
	return s.repository.UpdateBio(ctx, id, bio)
}

// UpdateAvatarURL updates the avatar URL of a user. Empty string clears the stored avatar.
//
// Errors:
//   - domain.ErrUserNotFound: Occurs if no user is found with the given id.
func (s *Service) UpdateAvatarURL(ctx context.Context, id uuid.UUID, avatarURL *string) error {
	normalizedAvatarURL := normalizeOptionalString(avatarURL)

	currentUser, err := s.repository.GetUserById(ctx, id)
	if err != nil {
		return err
	}

	if err = s.repository.UpdateAvatarURL(ctx, id, normalizedAvatarURL); err != nil {
		return err
	}

	if s.avatarStorage == nil || currentUser.AvatarURL == nil {
		return nil
	}

	if s.avatarStorage.IsManagedAvatarURL(*currentUser.AvatarURL) &&
		(normalizedAvatarURL == nil || *normalizedAvatarURL != *currentUser.AvatarURL) {
		_ = s.avatarStorage.DeleteAvatar(ctx, *currentUser.AvatarURL)
	}

	return nil
}

// UpdatePhoneNumber updates the phone number of a user. Empty string clears the stored phone number.
//
// Errors:
//   - domain.ErrUserNotFound: Occurs if no user is found with the given id.
func (s *Service) UpdatePhoneNumber(ctx context.Context, id uuid.UUID, phoneNumber *string) error {
	normalizedPhoneNumber := normalizeOptionalString(phoneNumber)
	if normalizedPhoneNumber != nil && !phoneNumberRegexp.MatchString(*normalizedPhoneNumber) {
		return domain.ErrInvalidPhoneNumber
	}

	return s.repository.UpdatePhoneNumber(ctx, id, normalizedPhoneNumber)
}

func (s *Service) UploadAvatar(ctx context.Context, id uuid.UUID, contentType string, content []byte) (string, error) {
	if _, err := s.repository.GetUserById(ctx, id); err != nil {
		return "", err
	}
	if s.avatarStorage == nil {
		return "", ErrAvatarStorageNotConfigured
	}

	return s.avatarStorage.UploadAvatar(ctx, id, contentType, content)
}

// GetNamesForUserIDs returns a map of user IDs to their corresponding names.
//
// No domain Errors
func (s *Service) GetNamesForUserIDs(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]*string, error) {
	names, err := s.repository.GetNamesForUserIDs(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("repository.GetNamesForUserIDs: %w", err)
	}
	return names, nil
}

// ListUsers returns all users.
//
// No domain Errors
func (s *Service) ListUsers(ctx context.Context) ([]domain.User, error) {
	users, err := s.repository.ListUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("repository.ListUsers: %w", err)
	}
	return users, nil
}

func normalizeOptionalString(value *string) *string {
	if value == nil {
		return nil
	}
	if *value == "" {
		return nil
	}
	return value
}

// Subscribe subscribes subscriberID to targetUserID.
//
// Errors:
//   - domain.ErrUserNotFound: target user does not exist.
//   - domain.ErrAlreadySubscribed: already subscribed.
//   - errors.New("cannot subscribe to yourself"): subscriberID == targetUserID.
func (s *Service) Subscribe(ctx context.Context, subscriberID, targetUserID uuid.UUID) error {
	if subscriberID == targetUserID {
		return domain.ErrCannotSubscribeToYourself
	}
	return s.repository.Subscribe(ctx, subscriberID, targetUserID)
}

// Unsubscribe removes the subscription of subscriberID from targetUserID.
//
// Errors:
//   - domain.ErrNotSubscribed: not subscribed.
//   - domain.ErrCannotSubscribeToYourself: subscriberID == targetUserID.
func (s *Service) Unsubscribe(ctx context.Context, subscriberID, targetUserID uuid.UUID) error {
	if subscriberID == targetUserID {
		return domain.ErrCannotSubscribeToYourself
	}
	return s.repository.Unsubscribe(ctx, subscriberID, targetUserID)
}

// GetSubscriptions returns users that userID is subscribed to.
//
// No domain Errors
func (s *Service) GetSubscriptions(ctx context.Context, userID uuid.UUID) ([]domain.User, error) {
	users, err := s.repository.GetSubscriptions(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("repository.GetSubscriptions: %w", err)
	}
	return users, nil
}

// GetSubscriptionsUserInfo returns basic info of users that userID is subscribed to.
//
// No domain Errors
func (s *Service) GetSubscriptionsUserInfo(ctx context.Context, userID uuid.UUID) ([]domain.UserInfo, error) {
	users, err := s.repository.GetSubscriptions(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("repository.GetSubscriptions: %w", err)
	}

	userInfos := make([]domain.UserInfo, len(users))
	for i, u := range users {
		userInfos[i] = u.GetInfo()
	}

	return userInfos, nil
}

// GetSubscribers returns users subscribed to userID.
//
// No domain Errors
func (s *Service) GetSubscribers(ctx context.Context, userID uuid.UUID) ([]domain.User, error) {
	users, err := s.repository.GetSubscribers(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("repository.GetSubscribers: %w", err)
	}
	return users, nil
}

// CheckSubscription проверяет, подписан ли target на requester,
//   - если нет, то isTargetSubscribed будет false, функия завершится
//   - если да, то isTargetSubscribed будет true
//
// Функция также проверяет, подписан ли requester на target
//   - если да, то hasCreatedSubscription будет false
//   - если нет, то hasCreatedSubscription будет true, так как функция подпишет requester на target
func (s *Service) CheckSubscription(ctx context.Context, requester, target uuid.UUID) (
	isTargetSubscribed bool,
	hasCreatedSubscription bool,
	err error,
) {
	isTargetSubscribed, err = s.repository.IsSubscribed(ctx, target, requester)
	if err != nil || !isTargetSubscribed {
		return isTargetSubscribed, false, err
	}

	// target is subscribed to requester — try to subscribe requester to target
	err = s.repository.Subscribe(ctx, requester, target)
	if err != nil {
		if errors.Is(err, domain.ErrAlreadySubscribed) {
			return true, false, nil
		}
		return false, false, err
	}
	return true, true, nil
}

// CanCreateChat checks if a chat can be created between requester and target.
func (s *Service) CanCreateChat(ctx context.Context, requester, target uuid.UUID) (bool, error) {
	if requester == target {
		return false, nil
	}

	isRequesterSubscribed, err := s.repository.IsSubscribed(ctx, requester, target)
	if err != nil {
		return false, fmt.Errorf("(requester) repository.IsSubscribed: %w", err)
	}
	if !isRequesterSubscribed {
		return false, nil
	}

	isTargetSubscribed, err := s.repository.IsSubscribed(ctx, target, requester)
	if err != nil {
		return false, fmt.Errorf("(target) repository.IsSubscribed: %w", err)
	}
	if !isTargetSubscribed {
		return false, nil
	}

	return true, nil
}
