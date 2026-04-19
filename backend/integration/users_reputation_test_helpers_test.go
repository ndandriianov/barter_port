package integration

import (
	usertypes "barter-port/contracts/openapi/users/types"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

const integrationAsyncWaitTimeout = 60 * time.Second

type userReputationEventRecord struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	SourceType string
	SourceID   uuid.UUID
	Delta      int
	Comment    *string
}

func dumpUsersLogs(t *testing.T) {
	t.Helper()
	DumpLogsOnFailure(t, globalFixture.Users, "users")
}

func mustGetCurrentUser(t *testing.T, fixture *Fixture, userID uuid.UUID) usertypes.Me {
	t.Helper()

	req, err := http.NewRequest(http.MethodGet, fixture.UsersURL+"/users/me", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+mustAccessToken(t, userID))

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var me usertypes.Me
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&me))

	return me
}

func mustGetCurrentUserReputationEvents(t *testing.T, fixture *Fixture, userID uuid.UUID) usertypes.GetReputationEventsResponse {
	t.Helper()

	req, err := http.NewRequest(http.MethodGet, fixture.UsersURL+"/users/reputation-events", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+mustAccessToken(t, userID))

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var events usertypes.GetReputationEventsResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&events))

	return events
}

func waitForUserReputationEvent(
	t *testing.T,
	fixture *Fixture,
	userID uuid.UUID,
	sourceType string,
	sourceID uuid.UUID,
) userReputationEventRecord {
	t.Helper()

	pool := OpenDatabase(t, fixture, UsersDBName)
	var event userReputationEventRecord
	deadline := time.Now().Add(integrationAsyncWaitTimeout)

	for time.Now().Before(deadline) {
		err := pool.QueryRow(
			context.Background(),
			`SELECT id, user_id, source_type, source_id, delta, comment
			 FROM user_reputation_events
			 WHERE user_id = $1 AND source_type = $2 AND source_id = $3`,
			userID,
			sourceType,
			sourceID,
		).Scan(
			&event.ID,
			&event.UserID,
			&event.SourceType,
			&event.SourceID,
			&event.Delta,
			&event.Comment,
		)
		if err == nil {
			return event
		}

		time.Sleep(100 * time.Millisecond)
	}

	rows, err := pool.Query(
		context.Background(),
		`SELECT source_type, source_id, user_id, delta, comment
		 FROM user_reputation_inbox
		 ORDER BY created_at DESC, id DESC
		 LIMIT 10`,
	)
	require.NoError(t, err)
	defer rows.Close()

	var inboxRows []string
	for rows.Next() {
		var rowSourceType string
		var rowSourceID uuid.UUID
		var rowUserID uuid.UUID
		var rowDelta int
		var rowComment *string
		require.NoError(t, rows.Scan(&rowSourceType, &rowSourceID, &rowUserID, &rowDelta, &rowComment))
		inboxRows = append(inboxRows, fmt.Sprintf(
			"source_type=%s source_id=%s user_id=%s delta=%d comment=%v",
			rowSourceType,
			rowSourceID,
			rowUserID,
			rowDelta,
			rowComment,
		))
	}

	t.Logf("reputation event not found in users.user_reputation_events; inbox snapshot: %v", inboxRows)
	require.FailNowf(
		t,
		"reputation event not found",
		"user_id=%s source_type=%s source_id=%s",
		userID,
		sourceType,
		sourceID,
	)

	return event
}

func mustFindReputationEvent(
	t *testing.T,
	events usertypes.GetReputationEventsResponse,
	sourceType string,
	sourceID uuid.UUID,
) usertypes.ReputationEvent {
	t.Helper()

	for _, event := range events {
		if event.SourceType == sourceType && uuid.UUID(event.SourceId) == sourceID {
			return event
		}
	}

	require.FailNowf(t, "reputation event not found", "source_type=%s source_id=%s", sourceType, sourceID)
	return usertypes.ReputationEvent{}
}
