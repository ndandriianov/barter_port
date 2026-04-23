package integration

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	adminEmail    = "admin@barterport.com"
	adminPassword = "admin123"
)

func mustAdminAccessToken(t *testing.T) string {
	t.Helper()

	resp := mustLogin(t, adminEmail, adminPassword)
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var result loginResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	require.NotEmpty(t, result.AccessToken)

	return result.AccessToken
}
