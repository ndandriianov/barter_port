package integration

import (
	"barter-port/pkg/httpx"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// ────────────────────────────────────────────────────────────────
// DTO
// ────────────────────────────────────────────────────────────────

type loginResponse struct {
	AccessToken string `json:"accessToken"`
}

type refreshTokenResponse struct {
	AccessToken string `json:"access_token"`
}

type changePasswordRequest struct {
	OldEmail    string `json:"oldEmail"`
	OldPassword string `json:"oldPassword"`
	NewPassword string `json:"newPassword"`
}

type authMeResponse struct {
	UserID uuid.UUID `json:"userId"`
}

type userCreationStatusResponse struct {
	Status string `json:"status"`
}

// ────────────────────────────────────────────────────────────────
// Тесты
// ────────────────────────────────────────────────────────────────

// dumpAuthLogs регистрирует вывод логов auth-контейнера при падении теста.
func dumpAuthLogs(t *testing.T) {
	t.Helper()
	DumpLogsOnFailure(t, globalFixture.Auth, "auth")
}

func TestAuthRegisterSuccess(t *testing.T) {
	t.Parallel()
	dumpAuthLogs(t)

	registered := registerAuthUser(t, globalFixture)

	require.NotEqual(t, uuid.Nil, registered.UserID)
	require.NotEmpty(t, registered.Email)
}

func TestAuthRegisterDuplicateEmail(t *testing.T) {
	t.Parallel()
	dumpAuthLogs(t)

	email := fmt.Sprintf("dup-%d@example.com", time.Now().UnixNano())
	payload, err := json.Marshal(registerRequest{Email: email, Password: "password123"})
	require.NoError(t, err)

	resp1, err := http.Post(globalFixture.AuthURL+"/auth/register", "application/json", bytes.NewReader(payload))
	require.NoError(t, err)
	_ = resp1.Body.Close()
	require.Equal(t, http.StatusOK, resp1.StatusCode)

	resp2, err := http.Post(globalFixture.AuthURL+"/auth/register", "application/json", bytes.NewReader(payload))
	require.NoError(t, err)
	defer func() { _ = resp2.Body.Close() }()

	require.Equal(t, http.StatusBadRequest, resp2.StatusCode)

	var errResp httpx.ErrorResponse
	require.NoError(t, json.NewDecoder(resp2.Body).Decode(&errResp))
	require.Equal(t, "email already in use", *errResp.Message)
}

func TestAuthRegisterInvalidEmail(t *testing.T) {
	t.Parallel()
	dumpAuthLogs(t)

	payload, err := json.Marshal(registerRequest{Email: "not-an-email", Password: "password123"})
	require.NoError(t, err)

	resp, err := http.Post(globalFixture.AuthURL+"/auth/register", "application/json", bytes.NewReader(payload))
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var errResp httpx.ErrorResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&errResp))
	require.Equal(t, "invalid email", *errResp.Message)
}

func TestAuthRegisterShortPassword(t *testing.T) {
	t.Parallel()
	dumpAuthLogs(t)

	payload, err := json.Marshal(registerRequest{
		Email:    fmt.Sprintf("pw-%d@example.com", time.Now().UnixNano()),
		Password: "abc",
	})
	require.NoError(t, err)

	resp, err := http.Post(globalFixture.AuthURL+"/auth/register", "application/json", bytes.NewReader(payload))
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var errResp httpx.ErrorResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&errResp))
	require.Equal(t, "password too short", *errResp.Message)
}

func TestAuthLoginSuccess(t *testing.T) {
	t.Parallel()
	dumpAuthLogs(t)

	email, password := uniqueEmail("login"), "password123"
	mustRegister(t, email, password)

	resp := mustLogin(t, email, password)
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var body loginResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	require.NotEmpty(t, body.AccessToken)

	var refreshCookie *http.Cookie
	for _, c := range resp.Cookies() {
		if c.Name == "refresh_token" {
			refreshCookie = c
			break
		}
	}
	require.NotNil(t, refreshCookie, "refresh_token cookie must be set")
	require.NotEmpty(t, refreshCookie.Value)
	require.True(t, refreshCookie.HttpOnly)
}

func TestAuthLoginInvalidCredentials(t *testing.T) {
	t.Parallel()
	dumpAuthLogs(t)

	resp := mustLogin(t, "nobody@example.com", "password123")
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var errResp httpx.ErrorResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&errResp))
	require.Equal(t, "invalid credentials", *errResp.Message)
}

func TestAuthLoginWrongPassword(t *testing.T) {
	t.Parallel()
	dumpAuthLogs(t)

	email := uniqueEmail("wrongpw")
	mustRegister(t, email, "correct-password")

	resp := mustLogin(t, email, "wrong-password")
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	var errResp httpx.ErrorResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&errResp))
	require.Equal(t, "incorrect password", *errResp.Message)
}

func TestAuthMe(t *testing.T) {
	t.Parallel()
	dumpAuthLogs(t)

	registered := registerAuthUser(t, globalFixture)

	req, err := http.NewRequest(http.MethodGet, globalFixture.AuthURL+"/auth/me", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+mustAccessToken(t, registered.UserID))

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var me authMeResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&me))
	require.Equal(t, registered.UserID, me.UserID)
}

func TestAuthMeUnauthorized(t *testing.T) {
	t.Parallel()
	dumpAuthLogs(t)

	resp, err := http.Get(globalFixture.AuthURL + "/auth/me")
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestAuthRefresh(t *testing.T) {
	t.Parallel()
	dumpAuthLogs(t)

	email, password := uniqueEmail("refresh"), "password123"
	mustRegister(t, email, password)

	client := clientWithCookieJar(t)

	// Логин — cookie jar сохраняет refresh_token
	loginPayload, err := json.Marshal(registerRequest{Email: email, Password: password})
	require.NoError(t, err)
	loginReq, err := http.NewRequest(http.MethodPost, globalFixture.AuthURL+"/auth/login", bytes.NewReader(loginPayload))
	require.NoError(t, err)
	loginReq.Header.Set("Content-Type", "application/json")
	loginResp, err := client.Do(loginReq)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, loginResp.StatusCode)
	var firstLogin loginResponse
	require.NoError(t, json.NewDecoder(loginResp.Body).Decode(&firstLogin))
	_ = loginResp.Body.Close()

	// Refresh — cookie автоматически передаётся из jar
	refreshReq, err := http.NewRequest(http.MethodPost, globalFixture.AuthURL+"/auth/refresh", nil)
	require.NoError(t, err)
	refreshResp, err := client.Do(refreshReq)
	require.NoError(t, err)
	defer func() { _ = refreshResp.Body.Close() }()

	require.Equal(t, http.StatusOK, refreshResp.StatusCode)

	var result refreshTokenResponse
	require.NoError(t, json.NewDecoder(refreshResp.Body).Decode(&result))
	require.NotEmpty(t, result.AccessToken)
	require.NotEqual(t, firstLogin.AccessToken, result.AccessToken)
}

func TestAuthRefreshNoCookie(t *testing.T) {
	t.Parallel()
	dumpAuthLogs(t)

	resp, err := http.Post(globalFixture.AuthURL+"/auth/refresh", "application/json", nil)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestAuthLogout(t *testing.T) {
	t.Parallel()
	dumpAuthLogs(t)

	email, password := uniqueEmail("logout"), "password123"
	mustRegister(t, email, password)

	client := clientWithCookieJar(t)

	// Логин
	loginPayload, err := json.Marshal(registerRequest{Email: email, Password: password})
	require.NoError(t, err)
	loginReq, err := http.NewRequest(http.MethodPost, globalFixture.AuthURL+"/auth/login", bytes.NewReader(loginPayload))
	require.NoError(t, err)
	loginReq.Header.Set("Content-Type", "application/json")
	loginResp, err := client.Do(loginReq)
	require.NoError(t, err)
	_ = loginResp.Body.Close()
	require.Equal(t, http.StatusOK, loginResp.StatusCode)

	// Logout
	logoutReq, err := http.NewRequest(http.MethodPost, globalFixture.AuthURL+"/auth/logout", nil)
	require.NoError(t, err)
	logoutResp, err := client.Do(logoutReq)
	require.NoError(t, err)
	defer func() { _ = logoutResp.Body.Close() }()
	require.Equal(t, http.StatusOK, logoutResp.StatusCode)

	// После logout refresh должен вернуть 401 (cookie очищен или токен отозван)
	refreshReq, err := http.NewRequest(http.MethodPost, globalFixture.AuthURL+"/auth/refresh", nil)
	require.NoError(t, err)
	refreshResp, err := client.Do(refreshReq)
	require.NoError(t, err)
	defer func() { _ = refreshResp.Body.Close() }()
	require.Equal(t, http.StatusUnauthorized, refreshResp.StatusCode)
}

func TestAuthLogoutWithoutCookie(t *testing.T) {
	t.Parallel()
	dumpAuthLogs(t)

	resp, err := http.Post(globalFixture.AuthURL+"/auth/logout", "application/json", nil)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	// Logout идемпотентен — работает даже без cookie
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestAuthChangePasswordSuccess(t *testing.T) {
	t.Parallel()
	dumpAuthLogs(t)

	email := uniqueEmail("change-password")
	oldPassword := "password123"
	newPassword := "password456"
	mustRegister(t, email, oldPassword)

	accessToken := mustLoginAccessToken(t, email, oldPassword)

	changePassword(t, accessToken, changePasswordRequest{
		OldEmail:    email,
		OldPassword: oldPassword,
		NewPassword: newPassword,
	}, http.StatusOK)

	oldLoginResp := mustLogin(t, email, oldPassword)
	defer func() { _ = oldLoginResp.Body.Close() }()
	require.Equal(t, http.StatusUnauthorized, oldLoginResp.StatusCode)

	newLoginResp := mustLogin(t, email, newPassword)
	defer func() { _ = newLoginResp.Body.Close() }()
	require.Equal(t, http.StatusOK, newLoginResp.StatusCode)
}

func TestAuthChangePasswordMissingOldCredentials(t *testing.T) {
	t.Parallel()
	dumpAuthLogs(t)

	email := uniqueEmail("change-password-missing-old")
	oldPassword := "password123"
	mustRegister(t, email, oldPassword)

	accessToken := mustLoginAccessToken(t, email, oldPassword)

	resp := changePassword(t, accessToken, changePasswordRequest{
		NewPassword: "password456",
	}, http.StatusForbidden)
	defer func() { _ = resp.Body.Close() }()

	var errResp httpx.ErrorResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&errResp))
	require.Equal(t, "old credentials are invalid", *errResp.Message)
}

func TestAuthChangePasswordWrongOldPassword(t *testing.T) {
	t.Parallel()
	dumpAuthLogs(t)

	email := uniqueEmail("change-password-wrong-old")
	oldPassword := "password123"
	mustRegister(t, email, oldPassword)

	accessToken := mustLoginAccessToken(t, email, oldPassword)

	resp := changePassword(t, accessToken, changePasswordRequest{
		OldEmail:    email,
		OldPassword: "wrong-password",
		NewPassword: "password456",
	}, http.StatusForbidden)
	defer func() { _ = resp.Body.Close() }()

	var errResp httpx.ErrorResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&errResp))
	require.Equal(t, "old credentials are invalid", *errResp.Message)
}

func TestAuthChangePasswordShortNewPassword(t *testing.T) {
	t.Parallel()
	dumpAuthLogs(t)

	email := uniqueEmail("change-password-short-new")
	oldPassword := "password123"
	mustRegister(t, email, oldPassword)

	accessToken := mustLoginAccessToken(t, email, oldPassword)

	resp := changePassword(t, accessToken, changePasswordRequest{
		OldEmail:    email,
		OldPassword: oldPassword,
		NewPassword: "123",
	}, http.StatusBadRequest)
	defer func() { _ = resp.Body.Close() }()

	var errResp httpx.ErrorResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&errResp))
	require.Equal(t, "password too short", *errResp.Message)

	loginResp := mustLogin(t, email, oldPassword)
	defer func() { _ = loginResp.Body.Close() }()
	require.Equal(t, http.StatusOK, loginResp.StatusCode)
}

func TestAuthGetUserCreationStatus(t *testing.T) {
	t.Parallel()
	dumpAuthLogs(t)

	registered := registerAuthUser(t, globalFixture)

	resp, err := http.Get(globalFixture.AuthURL + "/auth/status/" + registered.UserID.String())
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var statusResp userCreationStatusResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&statusResp))
	require.NotEmpty(t, statusResp.Status)
}

func TestAuthGetUserCreationStatusInvalidID(t *testing.T) {
	t.Parallel()
	dumpAuthLogs(t)

	resp, err := http.Get(globalFixture.AuthURL + "/auth/status/not-a-uuid")
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestAuthGetUserCreationStatusNotFound(t *testing.T) {
	t.Parallel()
	dumpAuthLogs(t)

	resp, err := http.Get(globalFixture.AuthURL + "/auth/status/" + uuid.NewString())
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

// ────────────────────────────────────────────────────────────────
// Вспомогательные функции
// ────────────────────────────────────────────────────────────────

func uniqueEmail(prefix string) string {
	return fmt.Sprintf("%s-%d@example.com", prefix, time.Now().UnixNano())
}

func mustRegister(t *testing.T, email, password string) {
	t.Helper()

	payload, err := json.Marshal(registerRequest{Email: email, Password: password})
	require.NoError(t, err)

	resp, err := http.Post(globalFixture.AuthURL+"/auth/register", "application/json", bytes.NewReader(payload))
	require.NoError(t, err)
	_ = resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func mustLogin(t *testing.T, email, password string) *http.Response {
	t.Helper()

	payload, err := json.Marshal(registerRequest{Email: email, Password: password})
	require.NoError(t, err)

	resp, err := http.Post(globalFixture.AuthURL+"/auth/login", "application/json", bytes.NewReader(payload))
	require.NoError(t, err)

	return resp
}

func mustLoginAccessToken(t *testing.T, email, password string) string {
	t.Helper()

	resp := mustLogin(t, email, password)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var login loginResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&login))
	require.NotEmpty(t, login.AccessToken)

	return login.AccessToken
}

func changePassword(t *testing.T, accessToken string, reqBody changePasswordRequest, expectedStatus int) *http.Response {
	t.Helper()

	payload, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, globalFixture.AuthURL+"/auth/change-password", bytes.NewReader(payload))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, expectedStatus, resp.StatusCode)

	return resp
}

func mustUserIDByCreds(t *testing.T, email, password string) uuid.UUID {
	t.Helper()

	accessToken := mustLoginAccessToken(t, email, password)

	req, err := http.NewRequest(http.MethodGet, globalFixture.AuthURL+"/auth/me", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var me authMeResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&me))
	require.NotEqual(t, uuid.Nil, me.UserID)

	return me.UserID
}

// clientWithCookieJar возвращает HTTP-клиент с автоматической обработкой cookie.
func clientWithCookieJar(t *testing.T) *http.Client {
	t.Helper()

	jar, err := cookiejar.New(nil)
	require.NoError(t, err)

	return &http.Client{Jar: jar}
}
