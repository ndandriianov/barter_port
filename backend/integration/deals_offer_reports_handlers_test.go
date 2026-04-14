package integration

import (
	"barter-port/contracts/openapi/deals/types"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// ─────────────────────────────────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────────────────────────────────

func doCreateOfferReport(t *testing.T, reporterID uuid.UUID, offerID uuid.UUID, message string) *http.Response {
	t.Helper()

	body := types.CreateOfferReportRequest{Message: message}
	req := mustUserRequest(t, http.MethodPost, dealsURL()+"/offers/"+offerID.String()+"/reports", reporterID, mustJSONBody(t, body))
	req.Header.Set("Content-Type", "application/json")
	return mustDo(t, req)
}

func mustCreateOfferReport(t *testing.T, reporterID uuid.UUID, offerID uuid.UUID, message string) types.OfferReport {
	t.Helper()

	resp := doCreateOfferReport(t, reporterID, offerID, message)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var report types.OfferReport
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&report))
	return report
}

func doGetOfferReports(t *testing.T, userID uuid.UUID, offerID uuid.UUID) *http.Response {
	t.Helper()

	req := mustUserRequest(t, http.MethodGet, dealsURL()+"/offers/"+offerID.String()+"/reports", userID, nil)
	return mustDo(t, req)
}

func mustGetOfferReports(t *testing.T, userID uuid.UUID, offerID uuid.UUID) types.OfferReportsForOffer {
	t.Helper()

	resp := doGetOfferReports(t, userID, offerID)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var result types.OfferReportsForOffer
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	return result
}

func doListAdminReports(t *testing.T, adminToken string, status *types.OfferReportStatus) *http.Response {
	t.Helper()

	url := dealsURL() + "/admin/offer-reports"
	if status != nil {
		url += "?status=" + string(*status)
	}
	req := mustBearerRequest(t, http.MethodGet, url, adminToken, nil)
	return mustDo(t, req)
}

func mustListAdminReports(t *testing.T, adminToken string, status *types.OfferReportStatus) types.ListOfferReportsResponse {
	t.Helper()

	resp := doListAdminReports(t, adminToken, status)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var result types.ListOfferReportsResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	return result
}

func doGetAdminReportDetails(t *testing.T, adminToken string, reportID uuid.UUID) *http.Response {
	t.Helper()

	req := mustBearerRequest(t, http.MethodGet, dealsURL()+"/admin/offer-reports/"+reportID.String(), adminToken, nil)
	return mustDo(t, req)
}

func mustGetAdminReportDetails(t *testing.T, adminToken string, reportID uuid.UUID) types.OfferReportDetails {
	t.Helper()

	resp := doGetAdminReportDetails(t, adminToken, reportID)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var result types.OfferReportDetails
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	return result
}

func doResolveReport(t *testing.T, adminToken string, reportID uuid.UUID, accepted bool, comment *string) *http.Response {
	t.Helper()

	body := types.ResolveOfferReportRequest{Accepted: accepted, Comment: comment}
	req := mustBearerRequest(t, http.MethodPost, dealsURL()+"/admin/offer-reports/"+reportID.String()+"/resolution", adminToken, mustJSONBody(t, body))
	req.Header.Set("Content-Type", "application/json")
	return mustDo(t, req)
}

func mustResolveReport(t *testing.T, adminToken string, reportID uuid.UUID, accepted bool, comment *string) types.OfferReport {
	t.Helper()

	resp := doResolveReport(t, adminToken, reportID, accepted, comment)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var report types.OfferReport
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&report))
	return report
}

// ─────────────────────────────────────────────────────────────────────────────
// Tests
// ─────────────────────────────────────────────────────────────────────────────

func TestCreateOfferReportSuccess(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	authorID := uuid.New()
	reporterID := uuid.New()
	offerID := mustCreateOffer(t, authorID)

	report := mustCreateOfferReport(t, reporterID, offerID, "spam offer")

	require.NotEqual(t, uuid.Nil, report.Id)
	require.Equal(t, offerID, report.OfferId)
	require.Equal(t, authorID, report.OfferAuthorId)
	require.Equal(t, types.Pending, report.Status)
}

func TestCreateOfferReportOfferBecomesBlocked(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	authorID := uuid.New()
	reporterID := uuid.New()
	offerID := mustCreateOffer(t, authorID)

	mustCreateOfferReport(t, reporterID, offerID, "please review")

	// Author cannot update a blocked offer
	req := mustUserRequest(t, http.MethodPatch, dealsURL()+"/offers/"+offerID.String(), authorID, mustJSONBody(t, types.UpdateOfferRequest{
		Name: new(fmt.Sprintf("updated-%d", time.Now().UnixNano())),
	}))
	req.Header.Set("Content-Type", "application/json")
	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusConflict, resp.StatusCode)
}

func TestCreateOfferReportSecondReporterJoinsExisting(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	authorID := uuid.New()
	reporterA := uuid.New()
	reporterB := uuid.New()
	offerID := mustCreateOffer(t, authorID)

	mustCreateOfferReport(t, reporterA, offerID, "first complaint")

	// Second reporter — should return 200 (joined existing report)
	resp := doCreateOfferReport(t, reporterB, offerID, "second complaint")
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var report types.OfferReport
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&report))
	require.Equal(t, types.Pending, report.Status)
}

func TestCreateOfferReportSameReporterReturnsDuplicate(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	authorID := uuid.New()
	reporterID := uuid.New()
	offerID := mustCreateOffer(t, authorID)

	mustCreateOfferReport(t, reporterID, offerID, "duplicate complaint")

	// Same reporter tries again — should be 409
	resp := doCreateOfferReport(t, reporterID, offerID, "duplicate complaint again")
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusConflict, resp.StatusCode)
}

func TestCreateOfferReportSelfReportForbidden(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	authorID := uuid.New()
	offerID := mustCreateOffer(t, authorID)

	resp := doCreateOfferReport(t, authorID, offerID, "self complaint")
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestGetOfferReportsAuthorCanView(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	authorID := uuid.New()
	reporterID := uuid.New()
	offerID := mustCreateOffer(t, authorID)

	mustCreateOfferReport(t, reporterID, offerID, "report for author view")

	result := mustGetOfferReports(t, authorID, offerID)
	require.Equal(t, offerID, result.Offer.Id)
	require.Len(t, result.Reports, 1)
	require.Len(t, result.Reports[0].Messages, 1)
}

func TestGetOfferReportsNonAuthorForbidden(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	authorID := uuid.New()
	reporterID := uuid.New()
	strangerID := uuid.New()
	offerID := mustCreateOffer(t, authorID)

	mustCreateOfferReport(t, reporterID, offerID, "report for forbidden view")

	resp := doGetOfferReports(t, strangerID, offerID)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestAdminListReports(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	adminToken := mustAdminAccessToken(t)
	authorID := uuid.New()
	reporterID := uuid.New()
	offerID := mustCreateOffer(t, authorID)
	report := mustCreateOfferReport(t, reporterID, offerID, "admin pending list")

	reports := mustListAdminReports(t, adminToken, new(types.Pending))

	found := false
	for _, r := range reports {
		if r.Id == report.Id {
			found = true
			break
		}
	}
	require.True(t, found, "created report not found in admin list")
}

func TestAdminGetReportDetails(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	adminToken := mustAdminAccessToken(t)
	authorID := uuid.New()
	reporterID := uuid.New()
	offerID := mustCreateOffer(t, authorID)
	report := mustCreateOfferReport(t, reporterID, offerID, "admin details")

	details := mustGetAdminReportDetails(t, adminToken, report.Id)
	require.Equal(t, report.Id, details.Report.Id)
	require.Equal(t, offerID, details.Offer.Id)
	require.Len(t, details.Messages, 1)
	require.Equal(t, reporterID, details.Messages[0].AuthorId)
	require.Equal(t, "admin details", details.Messages[0].Message)
}

func TestAdminResolveReportAcceptHidesOffer(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	adminToken := mustAdminAccessToken(t)
	authorID := uuid.New()
	reporterID := uuid.New()
	offerID := mustCreateOffer(t, authorID)
	report := mustCreateOfferReport(t, reporterID, offerID, "accept report")

	resolved := mustResolveReport(t, adminToken, report.Id, true, new("violation confirmed"))
	require.Equal(t, types.Accepted, resolved.Status)
	require.NotNil(t, resolved.ResolutionComment)
	require.Equal(t, "violation confirmed", *resolved.ResolutionComment)

	// Stranger cannot see hidden offer
	req := mustUserRequest(t, http.MethodGet, dealsURL()+"/offers/"+offerID.String(), reporterID, nil)
	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusNotFound, resp.StatusCode)

	// Author still can see their own hidden offer
	offerForAuthor := mustGetOfferByID(t, authorID, offerID)
	require.Equal(t, offerID, offerForAuthor.Id)
}

func TestAdminResolveReportAcceptRemovedFromPublicList(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	adminToken := mustAdminAccessToken(t)
	authorID := uuid.New()
	reporterID := uuid.New()
	offerID := mustCreateOffer(t, authorID)
	report := mustCreateOfferReport(t, reporterID, offerID, "remove from list")

	mustResolveReport(t, adminToken, report.Id, true, nil)

	// Offer must not appear in the public list
	strangerID := uuid.New()
	result := mustGetOffers(t, strangerID, nil)
	for _, o := range result.Offers {
		require.NotEqual(t, offerID, o.Id, "hidden offer must not appear in public list")
	}
}

func TestAdminResolveReportRejectKeepsOfferVisible(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	adminToken := mustAdminAccessToken(t)
	authorID := uuid.New()
	reporterID := uuid.New()
	offerID := mustCreateOffer(t, authorID)
	report := mustCreateOfferReport(t, reporterID, offerID, "reject report")

	resolved := mustResolveReport(t, adminToken, report.Id, false, new("no violation"))
	require.Equal(t, types.Rejected, resolved.Status)

	// Offer must still be accessible to strangers
	strangerID := uuid.New()
	fetched := mustGetOfferByID(t, strangerID, offerID)
	require.Equal(t, offerID, fetched.Id)
}

func TestAdminResolveAlreadyResolvedReturnsConflict(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	adminToken := mustAdminAccessToken(t)
	authorID := uuid.New()
	reporterID := uuid.New()
	offerID := mustCreateOffer(t, authorID)
	report := mustCreateOfferReport(t, reporterID, offerID, "already resolved")

	mustResolveReport(t, adminToken, report.Id, false, nil)

	// Resolving again must be 409
	resp := doResolveReport(t, adminToken, report.Id, true, nil)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusConflict, resp.StatusCode)
}

func TestUpdateBlockedOfferReturnsConflict(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	authorID := uuid.New()
	reporterID := uuid.New()
	offerID := mustCreateOffer(t, authorID)

	mustCreateOfferReport(t, reporterID, offerID, "block update")

	req := mustUserRequest(t, http.MethodPatch, dealsURL()+"/offers/"+offerID.String(), authorID, mustJSONBody(t, types.UpdateOfferRequest{
		Name: new(fmt.Sprintf("updated-%d", time.Now().UnixNano())),
	}))
	req.Header.Set("Content-Type", "application/json")
	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusConflict, resp.StatusCode)
}

func TestDeleteBlockedOfferReturnsConflict(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	authorID := uuid.New()
	reporterID := uuid.New()
	offerID := mustCreateOffer(t, authorID)

	mustCreateOfferReport(t, reporterID, offerID, "block delete")

	req := mustUserRequest(t, http.MethodDelete, dealsURL()+"/offers/"+offerID.String(), authorID, nil)
	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusConflict, resp.StatusCode)
}

func TestAdminListReportsFilterByStatus(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	adminToken := mustAdminAccessToken(t)
	authorID := uuid.New()
	reporterID := uuid.New()
	offerID := mustCreateOffer(t, authorID)
	report := mustCreateOfferReport(t, reporterID, offerID, "status filter")

	mustResolveReport(t, adminToken, report.Id, false, nil)

	// Should not appear in pending list
	pendingReports := mustListAdminReports(t, adminToken, new(types.Pending))
	for _, r := range pendingReports {
		require.NotEqual(t, report.Id, r.Id, "resolved report must not appear in pending list")
	}

	// Should appear in rejected list
	rejectedReports := mustListAdminReports(t, adminToken, new(types.Rejected))
	found := false
	for _, r := range rejectedReports {
		if r.Id == report.Id {
			found = true
			break
		}
	}
	require.True(t, found, "resolved report not found in rejected list")
}

func TestAdminGetReportDetailsNonAdminForbidden(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	authorID := uuid.New()
	reporterID := uuid.New()
	offerID := mustCreateOffer(t, authorID)
	report := mustCreateOfferReport(t, reporterID, offerID, "non-admin details")

	resp := doGetAdminReportDetails(t, mustAccessToken(t, uuid.New()), report.Id)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusForbidden, resp.StatusCode)
}
