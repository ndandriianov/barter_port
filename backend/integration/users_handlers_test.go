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
	"github.com/testcontainers/testcontainers-go/network"
)

// globalFixture — единый стек контейнеров, разделяемый всеми тестами пакета.
var globalFixture *Fixture

func TestMain(m *testing.M) {
	os.Exit(runTests(m))
}

func runTests(m *testing.M) int {
	ctx := context.Background()

	net, err := network.New(ctx)
	if err != nil {
		log.Printf("не удалось создать docker-сеть: %v", err)
		return 1
	}

	var setupErr error
	globalFixture, setupErr = newSharedFixture(ctx, net, FixtureOptions{
		NeedAuth:  true,
		NeedUsers: true,
		NeedItems: true,
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
	require.False(t, me.CreatedAt.IsZero())
}

func TestUsersUpdateMeAndGetUser(t *testing.T) {
	t.Parallel()

	fixture := globalFixture

	registered := registerAuthUser(t, fixture)
	waitForUsersProjection(t, fixture, registered.UserID)

	accessToken := mustAccessToken(t, registered.UserID)
	updateBody := []byte(`{"name":"Nick","bio":"barter enthusiast"}`)

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
