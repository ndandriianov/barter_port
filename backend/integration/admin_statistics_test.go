package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type authAdminPlatformStatisticsResponse struct {
	Users struct {
		TotalRegistered int `json:"totalRegistered"`
		VerifiedEmails  int `json:"verifiedEmails"`
	} `json:"users"`
}

type authAdminUserStatisticsResponse struct {
	UserID        uuid.UUID `json:"userId"`
	RegisteredAt  time.Time `json:"registeredAt"`
	EmailVerified bool      `json:"emailVerified"`
}

type usersAdminPlatformStatisticsResponse struct {
	Reputation struct {
		Average  float64 `json:"average"`
		Median   float64 `json:"median"`
		TopUsers []struct {
			UserID           uuid.UUID `json:"userId"`
			Name             *string   `json:"name"`
			ReputationPoints int       `json:"reputationPoints"`
		} `json:"topUsers"`
	} `json:"reputation"`
}

type usersAdminUserStatisticsResponse struct {
	Reputation struct {
		CurrentPoints int `json:"currentPoints"`
		History       []struct {
			ID        uuid.UUID `json:"id"`
			SourceID  uuid.UUID `json:"sourceId"`
			Delta     int       `json:"delta"`
			Comment   *string   `json:"comment"`
			CreatedAt time.Time `json:"createdAt"`
		} `json:"history"`
	} `json:"reputation"`
	Social struct {
		FollowersCount     int `json:"followersCount"`
		SubscriptionsCount int `json:"subscriptionsCount"`
	} `json:"social"`
}

type usersAdminUsersListItem struct {
	ID               uuid.UUID `json:"id"`
	Name             *string   `json:"name"`
	Bio              *string   `json:"bio"`
	AvatarURL        *string   `json:"avatarUrl"`
	PhoneNumber      *string   `json:"phoneNumber"`
	ReputationPoints int       `json:"reputationPoints"`
}

type chatsAdminPlatformStatisticsResponse struct {
	Chats struct {
		Total int `json:"total"`
	} `json:"chats"`
}

type dealsAdminPlatformStatisticsResponse struct {
	Offers struct {
		Total      int   `json:"total"`
		Drafts     int   `json:"drafts"`
		TotalViews int64 `json:"totalViews"`
		Hidden     struct {
			Moderated      int `json:"moderated"`
			HiddenByAuthor int `json:"hiddenByAuthor"`
		} `json:"hidden"`
		ByType struct {
			Good    int `json:"good"`
			Service int `json:"service"`
		} `json:"byType"`
		ByAction struct {
			Give int `json:"give"`
			Take int `json:"take"`
		} `json:"byAction"`
		TopByFavorites []struct {
			OfferID        uuid.UUID `json:"offerId"`
			FavoritesCount int       `json:"favoritesCount"`
		} `json:"topByFavorites"`
	} `json:"offers"`
	Deals struct {
		Total    int `json:"total"`
		ByStatus struct {
			LookingForParticipants int `json:"lookingForParticipants"`
			Completed              int `json:"completed"`
			Failed                 int `json:"failed"`
			Cancelled              int `json:"cancelled"`
		} `json:"byStatus"`
	} `json:"deals"`
	Reports struct {
		Total                     int `json:"total"`
		Pending                   int `json:"pending"`
		BlockedOffers             int `json:"blockedOffers"`
		AdminFailureResolutions   int `json:"adminFailureResolutions"`
		TopUsersByReceivedReports []struct {
			UserID       uuid.UUID `json:"userId"`
			ReportsCount int       `json:"reportsCount"`
		} `json:"topUsersByReceivedReports"`
	} `json:"reports"`
	Reviews struct {
		Total              int `json:"total"`
		RatingDistribution struct {
			ThreeStars int `json:"threeStars"`
			FiveStars  int `json:"fiveStars"`
		} `json:"ratingDistribution"`
	} `json:"reviews"`
}

type dealsAdminUserStatisticsResponse struct {
	Deals struct {
		Completed int `json:"completed"`
		Active    int `json:"active"`
		Failed    struct {
			Total       int `json:"total"`
			Responsible int `json:"responsible"`
			Affected    int `json:"affected"`
		} `json:"failed"`
		Cancelled int `json:"cancelled"`
	} `json:"deals"`
	Offers struct {
		Published  int   `json:"published"`
		TotalViews int64 `json:"totalViews"`
	} `json:"offers"`
	Reviews struct {
		Received              int      `json:"received"`
		AverageReceivedRating *float64 `json:"averageReceivedRating"`
		Written               int      `json:"written"`
	} `json:"reviews"`
	Reports struct {
		Received struct {
			Accepted int `json:"accepted"`
			Rejected int `json:"rejected"`
		} `json:"received"`
		Filed int `json:"filed"`
	} `json:"reports"`
}

func TestAdminStatisticsEndpoints(t *testing.T) {
	fixture := globalFixture
	DumpLogsOnFailure(t, fixture.Auth, "auth")
	DumpLogsOnFailure(t, fixture.Users, "users")
	DumpLogsOnFailure(t, fixture.Chats, "chats")
	DumpLogsOnFailure(t, fixture.Items, "deals")

	adminToken := mustAdminAccessToken(t)
	chatsBefore := mustGetChatsAdminPlatformStatistics(t, fixture, adminToken)
	dealsBefore := mustGetDealsAdminPlatformStatistics(t, adminToken)

	userOne := registerAuthUser(t, fixture)
	userTwo := registerAuthUser(t, fixture)
	waitForUsersProjection(t, fixture, userOne.UserID)
	waitForUsersProjection(t, fixture, userTwo.UserID)

	authDB := OpenDatabase(t, fixture, AuthDBName)
	usersDB := OpenDatabase(t, fixture, UsersDBName)
	dealsDB := OpenDatabase(t, fixture, "deals_db")
	chatsDB := OpenDatabase(t, fixture, "chats_db")

	var userOneRegisteredAt time.Time
	err := authDB.QueryRow(context.Background(), `SELECT created_at FROM users WHERE id = $1`, userOne.UserID).Scan(&userOneRegisteredAt)
	require.NoError(t, err)

	_, err = authDB.Exec(context.Background(), `UPDATE users SET email_verified = true WHERE id = $1`, userOne.UserID)
	require.NoError(t, err)

	var expectedTotalRegistered int
	var expectedVerifiedEmails int
	err = authDB.QueryRow(
		context.Background(),
		`SELECT COUNT(*), COUNT(*) FILTER (WHERE email_verified) FROM users`,
	).Scan(&expectedTotalRegistered, &expectedVerifiedEmails)
	require.NoError(t, err)

	userOneName := "Admin Stats User One"
	userTwoName := "Admin Stats User Two"
	_, err = usersDB.Exec(context.Background(), `UPDATE users SET name = $2, reputation_points = $3 WHERE id = $1`, userOne.UserID, userOneName, 1200)
	require.NoError(t, err)
	_, err = usersDB.Exec(context.Background(), `UPDATE users SET name = $2, reputation_points = $3 WHERE id = $1`, userTwo.UserID, userTwoName, 900)
	require.NoError(t, err)

	firstEventID := uuid.New()
	secondEventID := uuid.New()
	firstSourceID := uuid.New()
	secondSourceID := uuid.New()
	firstComment := "admin stats reward"
	secondComment := "admin stats penalty"
	firstCreatedAt := time.Now().UTC().Add(-2 * time.Minute)
	secondCreatedAt := time.Now().UTC().Add(-1 * time.Minute)

	_, err = usersDB.Exec(
		context.Background(),
		`INSERT INTO user_reputation_events (id, user_id, source_type, source_id, delta, created_at, comment)
		 VALUES ($1, $2, 'DealsOfferReportPenalty', $3, 7, $4, $5),
		        ($6, $2, 'DealsOfferReportPenalty', $7, -3, $8, $9)`,
		firstEventID, userOne.UserID, firstSourceID, firstCreatedAt, firstComment,
		secondEventID, secondSourceID, secondCreatedAt, secondComment,
	)
	require.NoError(t, err)

	_, err = usersDB.Exec(
		context.Background(),
		`INSERT INTO subscriptions (target_user_id, subscriber_id) VALUES ($1, $2), ($2, $1)`,
		userOne.UserID, userTwo.UserID,
	)
	require.NoError(t, err)

	chatID := uuid.New()
	_, err = chatsDB.Exec(
		context.Background(),
		`INSERT INTO chats (id, created_at) VALUES ($1, $2)`,
		chatID, time.Now().UTC(),
	)
	require.NoError(t, err)
	_, err = chatsDB.Exec(
		context.Background(),
		`INSERT INTO chat_participants (chat_id, user_id) VALUES ($1, $2), ($1, $3)`,
		chatID, userOne.UserID, userTwo.UserID,
	)
	require.NoError(t, err)

	now := time.Now().UTC()
	offerOneID := uuid.New()
	offerTwoID := uuid.New()
	offerThreeID := uuid.New()
	draftID := uuid.New()
	activeThirdUserID := uuid.New()
	adminID := uuid.New()
	reporterA := uuid.New()
	reporterB := uuid.New()
	reporterC := uuid.New()

	_, err = dealsDB.Exec(
		context.Background(),
		`INSERT INTO offers (id, author_id, name, type, action, description, created_at, views, is_hidden, hidden_at, hidden_reason)
		 VALUES
		   ($1, $2, 'admin-stats-offer-1', 'good', 'give', 'stats offer 1', $4, 11, TRUE, $4, 'Accepted report'),
		   ($5, $2, 'admin-stats-offer-2', 'service', 'take', 'stats offer 2', $4, 9, TRUE, $4, NULL),
		   ($6, $3, 'admin-stats-offer-3', 'good', 'give', 'stats offer 3', $4, 4, FALSE, NULL, NULL)`,
		offerOneID, userOne.UserID, userTwo.UserID, now,
		offerTwoID, offerThreeID,
	)
	require.NoError(t, err)

	_, err = dealsDB.Exec(context.Background(), `INSERT INTO tags (name) VALUES ('admstats') ON CONFLICT DO NOTHING`)
	require.NoError(t, err)
	_, err = dealsDB.Exec(
		context.Background(),
		`INSERT INTO offer_tags (offer_id, tag_name) VALUES ($1, 'admstats'), ($2, 'admstats')`,
		offerOneID, offerTwoID,
	)
	require.NoError(t, err)
	_, err = dealsDB.Exec(
		context.Background(),
		`INSERT INTO draft_deals (id, author_id, name, description, created_at) VALUES ($1, $2, 'draft', 'draft', $3)`,
		draftID, userOne.UserID, now,
	)
	require.NoError(t, err)

	for i := 0; i < 20; i++ {
		_, err = dealsDB.Exec(
			context.Background(),
			`INSERT INTO favorite_offers (user_id, offer_id, created_at) VALUES ($1, $2, $3)`,
			uuid.New(), offerOneID, now,
		)
		require.NoError(t, err)
	}

	dealCompletedID := uuid.New()
	dealActiveID := uuid.New()
	dealFailedResponsibleID := uuid.New()
	dealFailedAffectedID := uuid.New()
	dealCancelledID := uuid.New()
	_, err = dealsDB.Exec(
		context.Background(),
		`INSERT INTO deals (id, name, description, created_at, status) VALUES
		 ($1, 'completed', 'completed', $6, 'Completed'),
		 ($2, 'active', 'active', $6, 'LookingForParticipants'),
		 ($3, 'failed-responsible', 'failed', $6, 'Failed'),
		 ($4, 'failed-affected', 'failed', $6, 'Failed'),
		 ($5, 'cancelled', 'cancelled', $6, 'Cancelled')`,
		dealCompletedID, dealActiveID, dealFailedResponsibleID, dealFailedAffectedID, dealCancelledID, now,
	)
	require.NoError(t, err)

	_, err = dealsDB.Exec(
		context.Background(),
		`INSERT INTO participants (deal_id, user_id) VALUES
		 ($1, $2), ($1, $3),
		 ($4, $2), ($4, $3), ($4, $5),
		 ($6, $2), ($6, $3),
		 ($7, $2), ($7, $3),
		 ($8, $2), ($8, $3)`,
		dealCompletedID, userOne.UserID, userTwo.UserID,
		dealActiveID, activeThirdUserID,
		dealFailedResponsibleID,
		dealFailedAffectedID,
		dealCancelledID,
	)
	require.NoError(t, err)

	_, err = dealsDB.Exec(
		context.Background(),
		`INSERT INTO deal_failures (deal_id, user_id, confirmed_by_admin, admin_comment, punishment_points) VALUES
		 ($1, $2, TRUE, 'confirmed', -5),
		 ($3, $4, FALSE, 'rejected', 0)`,
		dealFailedResponsibleID, userOne.UserID,
		dealFailedAffectedID, userTwo.UserID,
	)
	require.NoError(t, err)

	_, err = dealsDB.Exec(
		context.Background(),
		`INSERT INTO deal_reviews (id, deal_id, offer_id, author_id, provider_id, rating, comment, created_at) VALUES
		 ($1, $2, $3, $4, $5, 5, 'great', $8),
		 ($6, $7, $9, $5, $4, 3, 'ok', $8)`,
		uuid.New(), dealCompletedID, offerOneID, userTwo.UserID, userOne.UserID,
		uuid.New(), dealFailedAffectedID, now, offerThreeID,
	)
	require.NoError(t, err)

	reportAcceptedID := uuid.New()
	reportRejectedID := uuid.New()
	reportPendingID := uuid.New()
	reportFiledByUserID := uuid.New()
	_, err = dealsDB.Exec(
		context.Background(),
		`INSERT INTO offer_reports
		 (id, offer_id, offer_author_id, status, created_at, reviewed_at, reviewed_by, resolution_comment)
		 VALUES
		 ($1, $2, $3, 'Accepted', $10, $10, $4, 'accepted'),
		 ($5, $6, $3, 'Rejected', $10, $10, $4, 'rejected'),
		 ($7, $6, $3, 'Pending',  $10, NULL, NULL, NULL),
		 ($8, $9, $11, 'Pending', $10, NULL, NULL, NULL)`,
		reportAcceptedID, offerOneID, userOne.UserID, adminID,
		reportRejectedID, offerTwoID,
		reportPendingID,
		reportFiledByUserID, offerThreeID, now, userTwo.UserID,
	)
	require.NoError(t, err)

	_, err = dealsDB.Exec(
		context.Background(),
		`INSERT INTO offer_reports_messages (offer_report_id, author_id, message) VALUES
		 ($1, $2, 'accepted'),
		 ($3, $4, 'rejected'),
		 ($5, $6, 'pending'),
		 ($7, $8, 'filed by user one')`,
		reportAcceptedID, reporterA,
		reportRejectedID, reporterB,
		reportPendingID, reporterC,
		reportFiledByUserID, userOne.UserID,
	)
	require.NoError(t, err)

	authAfter := mustGetAuthAdminPlatformStatistics(t, fixture, adminToken)
	require.Equal(t, expectedTotalRegistered, authAfter.Users.TotalRegistered)
	require.Equal(t, expectedVerifiedEmails, authAfter.Users.VerifiedEmails)

	authUserStats := mustGetAuthAdminUserStatistics(t, fixture, adminToken, userOne.UserID)
	require.Equal(t, userOne.UserID, authUserStats.UserID)
	require.WithinDuration(t, userOneRegisteredAt.UTC(), authUserStats.RegisteredAt.UTC(), time.Second)
	require.True(t, authUserStats.EmailVerified)

	usersPlatform := mustGetUsersAdminPlatformStatistics(t, fixture, adminToken)
	require.NotEmpty(t, usersPlatform.Reputation.TopUsers)
	require.Equal(t, userOne.UserID, usersPlatform.Reputation.TopUsers[0].UserID)
	require.Equal(t, 1200, usersPlatform.Reputation.TopUsers[0].ReputationPoints)

	usersUserStats := mustGetUsersAdminUserStatistics(t, fixture, adminToken, userOne.UserID)
	require.Equal(t, 1200, usersUserStats.Reputation.CurrentPoints)
	require.Len(t, usersUserStats.Reputation.History, 2)
	require.Equal(t, 1, usersUserStats.Social.FollowersCount)
	require.Equal(t, 1, usersUserStats.Social.SubscriptionsCount)

	usersList := mustGetUsersAdminUsersList(t, fixture, adminToken)
	userOneFromList := findUsersAdminListItem(t, usersList, userOne.UserID)
	require.NotNil(t, userOneFromList.Name)
	require.Equal(t, userOneName, *userOneFromList.Name)
	require.Equal(t, 1200, userOneFromList.ReputationPoints)

	userTwoFromList := findUsersAdminListItem(t, usersList, userTwo.UserID)
	require.NotNil(t, userTwoFromList.Name)
	require.Equal(t, userTwoName, *userTwoFromList.Name)
	require.Equal(t, 900, userTwoFromList.ReputationPoints)

	chatsAfter := mustGetChatsAdminPlatformStatistics(t, fixture, adminToken)
	require.Equal(t, chatsBefore.Chats.Total+1, chatsAfter.Chats.Total)

	dealsAfter := mustGetDealsAdminPlatformStatistics(t, adminToken)
	require.Equal(t, dealsBefore.Offers.Total+3, dealsAfter.Offers.Total)
	require.Equal(t, dealsBefore.Offers.Drafts+1, dealsAfter.Offers.Drafts)
	require.Equal(t, dealsBefore.Offers.TotalViews+24, dealsAfter.Offers.TotalViews)
	require.Equal(t, dealsBefore.Offers.Hidden.Moderated+1, dealsAfter.Offers.Hidden.Moderated)
	require.Equal(t, dealsBefore.Offers.Hidden.HiddenByAuthor+1, dealsAfter.Offers.Hidden.HiddenByAuthor)
	require.Equal(t, dealsBefore.Offers.ByType.Good+2, dealsAfter.Offers.ByType.Good)
	require.Equal(t, dealsBefore.Offers.ByType.Service+1, dealsAfter.Offers.ByType.Service)
	require.Equal(t, dealsBefore.Offers.ByAction.Give+2, dealsAfter.Offers.ByAction.Give)
	require.Equal(t, dealsBefore.Offers.ByAction.Take+1, dealsAfter.Offers.ByAction.Take)
	require.NotEmpty(t, dealsAfter.Offers.TopByFavorites)
	require.Equal(t, offerOneID, dealsAfter.Offers.TopByFavorites[0].OfferID)
	require.Equal(t, 20, dealsAfter.Offers.TopByFavorites[0].FavoritesCount)
	require.Equal(t, dealsBefore.Deals.Total+5, dealsAfter.Deals.Total)
	require.Equal(t, dealsBefore.Deals.ByStatus.LookingForParticipants+1, dealsAfter.Deals.ByStatus.LookingForParticipants)
	require.Equal(t, dealsBefore.Deals.ByStatus.Completed+1, dealsAfter.Deals.ByStatus.Completed)
	require.Equal(t, dealsBefore.Deals.ByStatus.Failed+2, dealsAfter.Deals.ByStatus.Failed)
	require.Equal(t, dealsBefore.Deals.ByStatus.Cancelled+1, dealsAfter.Deals.ByStatus.Cancelled)
	require.Equal(t, dealsBefore.Reports.Total+4, dealsAfter.Reports.Total)
	require.Equal(t, dealsBefore.Reports.Pending+2, dealsAfter.Reports.Pending)
	require.Equal(t, dealsBefore.Reports.BlockedOffers+1, dealsAfter.Reports.BlockedOffers)
	require.Equal(t, dealsBefore.Reports.AdminFailureResolutions+2, dealsAfter.Reports.AdminFailureResolutions)
	require.NotEmpty(t, dealsAfter.Reports.TopUsersByReceivedReports)
	require.Equal(t, userOne.UserID, dealsAfter.Reports.TopUsersByReceivedReports[0].UserID)
	require.Equal(t, 3, dealsAfter.Reports.TopUsersByReceivedReports[0].ReportsCount)
	require.Equal(t, dealsBefore.Reviews.Total+2, dealsAfter.Reviews.Total)
	require.Equal(t, dealsBefore.Reviews.RatingDistribution.FiveStars+1, dealsAfter.Reviews.RatingDistribution.FiveStars)
	require.Equal(t, dealsBefore.Reviews.RatingDistribution.ThreeStars+1, dealsAfter.Reviews.RatingDistribution.ThreeStars)

	dealsUserStats := mustGetDealsAdminUserStatistics(t, adminToken, userOne.UserID)
	require.Equal(t, 1, dealsUserStats.Deals.Completed)
	require.Equal(t, 1, dealsUserStats.Deals.Active)
	require.Equal(t, 2, dealsUserStats.Deals.Failed.Total)
	require.Equal(t, 1, dealsUserStats.Deals.Failed.Responsible)
	require.Equal(t, 1, dealsUserStats.Deals.Failed.Affected)
	require.Equal(t, 1, dealsUserStats.Deals.Cancelled)
	require.Equal(t, 2, dealsUserStats.Offers.Published)
	require.Equal(t, int64(20), dealsUserStats.Offers.TotalViews)
	require.Equal(t, 1, dealsUserStats.Reviews.Received)
	require.NotNil(t, dealsUserStats.Reviews.AverageReceivedRating)
	require.InDelta(t, 5, *dealsUserStats.Reviews.AverageReceivedRating, 0.001)
	require.Equal(t, 1, dealsUserStats.Reviews.Written)
	require.Equal(t, 1, dealsUserStats.Reports.Received.Accepted)
	require.Equal(t, 1, dealsUserStats.Reports.Received.Rejected)
	require.Equal(t, 1, dealsUserStats.Reports.Filed)
}

func mustGetAuthAdminPlatformStatistics(t *testing.T, fixture *Fixture, adminToken string) authAdminPlatformStatisticsResponse {
	t.Helper()

	req := mustBearerRequest(t, http.MethodGet, fixture.AuthURL+"/auth/admin/statistics/platform", adminToken, nil)
	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var body authAdminPlatformStatisticsResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	return body
}

func mustGetAuthAdminUserStatistics(t *testing.T, fixture *Fixture, adminToken string, userID uuid.UUID) authAdminUserStatisticsResponse {
	t.Helper()

	req := mustBearerRequest(t, http.MethodGet, fixture.AuthURL+"/auth/admin/users/"+userID.String()+"/statistics", adminToken, nil)
	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var body authAdminUserStatisticsResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	return body
}

func mustGetUsersAdminPlatformStatistics(t *testing.T, fixture *Fixture, adminToken string) usersAdminPlatformStatisticsResponse {
	t.Helper()

	req := mustBearerRequest(t, http.MethodGet, fixture.UsersURL+"/admin/statistics/platform", adminToken, nil)
	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var body usersAdminPlatformStatisticsResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	return body
}

func mustGetUsersAdminUserStatistics(t *testing.T, fixture *Fixture, adminToken string, userID uuid.UUID) usersAdminUserStatisticsResponse {
	t.Helper()

	req := mustBearerRequest(t, http.MethodGet, fixture.UsersURL+"/admin/users/"+userID.String()+"/statistics", adminToken, nil)
	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var body usersAdminUserStatisticsResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	return body
}

func mustGetUsersAdminUsersList(t *testing.T, fixture *Fixture, adminToken string) []usersAdminUsersListItem {
	t.Helper()

	req := mustBearerRequest(t, http.MethodGet, fixture.UsersURL+"/admin/users", adminToken, nil)
	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var body []usersAdminUsersListItem
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	return body
}

func findUsersAdminListItem(t *testing.T, items []usersAdminUsersListItem, userID uuid.UUID) usersAdminUsersListItem {
	t.Helper()

	for _, item := range items {
		if item.ID == userID {
			return item
		}
	}

	t.Fatalf("user %s not found in admin users list", userID)
	return usersAdminUsersListItem{}
}

func mustGetChatsAdminPlatformStatistics(t *testing.T, fixture *Fixture, adminToken string) chatsAdminPlatformStatisticsResponse {
	t.Helper()

	req := mustBearerRequest(t, http.MethodGet, fixture.ChatsURL+"/admin/statistics/platform", adminToken, nil)
	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var body chatsAdminPlatformStatisticsResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	return body
}

func mustGetDealsAdminPlatformStatistics(t *testing.T, adminToken string) dealsAdminPlatformStatisticsResponse {
	t.Helper()

	req := mustBearerRequest(t, http.MethodGet, dealsURL()+"/admin/statistics/platform", adminToken, nil)
	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var body dealsAdminPlatformStatisticsResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	return body
}

func mustGetDealsAdminUserStatistics(t *testing.T, adminToken string, userID uuid.UUID) dealsAdminUserStatisticsResponse {
	t.Helper()

	req := mustBearerRequest(t, http.MethodGet, dealsURL()+"/admin/users/"+userID.String()+"/statistics", adminToken, nil)
	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var body dealsAdminUserStatisticsResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	return body
}
