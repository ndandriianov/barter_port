package http

import (
	userspb "barter-port/contracts/grpc/users/v1"
	"barter-port/internal/chats/application"
	"barter-port/internal/chats/domain"
	"barter-port/pkg/authkit"
	"bytes"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
)

type testChatsRepo struct {
	createChatFn     func(ctx context.Context, dealID *uuid.UUID, participantIDs []uuid.UUID) (*domain.Chat, error)
	createChatCalled bool
}

func (r *testChatsRepo) CreateChat(ctx context.Context, dealID *uuid.UUID, participantIDs []uuid.UUID) (*domain.Chat, error) {
	r.createChatCalled = true
	if r.createChatFn != nil {
		return r.createChatFn(ctx, dealID, participantIDs)
	}
	return nil, nil
}

func (r *testChatsRepo) GetDealChatID(context.Context, uuid.UUID) (uuid.UUID, error) {
	return uuid.Nil, errors.New("not implemented")
}

func (r *testChatsRepo) GetChatByID(context.Context, uuid.UUID) (*domain.Chat, error) {
	return nil, errors.New("not implemented")
}

func (r *testChatsRepo) ListChatsForUser(context.Context, uuid.UUID) ([]domain.Chat, error) {
	return nil, errors.New("not implemented")
}

func (r *testChatsRepo) CountChats(context.Context) (int, error) {
	return 0, errors.New("not implemented")
}

func (r *testChatsRepo) IsParticipant(context.Context, uuid.UUID, uuid.UUID) (bool, error) {
	return false, errors.New("not implemented")
}

func (r *testChatsRepo) SendMessage(context.Context, uuid.UUID, uuid.UUID, string) (*domain.Message, error) {
	return nil, errors.New("not implemented")
}

func (r *testChatsRepo) GetMessages(context.Context, uuid.UUID, *time.Time) ([]domain.Message, error) {
	return nil, errors.New("not implemented")
}

type testUsersClient struct {
	checkResp *userspb.CheckSubscriptionResponse
	checkErr  error
	checkReq  *userspb.CheckSubscriptionRequest
}

func (c *testUsersClient) GetUsersWithInfo(context.Context, *userspb.GetUsersWithInfoRequest, ...grpc.CallOption) (*userspb.GetUsersWithInfoResponse, error) {
	return nil, errors.New("not implemented")
}

func (c *testUsersClient) ListUsers(context.Context, *userspb.ListUsersRequest, ...grpc.CallOption) (*userspb.ListUsersResponse, error) {
	return nil, errors.New("not implemented")
}

func (c *testUsersClient) ListUsersForChatCreation(context.Context, *userspb.ListUsersForChatCreationRequest, ...grpc.CallOption) (*userspb.ListUsersForChatCreationResponse, error) {
	return nil, errors.New("not implemented")
}

func (c *testUsersClient) CheckSubscription(_ context.Context, req *userspb.CheckSubscriptionRequest, _ ...grpc.CallOption) (*userspb.CheckSubscriptionResponse, error) {
	c.checkReq = req
	if c.checkErr != nil {
		return nil, c.checkErr
	}
	return c.checkResp, nil
}

func (c *testUsersClient) ListSubscriptions(context.Context, *userspb.ListSubscriptionsRequest, ...grpc.CallOption) (*userspb.ListSubscriptionsResponse, error) {
	return nil, errors.New("not implemented")
}

func (c *testUsersClient) ListHiddenUsers(context.Context, *userspb.ListHiddenUsersRequest, ...grpc.CallOption) (*userspb.ListHiddenUsersResponse, error) {
	return nil, errors.New("not implemented")
}

func (c *testUsersClient) IsUserHiddenByAnyOwners(context.Context, *userspb.IsUserHiddenByAnyOwnersRequest, ...grpc.CallOption) (*userspb.IsUserHiddenByAnyOwnersResponse, error) {
	return nil, errors.New("not implemented")
}

func (c *testUsersClient) GetUserLocation(context.Context, *userspb.GetUserLocationRequest, ...grpc.CallOption) (*userspb.GetUserLocationResponse, error) {
	return nil, errors.New("not implemented")
}

func TestCreateChat_CheckSubscriptionSuccess(t *testing.T) {
	requesterID := uuid.New()
	targetID := uuid.New()
	chatID := uuid.New()

	repo := &testChatsRepo{
		createChatFn: func(_ context.Context, dealID *uuid.UUID, participantIDs []uuid.UUID) (*domain.Chat, error) {
			if dealID != nil {
				t.Fatalf("expected direct chat with nil dealID")
			}
			if len(participantIDs) != 2 || participantIDs[0] != requesterID || participantIDs[1] != targetID {
				t.Fatalf("unexpected participants: %v", participantIDs)
			}
			return &domain.Chat{ID: chatID, Participants: domain.NewChatParticipantsWithoutNames(participantIDs), CreatedAt: time.Now()}, nil
		},
	}
	usersClient := &testUsersClient{checkResp: &userspb.CheckSubscriptionResponse{IsSubscribed: true}}
	h := NewHandlers(slog.Default(), application.NewService(repo), usersClient)

	req := httptest.NewRequest(http.MethodPost, "/chats", bytes.NewBufferString(`{"participant_id":"`+targetID.String()+`"}`))
	req = req.WithContext(authkit.WithUserID(req.Context(), requesterID))
	rr := httptest.NewRecorder()

	h.CreateChat(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rr.Code)
	}
	if usersClient.checkReq == nil {
		t.Fatal("expected CheckSubscription to be called")
	}
	if usersClient.checkReq.GetRequesterUserId() != requesterID.String() {
		t.Fatalf("unexpected requester_user_id: %s", usersClient.checkReq.GetRequesterUserId())
	}
	if usersClient.checkReq.GetTargetUserId() != targetID.String() {
		t.Fatalf("unexpected target_user_id: %s", usersClient.checkReq.GetTargetUserId())
	}
	if !repo.createChatCalled {
		t.Fatal("expected chatsService.CreateChat to be called")
	}
}

func TestCreateChat_CheckSubscriptionForbidden(t *testing.T) {
	requesterID := uuid.New()
	targetID := uuid.New()

	repo := &testChatsRepo{}
	usersClient := &testUsersClient{checkResp: &userspb.CheckSubscriptionResponse{IsSubscribed: false}}
	h := NewHandlers(slog.Default(), application.NewService(repo), usersClient)

	req := httptest.NewRequest(http.MethodPost, "/chats", bytes.NewBufferString(`{"participant_id":"`+targetID.String()+`"}`))
	req = req.WithContext(authkit.WithUserID(req.Context(), requesterID))
	rr := httptest.NewRecorder()

	h.CreateChat(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, rr.Code)
	}
	if repo.createChatCalled {
		t.Fatal("did not expect chatsService.CreateChat to be called")
	}
}

func TestCreateChat_CheckSubscriptionError(t *testing.T) {
	requesterID := uuid.New()
	targetID := uuid.New()

	repo := &testChatsRepo{}
	usersClient := &testUsersClient{checkErr: errors.New("users rpc down")}
	h := NewHandlers(slog.Default(), application.NewService(repo), usersClient)

	req := httptest.NewRequest(http.MethodPost, "/chats", bytes.NewBufferString(`{"participant_id":"`+targetID.String()+`"}`))
	req = req.WithContext(authkit.WithUserID(req.Context(), requesterID))
	rr := httptest.NewRecorder()

	h.CreateChat(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rr.Code)
	}
	if repo.createChatCalled {
		t.Fatal("did not expect chatsService.CreateChat to be called")
	}
}

func TestCreateChat_AlreadyExists(t *testing.T) {
	requesterID := uuid.New()
	targetID := uuid.New()

	repo := &testChatsRepo{
		createChatFn: func(context.Context, *uuid.UUID, []uuid.UUID) (*domain.Chat, error) {
			return nil, domain.ErrChatAlreadyExists
		},
	}
	usersClient := &testUsersClient{checkResp: &userspb.CheckSubscriptionResponse{IsSubscribed: true}}
	h := NewHandlers(slog.Default(), application.NewService(repo), usersClient)

	req := httptest.NewRequest(http.MethodPost, "/chats", bytes.NewBufferString(`{"participant_id":"`+targetID.String()+`"}`))
	req = req.WithContext(authkit.WithUserID(req.Context(), requesterID))
	rr := httptest.NewRecorder()

	h.CreateChat(rr, req)

	if rr.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d", http.StatusConflict, rr.Code)
	}
}
