package integration

import (
	dealsusers "barter-port/contracts/kafka/messages/deals-users"
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
const dealCompletionRewardPoints = 5
const reviewCreationRewardPoints = 2

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

func mustRegisterProjectedUser(t *testing.T, fixture *Fixture) uuid.UUID {
	t.Helper()

	registered := registerAuthUser(t, fixture)
	waitForUsersProjection(t, fixture, registered.UserID)

	return registered.UserID
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
		if string(event.SourceType) == sourceType && uuid.UUID(event.SourceId) == sourceID {
			return event
		}
	}

	require.FailNowf(t, "reputation event not found", "source_type=%s source_id=%s", sourceType, sourceID)
	return usertypes.ReputationEvent{}
}

func waitForCurrentUserReputationAPIEvent(
	t *testing.T,
	fixture *Fixture,
	userID uuid.UUID,
	sourceType string,
	sourceID uuid.UUID,
) usertypes.ReputationEvent {
	t.Helper()

	deadline := time.Now().Add(integrationAsyncWaitTimeout)
	var lastEvents usertypes.GetReputationEventsResponse

	for time.Now().Before(deadline) {
		lastEvents = mustGetCurrentUserReputationEvents(t, fixture, userID)
		for _, event := range lastEvents {
			if string(event.SourceType) == sourceType && uuid.UUID(event.SourceId) == sourceID {
				return event
			}
		}

		time.Sleep(100 * time.Millisecond)
	}

	t.Logf(
		"reputation event not found via users API; user_id=%s source_type=%s source_id=%s events_count=%d",
		userID,
		sourceType,
		sourceID,
		len(lastEvents),
	)
	require.FailNowf(t, "reputation event not found via users API", "source_type=%s source_id=%s", sourceType, sourceID)
	return usertypes.ReputationEvent{}
}

func waitForCurrentUserReputationPoints(
	t *testing.T,
	fixture *Fixture,
	userID uuid.UUID,
	expected int,
) usertypes.Me {
	t.Helper()

	deadline := time.Now().Add(integrationAsyncWaitTimeout)
	var lastMe usertypes.Me

	for time.Now().Before(deadline) {
		lastMe = mustGetCurrentUser(t, fixture, userID)
		if int(lastMe.ReputationPoints) == expected {
			return lastMe
		}

		time.Sleep(100 * time.Millisecond)
	}

	t.Logf(
		"reputation points did not reach expected value; user_id=%s expected=%d actual=%d",
		userID,
		expected,
		lastMe.ReputationPoints,
	)
	require.FailNowf(t, "unexpected reputation points", "expected=%d actual=%d", expected, lastMe.ReputationPoints)
	return lastMe
}

func countUserReputationEvents(
	t *testing.T,
	fixture *Fixture,
	userID uuid.UUID,
	sourceType string,
	sourceID uuid.UUID,
) int {
	t.Helper()

	pool := OpenDatabase(t, fixture, UsersDBName)
	var count int
	err := pool.QueryRow(
		context.Background(),
		`SELECT COUNT(*)
		 FROM user_reputation_events
		 WHERE user_id = $1 AND source_type = $2 AND source_id = $3`,
		userID,
		sourceType,
		sourceID,
	).Scan(&count)
	require.NoError(t, err)

	return count
}

func requireReviewCreationRewardEvent(
	t *testing.T,
	fixture *Fixture,
	userID uuid.UUID,
	dealID uuid.UUID,
	itemID, offerID *uuid.UUID,
	providerID uuid.UUID,
) userReputationEventRecord {
	t.Helper()

	sourceID := dealsusers.BuildReviewCreationRewardSourceID(dealID, itemID, offerID, userID, providerID)
	event := waitForUserReputationEvent(t, fixture, userID, dealsusers.ReviewCreationRewardMessageType, sourceID)
	require.Equal(t, reviewCreationRewardPoints, event.Delta)
	return event
}
