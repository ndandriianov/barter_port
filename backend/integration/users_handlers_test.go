package integration

import (
	usertypes "barter-port/contracts/openapi/users/types"
	"barter-port/pkg/jwt"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	os.Exit(runTests(m))
}

func runTests(m *testing.M) int {
	ctx := context.Background()

	var setupErr error
	globalFixture, setupErr = newSharedFixture(ctx, FixtureOptions{
		NeedAuth:  true,
		NeedUsers: true,
		NeedItems: true,
		NeedChats: true,
	})
	if setupErr != nil {
		log.Printf("не удалось поднять fixture: %v", setupErr)
		if globalFixture != nil {
			if cleanErr := globalFixture.TerminateAll(ctx); cleanErr != nil {
				log.Printf("ошибка очистки после сбоя setup: %v", cleanErr)
			}
		}
		return 1
	}

	defer func() {
		if cleanErr := globalFixture.TerminateAll(ctx); cleanErr != nil {
			log.Printf("ошибка очистки fixture: %v", cleanErr)
		}
	}()

	return m.Run()
}

// ────────────────────────────────────────────────────────────────
// Тесты
// ────────────────────────────────────────────────────────────────

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type registerResponse struct {
	UserID uuid.UUID `json:"userId"`
	Email  string    `json:"email"`
}

func TestUsersGetMe(t *testing.T) {
	t.Parallel()

	fixture := globalFixture

	registered := registerAuthUser(t, fixture)
	waitForUsersProjection(t, fixture, registered.UserID)

	req, err := http.NewRequest(http.MethodGet, fixture.UsersURL+"/users/me", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+mustAccessToken(t, registered.UserID))

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = resp.Body.Close()
	})

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var me usertypes.Me
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&me))
	require.Equal(t, registered.UserID, uuid.UUID(me.Id))
	require.Equal(t, registered.Email, string(me.Email))
	require.Nil(t, me.Name)
	require.Nil(t, me.Bio)
	require.Nil(t, me.AvatarUrl)
	require.False(t, me.CreatedAt.IsZero())
}

func TestUsersGetCurrentUserReputationEvents(t *testing.T) {
	t.Parallel()

	fixture := globalFixture

	registered := registerAuthUser(t, fixture)
	waitForUsersProjection(t, fixture, registered.UserID)

	usersDB := OpenDatabase(t, fixture, UsersDBName)
	olderEventID := uuid.New()
	olderSourceID := uuid.New()
	olderCreatedAt := time.Now().UTC().Add(-2 * time.Minute)
	olderComment := "manual adjustment"

	_, err := usersDB.Exec(
		context.Background(),
		`INSERT INTO user_reputation_events (id, user_id, source_type, source_id, delta, created_at, comment)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		olderEventID,
		registered.UserID,
		"offerreport",
		olderSourceID,
		-5,
		olderCreatedAt,
		olderComment,
	)
	require.NoError(t, err)

	newerEventID := uuid.New()
	newerSourceID := uuid.New()
	newerCreatedAt := time.Now().UTC().Add(-1 * time.Minute)

	_, err = usersDB.Exec(
		context.Background(),
		`INSERT INTO user_reputation_events (id, user_id, source_type, source_id, delta, created_at, comment)
		 VALUES ($1, $2, $3, $4, $5, $6, NULL)`,
		newerEventID,
		registered.UserID,
		"offerreport",
		newerSourceID,
		10,
		newerCreatedAt,
	)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodGet, fixture.UsersURL+"/users/reputation-events", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+mustAccessToken(t, registered.UserID))

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = resp.Body.Close()
	})

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var events usertypes.GetReputationEventsResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&events))
	require.Len(t, events, 2)

	require.Equal(t, newerEventID, uuid.UUID(events[0].Id))
	require.Equal(t, "offerreport", events[0].SourceType)
	require.Equal(t, newerSourceID, uuid.UUID(events[0].SourceId))
	require.Equal(t, 10, events[0].Delta)
	require.True(t, events[0].CreatedAt.Equal(newerCreatedAt))
	require.Nil(t, events[0].Comment)

	require.Equal(t, olderEventID, uuid.UUID(events[1].Id))
	require.Equal(t, "offerreport", events[1].SourceType)
	require.Equal(t, olderSourceID, uuid.UUID(events[1].SourceId))
	require.Equal(t, -5, events[1].Delta)
	require.True(t, events[1].CreatedAt.Equal(olderCreatedAt))
	require.NotNil(t, events[1].Comment)
	require.Equal(t, olderComment, *events[1].Comment)
}

func TestUsersGetCurrentUserReputationEventsUnauthorized(t *testing.T) {
	t.Parallel()

	fixture := globalFixture

	req, err := http.NewRequest(http.MethodGet, fixture.UsersURL+"/users/reputation-events", nil)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = resp.Body.Close()
	})

	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestUsersUpdateMeAndGetUser(t *testing.T) {
	t.Parallel()

	fixture := globalFixture

	registered := registerAuthUser(t, fixture)
	waitForUsersProjection(t, fixture, registered.UserID)

	accessToken := mustAccessToken(t, registered.UserID)
	updateBody := []byte(`{"name":"Nick","bio":"barter enthusiast","avatarUrl":"http://localhost:8333/avatars/user-1/avatar.jpg"}`)

	req, err := http.NewRequest(http.MethodPatch, fixture.UsersURL+"/users/me", bytes.NewReader(updateBody))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = resp.Body.Close()
	})

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var me usertypes.Me
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&me))
	require.NotNil(t, me.Name)
	require.Equal(t, "Nick", string(*me.Name))
	require.NotNil(t, me.Bio)
	require.Equal(t, "barter enthusiast", string(*me.Bio))
	require.NotNil(t, me.AvatarUrl)
	require.Equal(t, "http://localhost:8333/avatars/user-1/avatar.jpg", string(*me.AvatarUrl))

	getReq, err := http.NewRequest(http.MethodGet, fixture.UsersURL+"/users/"+registered.UserID.String(), nil)
	require.NoError(t, err)
	getReq.Header.Set("Authorization", "Bearer "+accessToken)

	getResp, err := http.DefaultClient.Do(getReq)
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = getResp.Body.Close()
	})

	require.Equal(t, http.StatusOK, getResp.StatusCode)

	var user usertypes.User
	require.NoError(t, json.NewDecoder(getResp.Body).Decode(&user))
	require.Equal(t, registered.UserID, uuid.UUID(user.Id))
	require.NotNil(t, user.Name)
	require.Equal(t, "Nick", string(*user.Name))
	require.NotNil(t, user.Bio)
	require.Equal(t, "barter enthusiast", string(*user.Bio))
	require.NotNil(t, user.AvatarUrl)
	require.Equal(t, "http://localhost:8333/avatars/user-1/avatar.jpg", string(*user.AvatarUrl))
}

func TestUsersUpdateAvatarOnlyAndClear(t *testing.T) {
	t.Parallel()

	fixture := globalFixture

	registered := registerAuthUser(t, fixture)
	waitForUsersProjection(t, fixture, registered.UserID)

	accessToken := mustAccessToken(t, registered.UserID)

	setReq, err := http.NewRequest(http.MethodPatch, fixture.UsersURL+"/users/me", bytes.NewReader([]byte(`{"avatarUrl":"http://localhost:8333/avatars/user-2/avatar.png"}`)))
	require.NoError(t, err)
	setReq.Header.Set("Authorization", "Bearer "+accessToken)
	setReq.Header.Set("Content-Type", "application/json")

	setResp, err := http.DefaultClient.Do(setReq)
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = setResp.Body.Close()
	})

	require.Equal(t, http.StatusOK, setResp.StatusCode)

	var afterSet usertypes.Me
	require.NoError(t, json.NewDecoder(setResp.Body).Decode(&afterSet))
	require.NotNil(t, afterSet.AvatarUrl)
	require.Equal(t, "http://localhost:8333/avatars/user-2/avatar.png", string(*afterSet.AvatarUrl))

	clearReq, err := http.NewRequest(http.MethodPatch, fixture.UsersURL+"/users/me", bytes.NewReader([]byte(`{"avatarUrl":""}`)))
	require.NoError(t, err)
	clearReq.Header.Set("Authorization", "Bearer "+accessToken)
	clearReq.Header.Set("Content-Type", "application/json")

	clearResp, err := http.DefaultClient.Do(clearReq)
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = clearResp.Body.Close()
	})

	require.Equal(t, http.StatusOK, clearResp.StatusCode)

	var afterClear usertypes.Me
	require.NoError(t, json.NewDecoder(clearResp.Body).Decode(&afterClear))
	require.Nil(t, afterClear.AvatarUrl)
}

func TestUsersUpdateMeRejectsEmptyPayload(t *testing.T) {
	t.Parallel()

	fixture := globalFixture

	registered := registerAuthUser(t, fixture)
	waitForUsersProjection(t, fixture, registered.UserID)

	clearReq, err := http.NewRequest(http.MethodPatch, fixture.UsersURL+"/users/me", bytes.NewReader([]byte(`{"bio":null}`)))
	require.NoError(t, err)
	clearReq.Header.Set("Authorization", "Bearer "+mustAccessToken(t, registered.UserID))
	clearReq.Header.Set("Content-Type", "application/json")

	clearResp, err := http.DefaultClient.Do(clearReq)
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = clearResp.Body.Close()
	})

	require.Equal(t, http.StatusBadRequest, clearResp.StatusCode)

	var apiErr usertypes.ErrorResponse
	require.NoError(t, json.NewDecoder(clearResp.Body).Decode(&apiErr))
	require.NotNil(t, apiErr.Message)
	require.Equal(t, "empty update payload", *apiErr.Message)
}

func TestUsersGetUserNotFound(t *testing.T) {
	t.Parallel()

	fixture := globalFixture

	registered := registerAuthUser(t, fixture)
	waitForUsersProjection(t, fixture, registered.UserID)

	req, err := http.NewRequest(http.MethodGet, fixture.UsersURL+"/users/"+uuid.NewString(), nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+mustAccessToken(t, registered.UserID))

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = resp.Body.Close()
	})

	require.Equal(t, http.StatusNotFound, resp.StatusCode)

	var apiErr usertypes.ErrorResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&apiErr))
}

// ────────────────────────────────────────────────────────────────
// Тесты подписок
// ────────────────────────────────────────────────────────────────

func TestSubscribeToUser(t *testing.T) {
	t.Parallel()

	fixture := globalFixture

	subscriber := registerAuthUser(t, fixture)
	target := registerAuthUser(t, fixture)
	waitForUsersProjection(t, fixture, subscriber.UserID)
	waitForUsersProjection(t, fixture, target.UserID)

	body, err := json.Marshal(map[string]string{"targetUserId": target.UserID.String()})
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, fixture.UsersURL+"/users/subscriptions", bytes.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+mustAccessToken(t, subscriber.UserID))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	t.Cleanup(func() { _ = resp.Body.Close() })

	require.Equal(t, http.StatusCreated, resp.StatusCode)
}

func TestSubscribeToUserAlreadySubscribed(t *testing.T) {
	t.Parallel()

	fixture := globalFixture

	subscriber := registerAuthUser(t, fixture)
	target := registerAuthUser(t, fixture)
	waitForUsersProjection(t, fixture, subscriber.UserID)
	waitForUsersProjection(t, fixture, target.UserID)

	body, err := json.Marshal(map[string]string{"targetUserId": target.UserID.String()})
	require.NoError(t, err)

	doSubscribe := func() *http.Response {
		req, err := http.NewRequest(http.MethodPost, fixture.UsersURL+"/users/subscriptions", bytes.NewReader(body))
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+mustAccessToken(t, subscriber.UserID))
		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		return resp
	}

	first := doSubscribe()
	t.Cleanup(func() { _ = first.Body.Close() })
	require.Equal(t, http.StatusCreated, first.StatusCode)

	second := doSubscribe()
	t.Cleanup(func() { _ = second.Body.Close() })
	require.Equal(t, http.StatusConflict, second.StatusCode)
}

func TestSubscribeToUserNotFound(t *testing.T) {
	t.Parallel()

	fixture := globalFixture

	subscriber := registerAuthUser(t, fixture)
	waitForUsersProjection(t, fixture, subscriber.UserID)

	body, err := json.Marshal(map[string]string{"targetUserId": uuid.NewString()})
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, fixture.UsersURL+"/users/subscriptions", bytes.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+mustAccessToken(t, subscriber.UserID))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	t.Cleanup(func() { _ = resp.Body.Close() })

	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestSubscribeToYourself(t *testing.T) {
	t.Parallel()

	fixture := globalFixture

	user := registerAuthUser(t, fixture)
	waitForUsersProjection(t, fixture, user.UserID)

	body, err := json.Marshal(map[string]string{"targetUserId": user.UserID.String()})
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, fixture.UsersURL+"/users/subscriptions", bytes.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+mustAccessToken(t, user.UserID))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	t.Cleanup(func() { _ = resp.Body.Close() })

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestSubscribeUnauthorized(t *testing.T) {
	t.Parallel()

	fixture := globalFixture

	body, err := json.Marshal(map[string]string{"targetUserId": uuid.NewString()})
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, fixture.UsersURL+"/users/subscriptions", bytes.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	t.Cleanup(func() { _ = resp.Body.Close() })

	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestUnsubscribeFromUser(t *testing.T) {
	t.Parallel()

	fixture := globalFixture

	subscriber := registerAuthUser(t, fixture)
	target := registerAuthUser(t, fixture)
	waitForUsersProjection(t, fixture, subscriber.UserID)
	waitForUsersProjection(t, fixture, target.UserID)

	body, err := json.Marshal(map[string]string{"targetUserId": target.UserID.String()})
	require.NoError(t, err)

	// подписываемся
	subReq, err := http.NewRequest(http.MethodPost, fixture.UsersURL+"/users/subscriptions", bytes.NewReader(body))
	require.NoError(t, err)
	subReq.Header.Set("Authorization", "Bearer "+mustAccessToken(t, subscriber.UserID))
	subReq.Header.Set("Content-Type", "application/json")
	subResp, err := http.DefaultClient.Do(subReq)
	require.NoError(t, err)
	_ = subResp.Body.Close()
	require.Equal(t, http.StatusCreated, subResp.StatusCode)

	// отписываемся
	unsubReq, err := http.NewRequest(http.MethodDelete, fixture.UsersURL+"/users/subscriptions", bytes.NewReader(body))
	require.NoError(t, err)
	unsubReq.Header.Set("Authorization", "Bearer "+mustAccessToken(t, subscriber.UserID))
	unsubReq.Header.Set("Content-Type", "application/json")
	unsubResp, err := http.DefaultClient.Do(unsubReq)
	require.NoError(t, err)
	t.Cleanup(func() { _ = unsubResp.Body.Close() })

	require.Equal(t, http.StatusNoContent, unsubResp.StatusCode)
}

func TestUnsubscribeNotSubscribed(t *testing.T) {
	t.Parallel()

	fixture := globalFixture

	subscriber := registerAuthUser(t, fixture)
	target := registerAuthUser(t, fixture)
	waitForUsersProjection(t, fixture, subscriber.UserID)
	waitForUsersProjection(t, fixture, target.UserID)

	body, err := json.Marshal(map[string]string{"targetUserId": target.UserID.String()})
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodDelete, fixture.UsersURL+"/users/subscriptions", bytes.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+mustAccessToken(t, subscriber.UserID))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	t.Cleanup(func() { _ = resp.Body.Close() })

	require.Equal(t, http.StatusConflict, resp.StatusCode)
}

func TestUnsubscribeUnauthorized(t *testing.T) {
	t.Parallel()

	fixture := globalFixture

	body, err := json.Marshal(map[string]string{"targetUserId": uuid.NewString()})
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodDelete, fixture.UsersURL+"/users/subscriptions", bytes.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	t.Cleanup(func() { _ = resp.Body.Close() })

	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestGetSubscriptionsEmpty(t *testing.T) {
	t.Parallel()

	fixture := globalFixture

	user := registerAuthUser(t, fixture)
	waitForUsersProjection(t, fixture, user.UserID)

	req, err := http.NewRequest(http.MethodGet, fixture.UsersURL+"/users/subscriptions", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+mustAccessToken(t, user.UserID))

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	t.Cleanup(func() { _ = resp.Body.Close() })

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var subs usertypes.GetSubscriptionsResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&subs))
	require.Empty(t, subs)
}

func TestGetSubscriptions(t *testing.T) {
	t.Parallel()

	fixture := globalFixture

	subscriber := registerAuthUser(t, fixture)
	target := registerAuthUser(t, fixture)
	waitForUsersProjection(t, fixture, subscriber.UserID)
	waitForUsersProjection(t, fixture, target.UserID)

	// подписываемся
	body, err := json.Marshal(map[string]string{"targetUserId": target.UserID.String()})
	require.NoError(t, err)
	subReq, err := http.NewRequest(http.MethodPost, fixture.UsersURL+"/users/subscriptions", bytes.NewReader(body))
	require.NoError(t, err)
	subReq.Header.Set("Authorization", "Bearer "+mustAccessToken(t, subscriber.UserID))
	subReq.Header.Set("Content-Type", "application/json")
	subResp, err := http.DefaultClient.Do(subReq)
	require.NoError(t, err)
	_ = subResp.Body.Close()
	require.Equal(t, http.StatusCreated, subResp.StatusCode)

	// получаем подписки
	req, err := http.NewRequest(http.MethodGet, fixture.UsersURL+"/users/subscriptions", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+mustAccessToken(t, subscriber.UserID))

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	t.Cleanup(func() { _ = resp.Body.Close() })

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var subs usertypes.GetSubscriptionsResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&subs))
	require.Len(t, subs, 1)
	require.Equal(t, target.UserID, uuid.UUID(subs[0].Id))
}

func TestGetSubscriptionsUnauthorized(t *testing.T) {
	t.Parallel()

	fixture := globalFixture

	req, err := http.NewRequest(http.MethodGet, fixture.UsersURL+"/users/subscriptions", nil)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	t.Cleanup(func() { _ = resp.Body.Close() })

	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestGetSubscribersEmpty(t *testing.T) {
	t.Parallel()

	fixture := globalFixture

	user := registerAuthUser(t, fixture)
	waitForUsersProjection(t, fixture, user.UserID)

	req, err := http.NewRequest(http.MethodGet, fixture.UsersURL+"/users/subscribers", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+mustAccessToken(t, user.UserID))

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	t.Cleanup(func() { _ = resp.Body.Close() })

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var subscribers usertypes.GetSubscriptionsResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&subscribers))
	require.Empty(t, subscribers)
}

func TestGetSubscribers(t *testing.T) {
	t.Parallel()

	fixture := globalFixture

	subscriber := registerAuthUser(t, fixture)
	target := registerAuthUser(t, fixture)
	waitForUsersProjection(t, fixture, subscriber.UserID)
	waitForUsersProjection(t, fixture, target.UserID)

	// subscriber подписывается на target
	body, err := json.Marshal(map[string]string{"targetUserId": target.UserID.String()})
	require.NoError(t, err)
	subReq, err := http.NewRequest(http.MethodPost, fixture.UsersURL+"/users/subscriptions", bytes.NewReader(body))
	require.NoError(t, err)
	subReq.Header.Set("Authorization", "Bearer "+mustAccessToken(t, subscriber.UserID))
	subReq.Header.Set("Content-Type", "application/json")
	subResp, err := http.DefaultClient.Do(subReq)
	require.NoError(t, err)
	_ = subResp.Body.Close()
	require.Equal(t, http.StatusCreated, subResp.StatusCode)

	// target видит subscriber в своих подписчиках
	req, err := http.NewRequest(http.MethodGet, fixture.UsersURL+"/users/subscribers", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+mustAccessToken(t, target.UserID))

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	t.Cleanup(func() { _ = resp.Body.Close() })

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var subscribers usertypes.GetSubscriptionsResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&subscribers))
	require.Len(t, subscribers, 1)
	require.Equal(t, subscriber.UserID, uuid.UUID(subscribers[0].Id))
}

func TestGetSubscribersUnauthorized(t *testing.T) {
	t.Parallel()

	fixture := globalFixture

	req, err := http.NewRequest(http.MethodGet, fixture.UsersURL+"/users/subscribers", nil)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	t.Cleanup(func() { _ = resp.Body.Close() })

	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestSubscribeUnsubscribeAndVerifyLists(t *testing.T) {
	t.Parallel()

	fixture := globalFixture

	subscriber := registerAuthUser(t, fixture)
	target := registerAuthUser(t, fixture)
	waitForUsersProjection(t, fixture, subscriber.UserID)
	waitForUsersProjection(t, fixture, target.UserID)

	body, err := json.Marshal(map[string]string{"targetUserId": target.UserID.String()})
	require.NoError(t, err)

	subscriberToken := mustAccessToken(t, subscriber.UserID)
	targetToken := mustAccessToken(t, target.UserID)

	// подписываемся
	subReq, err := http.NewRequest(http.MethodPost, fixture.UsersURL+"/users/subscriptions", bytes.NewReader(body))
	require.NoError(t, err)
	subReq.Header.Set("Authorization", "Bearer "+subscriberToken)
	subReq.Header.Set("Content-Type", "application/json")
	subResp, err := http.DefaultClient.Do(subReq)
	require.NoError(t, err)
	_ = subResp.Body.Close()
	require.Equal(t, http.StatusCreated, subResp.StatusCode)

	// подписки subscriber содержат target
	getSubsReq, err := http.NewRequest(http.MethodGet, fixture.UsersURL+"/users/subscriptions", nil)
	require.NoError(t, err)
	getSubsReq.Header.Set("Authorization", "Bearer "+subscriberToken)
	getSubsResp, err := http.DefaultClient.Do(getSubsReq)
	require.NoError(t, err)
	var subs usertypes.GetSubscriptionsResponse
	require.NoError(t, json.NewDecoder(getSubsResp.Body).Decode(&subs))
	_ = getSubsResp.Body.Close()
	require.Len(t, subs, 1)
	require.Equal(t, target.UserID, uuid.UUID(subs[0].Id))

	// подписчики target содержат subscriber
	getFollowersReq, err := http.NewRequest(http.MethodGet, fixture.UsersURL+"/users/subscribers", nil)
	require.NoError(t, err)
	getFollowersReq.Header.Set("Authorization", "Bearer "+targetToken)
	getFollowersResp, err := http.DefaultClient.Do(getFollowersReq)
	require.NoError(t, err)
	var followers usertypes.GetSubscriptionsResponse
	require.NoError(t, json.NewDecoder(getFollowersResp.Body).Decode(&followers))
	_ = getFollowersResp.Body.Close()
	require.Len(t, followers, 1)
	require.Equal(t, subscriber.UserID, uuid.UUID(followers[0].Id))

	// отписываемся
	unsubReq, err := http.NewRequest(http.MethodDelete, fixture.UsersURL+"/users/subscriptions", bytes.NewReader(body))
	require.NoError(t, err)
	unsubReq.Header.Set("Authorization", "Bearer "+subscriberToken)
	unsubReq.Header.Set("Content-Type", "application/json")
	unsubResp, err := http.DefaultClient.Do(unsubReq)
	require.NoError(t, err)
	_ = unsubResp.Body.Close()
	require.Equal(t, http.StatusNoContent, unsubResp.StatusCode)

	// подписки subscriber теперь пусты
	getSubsAfterReq, err := http.NewRequest(http.MethodGet, fixture.UsersURL+"/users/subscriptions", nil)
	require.NoError(t, err)
	getSubsAfterReq.Header.Set("Authorization", "Bearer "+subscriberToken)
	getSubsAfterResp, err := http.DefaultClient.Do(getSubsAfterReq)
	require.NoError(t, err)
	var subsAfter usertypes.GetSubscriptionsResponse
	require.NoError(t, json.NewDecoder(getSubsAfterResp.Body).Decode(&subsAfter))
	_ = getSubsAfterResp.Body.Close()
	require.Empty(t, subsAfter)

	// подписчики target тоже пусты
	getFollowersAfterReq, err := http.NewRequest(http.MethodGet, fixture.UsersURL+"/users/subscribers", nil)
	require.NoError(t, err)
	getFollowersAfterReq.Header.Set("Authorization", "Bearer "+targetToken)
	getFollowersAfterResp, err := http.DefaultClient.Do(getFollowersAfterReq)
	require.NoError(t, err)
	var followersAfter usertypes.GetSubscriptionsResponse
	require.NoError(t, json.NewDecoder(getFollowersAfterResp.Body).Decode(&followersAfter))
	_ = getFollowersAfterResp.Body.Close()
	require.Empty(t, followersAfter)
}

// ────────────────────────────────────────────────────────────────
// Вспомогательные функции
// ────────────────────────────────────────────────────────────────

func registerAuthUser(t *testing.T, fixture *Fixture) registerResponse {
	t.Helper()

	payload, err := json.Marshal(registerRequest{
		Email:    fmt.Sprintf("user-%d@example.com", time.Now().UnixNano()),
		Password: "password123",
	})
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, fixture.AuthURL+"/auth/register", bytes.NewReader(payload))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() {
		_ = resp.Body.Close()
	}()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var registered registerResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&registered))
	require.NotEmpty(t, registered.Email)
	require.NotEqual(t, uuid.Nil, registered.UserID)

	return registered
}

func waitForUsersProjection(t *testing.T, fixture *Fixture, userID uuid.UUID) {
	t.Helper()

	pool := OpenDatabase(t, fixture, UsersDBName)

	require.Eventually(t, func() bool {
		var count int
		err := pool.QueryRow(context.Background(), "SELECT count(*) FROM users WHERE id = $1", userID).Scan(&count)
		if err != nil {
			return false
		}
		return count == 1
	}, 20*time.Second, 50*time.Millisecond)
}

func mustAccessToken(t *testing.T, userID uuid.UUID) string {
	t.Helper()

	manager := jwt.NewManager(jwt.Config{
		AccessSecret:  testJWTAccessSecret,
		RefreshSecret: testJWTRefreshSecret,
		AccessTTL:     time.Hour,
		RefreshTTL:    time.Hour,
	})

	token, err := manager.GenerateAccessToken(userID)
	require.NoError(t, err)

	return token
}
