package seed_demo

import (
	chattypes "barter-port/contracts/openapi/chats/types"
	dealtypes "barter-port/contracts/openapi/deals/types"
	usertypes "barter-port/contracts/openapi/users/types"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	refreshCookieName        = "refresh_token"
	mediaUploadRetryAttempts = 3
)

func (c *SeedClient) register(ctx context.Context, email, password string) (registerResponse, error) {
	var respBody registerResponse
	if err := c.doJSON(ctx, http.MethodPost, "/auth/register", "", registerRequest{
		Email:    email,
		Password: password,
	}, &respBody, http.StatusOK); err != nil {
		return registerResponse{}, err
	}

	return respBody, nil
}

func (c *SeedClient) retrySendVerificationEmail(ctx context.Context, email, password string) error {
	return c.doJSON(ctx, http.MethodPost, "/auth/retry-send-verification-email", "", registerRequest{
		Email:    email,
		Password: password,
	}, nil, http.StatusOK)
}

func (c *SeedClient) verifyEmail(ctx context.Context, token string) error {
	return c.doJSON(ctx, http.MethodPost, "/auth/verify-email", "", verifyEmailRequest{
		Token: token,
	}, nil, http.StatusOK)
}

func (c *SeedClient) ensureUser(ctx context.Context, email, password string) (registerResponse, string, error) {
	lastSeenID := uuid.Nil
	if c.smtp4devConfigured() {
		inboxLastSeenID, inboxErr := c.waitForLatestSMTP4DevMessageID(ctx)
		if inboxErr != nil {
			return registerResponse{}, "", fmt.Errorf("prepare smtp4dev checkpoint: %w", inboxErr)
		}
		lastSeenID = inboxLastSeenID
	}

	registered, err := c.register(ctx, email, password)
	if err == nil {
		if err := c.waitForAuthProvisioning(ctx, registered.UserID); err != nil {
			return registerResponse{}, "", err
		}

		if c.smtp4devConfigured() {
			tokenValue, err := c.waitForVerificationTokenFromEmail(ctx, email, lastSeenID)
			if err != nil {
				return registerResponse{}, "", err
			}
			if err := c.verifyEmail(ctx, tokenValue); err != nil {
				return registerResponse{}, "", fmt.Errorf("verify email %s: %w", email, err)
			}
		}

		token, err := c.login(ctx, email, password)
		if err != nil {
			return registerResponse{}, "", err
		}

		if err := c.waitForUsersProjection(ctx, token); err != nil {
			return registerResponse{}, "", err
		}

		return registered, token, nil
	}

	if !strings.Contains(err.Error(), "email already in use") {
		return registerResponse{}, "", err
	}

	token, err := c.login(ctx, email, password)
	if err != nil && strings.Contains(strings.ToLower(err.Error()), "email not verified") {
		if !c.smtp4devConfigured() {
			return registerResponse{}, "", fmt.Errorf("user %s is not verified and smtp4dev is disabled", email)
		}
		lastSeenID, inboxErr := c.waitForLatestSMTP4DevMessageID(ctx)
		if inboxErr != nil {
			return registerResponse{}, "", fmt.Errorf("prepare smtp4dev checkpoint: %w", inboxErr)
		}
		if err := c.retrySendVerificationEmail(ctx, email, password); err != nil {
			return registerResponse{}, "", fmt.Errorf("retry verification email for %s: %w", email, err)
		}
		tokenValue, err := c.waitForVerificationTokenFromEmail(ctx, email, lastSeenID)
		if err != nil {
			return registerResponse{}, "", err
		}
		if err := c.verifyEmail(ctx, tokenValue); err != nil {
			return registerResponse{}, "", fmt.Errorf("verify email %s: %w", email, err)
		}
		token, err = c.login(ctx, email, password)
	}
	if err != nil {
		return registerResponse{}, "", err
	}

	if err := c.waitForUsersProjection(ctx, token); err != nil {
		return registerResponse{}, "", err
	}

	me, err := c.getMe(ctx, token)
	if err != nil {
		return registerResponse{}, "", err
	}

	return registerResponse{
		UserID: me.Id,
		Email:  string(me.Email),
	}, token, nil
}

func (c *SeedClient) ensureAdminToken(ctx context.Context, email, password string) (string, error) {
	token, err := c.login(ctx, email, password)
	if err == nil {
		return token, nil
	}

	lowerErr := strings.ToLower(err.Error())
	if strings.Contains(lowerErr, "email not verified") {
		if !c.smtp4devConfigured() {
			return "", fmt.Errorf("admin account %s is not verified and smtp4dev is disabled", email)
		}
		lastSeenID, inboxErr := c.waitForLatestSMTP4DevMessageID(ctx)
		if inboxErr != nil {
			return "", fmt.Errorf("prepare smtp4dev checkpoint for admin: %w", inboxErr)
		}
		if err := c.retrySendVerificationEmail(ctx, email, password); err != nil {
			return "", fmt.Errorf("retry verification email for admin %s: %w", email, err)
		}
		tokenValue, err := c.waitForVerificationTokenFromEmail(ctx, email, lastSeenID)
		if err != nil {
			return "", err
		}
		if err := c.verifyEmail(ctx, tokenValue); err != nil {
			return "", fmt.Errorf("verify admin email %s: %w", email, err)
		}

		token, err = c.login(ctx, email, password)
		if err != nil {
			return "", err
		}
		return token, nil
	}

	if !strings.Contains(lowerErr, "invalid credentials") {
		return "", err
	}

	lastSeenID := uuid.Nil
	if c.smtp4devConfigured() {
		inboxLastSeenID, inboxErr := c.waitForLatestSMTP4DevMessageID(ctx)
		if inboxErr != nil {
			return "", fmt.Errorf("prepare smtp4dev checkpoint for admin: %w", inboxErr)
		}
		lastSeenID = inboxLastSeenID
	}

	registered, registerErr := c.register(ctx, email, password)
	if registerErr == nil {
		if err := c.waitForAuthProvisioning(ctx, registered.UserID); err != nil {
			return "", err
		}

		if c.smtp4devConfigured() {
			tokenValue, err := c.waitForVerificationTokenFromEmail(ctx, email, lastSeenID)
			if err != nil {
				return "", err
			}
			if err := c.verifyEmail(ctx, tokenValue); err != nil {
				return "", fmt.Errorf("verify admin email %s: %w", email, err)
			}
		}

		token, err := c.login(ctx, email, password)
		if err != nil {
			return "", err
		}
		return token, nil
	}

	if strings.Contains(strings.ToLower(registerErr.Error()), "email already in use") {
		return "", fmt.Errorf("admin account %s already exists with a different password; set SEED_ADMIN_PASSWORD to the real password or run reseed-demo", email)
	}

	if strings.Contains(strings.ToLower(registerErr.Error()), "password too short") {
		return "", fmt.Errorf("admin account %s is not accessible with the provided password, and fallback registration is impossible because the password is shorter than the public auth minimum; set SEED_ADMIN_PASSWORD to the real password or use a 6+ character admin password in config/common.yaml for fresh environments", email)
	}

	return "", fmt.Errorf("ensure admin %s: login failed with %v; register fallback failed with %w", email, err, registerErr)
}

func (c *SeedClient) waitForAuthProvisioning(ctx context.Context, userID uuid.UUID) error {
	return c.poll(ctx, func(ctx context.Context) (bool, error) {
		resp, err := c.do(ctx, http.MethodGet, "/auth/status/"+userID.String(), "", nil)
		if err != nil {
			return false, err
		}
		defer closeBody(resp.Body)

		if resp.StatusCode == http.StatusNotFound {
			return false, nil
		}
		if resp.StatusCode != http.StatusOK {
			return false, responseError(resp, http.StatusOK)
		}

		var status authStatusResponse
		if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
			return false, fmt.Errorf("decode auth status: %w", err)
		}

		switch status.Status {
		case "Success":
			return true, nil
		case "Failed":
			return false, errors.New("auth provisioning failed")
		default:
			return false, nil
		}
	})
}

func (c *SeedClient) login(ctx context.Context, email, password string) (string, error) {
	var body loginResponse
	if err := c.doJSON(ctx, http.MethodPost, "/auth/login", "", registerRequest{
		Email:    email,
		Password: password,
	}, &body, http.StatusOK); err != nil {
		return "", err
	}

	if body.AccessToken == "" {
		return "", errors.New("login response has empty access token")
	}

	return body.AccessToken, nil
}

func (c *SeedClient) loginWithRefreshCookie(ctx context.Context, email, password string) (string, string, error) {
	resp, err := c.do(ctx, http.MethodPost, "/auth/login", "", registerRequest{
		Email:    email,
		Password: password,
	})
	if err != nil {
		return "", "", err
	}
	defer closeBody(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return "", "", responseError(resp, http.StatusOK)
	}

	var body loginResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "", "", fmt.Errorf("decode response POST /auth/login: %w", err)
	}
	if body.AccessToken == "" {
		return "", "", errors.New("login response has empty access token")
	}

	for _, cookie := range resp.Cookies() {
		if cookie.Name == refreshCookieName && cookie.Value != "" {
			return body.AccessToken, cookie.Value, nil
		}
	}

	return "", "", errors.New("login response has empty refresh_token cookie")
}

func (c *SeedClient) refresh(ctx context.Context, refreshCookie string) (string, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/auth/refresh", nil)
	if err != nil {
		return "", "", fmt.Errorf("build request POST /auth/refresh: %w", err)
	}
	req.Header.Set("Cookie", refreshCookieName+"="+refreshCookie)

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("perform request POST /auth/refresh: %w", err)
	}
	defer closeBody(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return "", "", responseError(resp, http.StatusOK)
	}

	var body refreshResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "", "", fmt.Errorf("decode response POST /auth/refresh: %w", err)
	}
	if body.AccessToken == "" {
		return "", "", errors.New("refresh response has empty access token")
	}

	for _, cookie := range resp.Cookies() {
		if cookie.Name == refreshCookieName && cookie.Value != "" {
			return body.AccessToken, cookie.Value, nil
		}
	}

	return "", "", errors.New("refresh response has empty refresh_token cookie")
}

func (c *SeedClient) logout(ctx context.Context, refreshCookie string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/auth/logout", nil)
	if err != nil {
		return fmt.Errorf("build request POST /auth/logout: %w", err)
	}
	req.Header.Set("Cookie", refreshCookieName+"="+refreshCookie)

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return fmt.Errorf("perform request POST /auth/logout: %w", err)
	}
	defer closeBody(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return responseError(resp, http.StatusOK)
	}

	return nil
}

func (c *SeedClient) waitForUsersProjection(ctx context.Context, token string) error {
	return c.poll(ctx, func(ctx context.Context) (bool, error) {
		resp, err := c.do(ctx, http.MethodGet, "/users/me", token, nil)
		if err != nil {
			return false, err
		}
		defer closeBody(resp.Body)

		switch resp.StatusCode {
		case http.StatusOK:
			return true, nil
		case http.StatusNotFound:
			return false, nil
		default:
			return false, responseError(resp, http.StatusOK, http.StatusNotFound)
		}
	})
}

func (c *SeedClient) updateMe(ctx context.Context, token string, req usertypes.UpdateUserRequest) (usertypes.Me, error) {
	var body usertypes.Me
	if err := c.doJSON(ctx, http.MethodPatch, "/users/me", token, req, &body, http.StatusOK); err != nil {
		return usertypes.Me{}, err
	}

	return body, nil
}

func (c *SeedClient) uploadMeAvatar(ctx context.Context, token string, avatarPath string) (usertypes.AvatarUploadResponse, error) {
	return retryMediaUpload(ctx, c.PollInterval, mediaUploadRetryAttempts, func() (usertypes.AvatarUploadResponse, error) {
		var body bytes.Buffer
		writer := multipart.NewWriter(&body)

		file, err := os.Open(avatarPath)
		if err != nil {
			return usertypes.AvatarUploadResponse{}, fmt.Errorf("open avatar %s: %w", avatarPath, err)
		}
		defer closeBody(file)

		part, err := writer.CreateFormFile("file", filepath.Base(avatarPath))
		if err != nil {
			return usertypes.AvatarUploadResponse{}, fmt.Errorf("create multipart avatar field for %s: %w", avatarPath, err)
		}
		if _, err := io.Copy(part, file); err != nil {
			return usertypes.AvatarUploadResponse{}, fmt.Errorf("write avatar %s: %w", avatarPath, err)
		}
		if err := writer.Close(); err != nil {
			return usertypes.AvatarUploadResponse{}, fmt.Errorf("close avatar multipart writer: %w", err)
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/users/me/avatar", &body)
		if err != nil {
			return usertypes.AvatarUploadResponse{}, fmt.Errorf("build request POST /users/me/avatar: %w", err)
		}

		req.Header.Set("Content-Type", writer.FormDataContentType())
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}

		resp, err := c.HttpClient.Do(req)
		if err != nil {
			return usertypes.AvatarUploadResponse{}, fmt.Errorf("perform request POST /users/me/avatar: %w", err)
		}
		defer closeBody(resp.Body)

		if resp.StatusCode != http.StatusOK {
			return usertypes.AvatarUploadResponse{}, responseError(resp, http.StatusOK)
		}

		var uploaded usertypes.AvatarUploadResponse
		if err := json.NewDecoder(resp.Body).Decode(&uploaded); err != nil {
			return usertypes.AvatarUploadResponse{}, fmt.Errorf("decode response POST /users/me/avatar: %w", err)
		}

		return uploaded, nil
	})
}

func (c *SeedClient) getMe(ctx context.Context, token string) (usertypes.Me, error) {
	var body usertypes.Me
	if err := c.doJSON(ctx, http.MethodGet, "/users/me", token, nil, &body, http.StatusOK); err != nil {
		return usertypes.Me{}, err
	}

	return body, nil
}

func (c *SeedClient) getUserByID(ctx context.Context, token string, userID uuid.UUID) (usertypes.User, error) {
	var body usertypes.User
	if err := c.doJSON(ctx, http.MethodGet, "/users/"+userID.String(), token, nil, &body, http.StatusOK); err != nil {
		return usertypes.User{}, err
	}

	return body, nil
}

func (c *SeedClient) getReputationEvents(ctx context.Context, token string) (usertypes.GetReputationEventsResponse, error) {
	var body usertypes.GetReputationEventsResponse
	if err := c.doJSON(ctx, http.MethodGet, "/users/reputation-events", token, nil, &body, http.StatusOK); err != nil {
		return nil, err
	}

	return body, nil
}

func (c *SeedClient) listSubscriptions(ctx context.Context, token string) (usertypes.GetSubscriptionsResponse, error) {
	var body usertypes.GetSubscriptionsResponse
	if err := c.doJSON(ctx, http.MethodGet, "/users/subscriptions", token, nil, &body, http.StatusOK); err != nil {
		return nil, err
	}

	return body, nil
}

func (c *SeedClient) listSubscriptionsByUser(ctx context.Context, token string, userID uuid.UUID) (usertypes.GetSubscriptionsResponse, error) {
	var body usertypes.GetSubscriptionsResponse
	path := fmt.Sprintf("/users/subscriptions/%s", userID)
	if err := c.doJSON(ctx, http.MethodGet, path, token, nil, &body, http.StatusOK); err != nil {
		return nil, err
	}

	return body, nil
}

func (c *SeedClient) listMySubscribers(ctx context.Context, token string) (usertypes.GetSubscriptionsResponse, error) {
	var body usertypes.GetSubscriptionsResponse
	if err := c.doJSON(ctx, http.MethodGet, "/users/subscribers", token, nil, &body, http.StatusOK); err != nil {
		return nil, err
	}

	return body, nil
}

func (c *SeedClient) listSubscribersByUser(ctx context.Context, token string, userID uuid.UUID) (usertypes.GetSubscriptionsResponse, error) {
	var body usertypes.GetSubscriptionsResponse
	path := fmt.Sprintf("/users/subscribers/%s", userID)
	if err := c.doJSON(ctx, http.MethodGet, path, token, nil, &body, http.StatusOK); err != nil {
		return nil, err
	}

	return body, nil
}

func (c *SeedClient) createOffers(ctx context.Context, user *seededUser, specs []offerSpec) (map[string]uuid.UUID, error) {
	result, _, err := c.createOffersWithWarnings(ctx, user, specs)
	return result, err
}

func (c *SeedClient) createOffersWithWarnings(ctx context.Context, user *seededUser, specs []offerSpec) (map[string]uuid.UUID, []string, error) {
	result := make(map[string]uuid.UUID, len(specs))
	warnings := make([]string, 0)
	for _, spec := range specs {
		photoPath, err := resolveOfferPhotoPath(user.Key, spec)
		if err != nil {
			return nil, nil, fmt.Errorf("resolve photo for offer %s for %s: %w", spec.Key, user.Key, err)
		}

		offer, err := c.createOffer(ctx, user.Token, spec, photoPath)
		if err != nil && photoPath != "" && isMediaUploadFallbackable(err) {
			warnings = append(warnings, fmt.Sprintf("offer photo skipped for %s/%s: %v", user.Key, spec.Key, err))
			offer, err = c.createOfferJSON(ctx, user.Token, spec)
		}
		if err != nil {
			return nil, nil, fmt.Errorf("create offer %s for %s: %w", spec.Key, user.Key, err)
		}

		result[spec.Key] = offer.Id
	}

	return result, warnings, nil
}

func (c *SeedClient) createOfferJSON(ctx context.Context, token string, spec offerSpec) (dealtypes.Offer, error) {
	req := dealtypes.CreateOfferRequest{
		Name:        spec.Name,
		Description: spec.Description,
		Type:        spec.Type,
		Action:      spec.Action,
		Latitude:    spec.Latitude,
		Longitude:   spec.Longitude,
	}
	if len(spec.Tags) > 0 {
		req.Tags = &spec.Tags
	}

	var offer dealtypes.Offer
	if err := c.doJSON(ctx, http.MethodPost, "/offers", token, req, &offer, http.StatusCreated); err != nil {
		return dealtypes.Offer{}, err
	}

	return offer, nil
}

func (c *SeedClient) createOffer(ctx context.Context, token string, spec offerSpec, photoPath string) (dealtypes.Offer, error) {
	return retryMediaUpload(ctx, c.PollInterval, mediaUploadRetryAttempts, func() (dealtypes.Offer, error) {
		var body bytes.Buffer
		writer := multipart.NewWriter(&body)

		if err := writer.WriteField("name", spec.Name); err != nil {
			return dealtypes.Offer{}, fmt.Errorf("write offer name: %w", err)
		}
		if err := writer.WriteField("description", spec.Description); err != nil {
			return dealtypes.Offer{}, fmt.Errorf("write offer description: %w", err)
		}
		if err := writer.WriteField("type", string(spec.Type)); err != nil {
			return dealtypes.Offer{}, fmt.Errorf("write offer type: %w", err)
		}
		if err := writer.WriteField("action", string(spec.Action)); err != nil {
			return dealtypes.Offer{}, fmt.Errorf("write offer action: %w", err)
		}
		for _, tag := range spec.Tags {
			if err := writer.WriteField("tags", string(tag)); err != nil {
				return dealtypes.Offer{}, fmt.Errorf("write offer tag: %w", err)
			}
		}

		if spec.Latitude != nil {
			if err := writer.WriteField("latitude", strconv.FormatFloat(*spec.Latitude, 'f', 6, 64)); err != nil {
				return dealtypes.Offer{}, fmt.Errorf("write offer latitude: %w", err)
			}
			if err := writer.WriteField("longitude", strconv.FormatFloat(*spec.Longitude, 'f', 6, 64)); err != nil {
				return dealtypes.Offer{}, fmt.Errorf("write offer longitude: %w", err)
			}
		}

		if photoPath != "" {
			if err := writeOfferPhotoPart(writer, photoPath); err != nil {
				return dealtypes.Offer{}, err
			}
		}

		if err := writer.Close(); err != nil {
			return dealtypes.Offer{}, fmt.Errorf("close multipart writer: %w", err)
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/offers", &body)
		if err != nil {
			return dealtypes.Offer{}, fmt.Errorf("build request POST /offers: %w", err)
		}

		req.Header.Set("Content-Type", writer.FormDataContentType())
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}

		resp, err := c.HttpClient.Do(req)
		if err != nil {
			return dealtypes.Offer{}, fmt.Errorf("perform request POST /offers: %w", err)
		}
		defer closeBody(resp.Body)

		if resp.StatusCode != http.StatusCreated {
			return dealtypes.Offer{}, responseError(resp, http.StatusCreated)
		}

		var offer dealtypes.Offer
		if err := json.NewDecoder(resp.Body).Decode(&offer); err != nil {
			return dealtypes.Offer{}, fmt.Errorf("decode response POST /offers: %w", err)
		}

		return offer, nil
	})
}

func writeOfferPhotoPart(writer *multipart.Writer, photoPath string) error {
	file, err := os.Open(photoPath)
	if err != nil {
		return fmt.Errorf("open offer photo %s: %w", photoPath, err)
	}
	defer closeBody(file)

	part, err := writer.CreateFormFile("photos", filepath.Base(photoPath))
	if err != nil {
		return fmt.Errorf("create multipart photo field for %s: %w", photoPath, err)
	}

	if _, err := io.Copy(part, file); err != nil {
		return fmt.Errorf("write offer photo %s: %w", photoPath, err)
	}

	return nil
}

func (c *SeedClient) listOffers(ctx context.Context, token string, query url.Values) (dealtypes.ListOffersResponse, error) {
	var body dealtypes.ListOffersResponse
	if err := c.doJSON(ctx, http.MethodGet, pathWithQuery("/offers", query), token, nil, &body, http.StatusOK); err != nil {
		return dealtypes.ListOffersResponse{}, err
	}

	return body, nil
}

func (c *SeedClient) listSubscribedOffers(ctx context.Context, token string, query url.Values) (dealtypes.ListOffersResponse, error) {
	var body dealtypes.ListOffersResponse
	if err := c.doJSON(ctx, http.MethodGet, pathWithQuery("/offers/subscriptions", query), token, nil, &body, http.StatusOK); err != nil {
		return dealtypes.ListOffersResponse{}, err
	}

	return body, nil
}

func (c *SeedClient) listFavoriteOffers(ctx context.Context, token string, query url.Values) (dealtypes.ListFavoriteOffersResponse, error) {
	var body dealtypes.ListFavoriteOffersResponse
	if err := c.doJSON(ctx, http.MethodGet, pathWithQuery("/offers/favorites", query), token, nil, &body, http.StatusOK); err != nil {
		return dealtypes.ListFavoriteOffersResponse{}, err
	}

	return body, nil
}

func (c *SeedClient) getOfferByID(ctx context.Context, token string, offerID uuid.UUID) (dealtypes.Offer, error) {
	var body dealtypes.Offer
	if err := c.doJSON(ctx, http.MethodGet, "/offers/"+offerID.String(), token, nil, &body, http.StatusOK); err != nil {
		return dealtypes.Offer{}, err
	}

	return body, nil
}

func (c *SeedClient) updateOffer(ctx context.Context, token string, offerID uuid.UUID, req dealtypes.UpdateOfferRequest) (dealtypes.Offer, error) {
	var body dealtypes.Offer
	path := fmt.Sprintf("/offers/%s", offerID)
	if err := c.doJSON(ctx, http.MethodPatch, path, token, req, &body, http.StatusOK); err != nil {
		return dealtypes.Offer{}, err
	}

	return body, nil
}

func (c *SeedClient) deleteOffer(ctx context.Context, token string, offerID uuid.UUID) error {
	return c.doJSON(ctx, http.MethodDelete, "/offers/"+offerID.String(), token, nil, nil, http.StatusNoContent)
}

func (c *SeedClient) addOfferToFavorites(ctx context.Context, token string, offerID uuid.UUID) error {
	path := fmt.Sprintf("/offers/%s/favorite", offerID)
	return c.doJSON(ctx, http.MethodPut, path, token, nil, nil, http.StatusNoContent)
}

func (c *SeedClient) removeOfferFromFavorites(ctx context.Context, token string, offerID uuid.UUID) error {
	path := fmt.Sprintf("/offers/%s/favorite", offerID)
	return c.doJSON(ctx, http.MethodDelete, path, token, nil, nil, http.StatusNoContent)
}

func (c *SeedClient) viewOffer(ctx context.Context, token string, offerID uuid.UUID) error {
	path := fmt.Sprintf("/offers/%s/view", offerID)
	return c.doJSON(ctx, http.MethodPost, path, token, nil, nil, http.StatusOK)
}

func (c *SeedClient) hideOfferByAuthor(ctx context.Context, token string, offerID uuid.UUID) error {
	path := fmt.Sprintf("/offers/%s/hidden", offerID)
	return c.doJSON(ctx, http.MethodPut, path, token, nil, nil, http.StatusNoContent)
}

func (c *SeedClient) unhideOfferByAuthor(ctx context.Context, token string, offerID uuid.UUID) error {
	path := fmt.Sprintf("/offers/%s/hidden", offerID)
	return c.doJSON(ctx, http.MethodDelete, path, token, nil, nil, http.StatusNoContent)
}

func (c *SeedClient) hideUser(ctx context.Context, token string, targetUserID uuid.UUID) error {
	return c.doJSON(ctx, http.MethodPost, "/users/hidden-users", token,
		map[string]uuid.UUID{"targetUserId": targetUserID}, nil, http.StatusNoContent)
}

func (c *SeedClient) listOfferGroupsQuery(ctx context.Context, token string, query url.Values) (dealtypes.ListOfferGroupsResponse, error) {
	var body dealtypes.ListOfferGroupsResponse
	if err := c.doJSON(ctx, http.MethodGet, pathWithQuery("/offer-groups", query), token, nil, &body, http.StatusOK); err != nil {
		return nil, err
	}
	return body, nil
}

func (c *SeedClient) getOfferReports(ctx context.Context, token string, offerID uuid.UUID) (dealtypes.OfferReportsForOffer, error) {
	var body dealtypes.OfferReportsForOffer
	path := fmt.Sprintf("/offers/%s/reports", offerID)
	if err := c.doJSON(ctx, http.MethodGet, path, token, nil, &body, http.StatusOK); err != nil {
		return dealtypes.OfferReportsForOffer{}, err
	}

	return body, nil
}

func (c *SeedClient) listOfferReportsForAdmin(ctx context.Context, token string, status *dealtypes.OfferReportStatus) (dealtypes.ListOfferReportsResponse, error) {
	query := url.Values{}
	if status != nil {
		query.Set("status", string(*status))
	}

	var body dealtypes.ListOfferReportsResponse
	if err := c.doJSON(ctx, http.MethodGet, pathWithQuery("/admin/offer-reports", query), token, nil, &body, http.StatusOK); err != nil {
		return nil, err
	}

	return body, nil
}

func (c *SeedClient) getOfferReportForAdmin(ctx context.Context, token string, reportID uuid.UUID) (dealtypes.OfferReportDetails, error) {
	var body dealtypes.OfferReportDetails
	path := fmt.Sprintf("/admin/offer-reports/%s", reportID)
	if err := c.doJSON(ctx, http.MethodGet, path, token, nil, &body, http.StatusOK); err != nil {
		return dealtypes.OfferReportDetails{}, err
	}

	return body, nil
}

func (c *SeedClient) listTags(ctx context.Context, token string) (dealtypes.ListTagsResponse, error) {
	var body dealtypes.ListTagsResponse
	if err := c.doJSON(ctx, http.MethodGet, "/tags", token, nil, &body, http.StatusOK); err != nil {
		return nil, err
	}

	return body, nil
}

func (c *SeedClient) deleteTag(ctx context.Context, token string, name string) error {
	query := url.Values{}
	query.Set("name", name)
	return c.doJSON(ctx, http.MethodDelete, pathWithQuery("/admin/tags", query), token, nil, nil, http.StatusNoContent)
}

func (c *SeedClient) createOfferGroup(ctx context.Context, token string, req offerGroupRequest) (uuid.UUID, error) {
	var body offerGroupResponse
	if err := c.doJSON(ctx, http.MethodPost, "/offer-groups", token, req, &body, http.StatusCreated); err != nil {
		return uuid.Nil, err
	}

	return body.ID, nil
}

func (c *SeedClient) listOfferGroups(ctx context.Context, token string) (dealtypes.ListOfferGroupsResponse, error) {
	var body dealtypes.ListOfferGroupsResponse
	if err := c.doJSON(ctx, http.MethodGet, "/offer-groups", token, nil, &body, http.StatusOK); err != nil {
		return nil, err
	}

	return body, nil
}

func (c *SeedClient) getOfferGroupByID(ctx context.Context, token string, offerGroupID uuid.UUID) (dealtypes.OfferGroup, error) {
	var body dealtypes.OfferGroup
	path := fmt.Sprintf("/offer-groups/%s", offerGroupID)
	if err := c.doJSON(ctx, http.MethodGet, path, token, nil, &body, http.StatusOK); err != nil {
		return dealtypes.OfferGroup{}, err
	}

	return body, nil
}

func (c *SeedClient) createDraftFromOfferGroup(ctx context.Context, token string, offerGroupID uuid.UUID, req offerGroupDraftRequest) (uuid.UUID, error) {
	var body dealtypes.CreateDraftDealResponse
	path := fmt.Sprintf("/offer-groups/%s/drafts", offerGroupID)
	if err := c.doJSON(ctx, http.MethodPost, path, token, req, &body, http.StatusCreated); err != nil {
		return uuid.Nil, err
	}

	return body.Id, nil
}

func (c *SeedClient) listMyDeals(ctx context.Context, token string) (dealtypes.GetDealsResponse, error) {
	var deals dealtypes.GetDealsResponse
	if err := c.doJSON(ctx, http.MethodGet, "/deals?my=true", token, nil, &deals, http.StatusOK); err != nil {
		return nil, err
	}

	return deals, nil
}

func (c *SeedClient) listDeals(ctx context.Context, token string, query url.Values) (dealtypes.GetDealsResponse, error) {
	var body dealtypes.GetDealsResponse
	if err := c.doJSON(ctx, http.MethodGet, pathWithQuery("/deals", query), token, nil, &body, http.StatusOK); err != nil {
		return nil, err
	}

	return body, nil
}

func (c *SeedClient) createDraft(ctx context.Context, token string, req dealtypes.CreateDraftDealRequest) (uuid.UUID, error) {
	var body dealtypes.CreateDraftDealResponse
	if err := c.doJSON(ctx, http.MethodPost, "/deals/drafts", token, req, &body, http.StatusCreated); err != nil {
		return uuid.Nil, err
	}

	return body.Id, nil
}

func (c *SeedClient) listDrafts(ctx context.Context, token string, createdByMe, participating bool) (dealtypes.GetMyDraftDealsResponse, error) {
	query := url.Values{}
	query.Set("createdByMe", fmt.Sprintf("%t", createdByMe))
	query.Set("participating", fmt.Sprintf("%t", participating))

	var body dealtypes.GetMyDraftDealsResponse
	if err := c.doJSON(ctx, http.MethodGet, pathWithQuery("/deals/drafts", query), token, nil, &body, http.StatusOK); err != nil {
		return nil, err
	}

	return body, nil
}

func (c *SeedClient) getDraftByID(ctx context.Context, token string, draftID uuid.UUID) (dealtypes.Draft, error) {
	var body dealtypes.Draft
	path := fmt.Sprintf("/deals/drafts/%s", draftID)
	if err := c.doJSON(ctx, http.MethodGet, path, token, nil, &body, http.StatusOK); err != nil {
		return dealtypes.Draft{}, err
	}

	return body, nil
}

func (c *SeedClient) confirmDraft(ctx context.Context, token string, draftID uuid.UUID) error {
	return c.doJSON(ctx, http.MethodPatch, "/deals/drafts/"+draftID.String(), token, nil, nil, http.StatusOK)
}

func (c *SeedClient) cancelDraft(ctx context.Context, token string, draftID uuid.UUID) error {
	return c.doJSON(ctx, http.MethodPatch, "/deals/drafts/"+draftID.String()+"/cancel", token, nil, nil, http.StatusOK)
}

func (c *SeedClient) deleteDraft(ctx context.Context, token string, draftID uuid.UUID) error {
	return c.doJSON(ctx, http.MethodDelete, "/deals/drafts/"+draftID.String(), token, nil, nil, http.StatusOK)
}

func (c *SeedClient) createTwoPartyDeal(
	ctx context.Context,
	userA *seededUser,
	userB *seededUser,
	offerA uuid.UUID,
	offerB uuid.UUID,
	name string,
	description string,
) (uuid.UUID, error) {
	before, err := c.listMyDeals(ctx, userA.Token)
	if err != nil {
		return uuid.Nil, fmt.Errorf("list deals before draft: %w", err)
	}

	draftID, err := c.createDraft(ctx, userA.Token, dealtypes.CreateDraftDealRequest{
		Name:        &name,
		Description: &description,
		Offers: []dealtypes.OfferIDAndQuantity{
			{OfferID: offerA, Quantity: 1},
			{OfferID: offerB, Quantity: 1},
		},
	})
	if err != nil {
		return uuid.Nil, err
	}

	if err := c.confirmDraft(ctx, userA.Token, draftID); err != nil {
		return uuid.Nil, fmt.Errorf("confirm draft by %s: %w", userA.Key, err)
	}
	if err := c.confirmDraft(ctx, userB.Token, draftID); err != nil {
		return uuid.Nil, fmt.Errorf("confirm draft by %s: %w", userB.Key, err)
	}

	return c.waitForNewDeal(ctx, userA.Token, before)
}

func (c *SeedClient) waitForNewDeal(ctx context.Context, token string, before dealtypes.GetDealsResponse) (uuid.UUID, error) {
	beforeSet := make(map[uuid.UUID]struct{}, len(before))
	for _, deal := range before {
		beforeSet[deal.Id] = struct{}{}
	}

	var created uuid.UUID
	err := c.poll(ctx, func(ctx context.Context) (bool, error) {
		after, err := c.listMyDeals(ctx, token)
		if err != nil {
			return false, err
		}

		for _, deal := range after {
			id := deal.Id
			if _, ok := beforeSet[id]; !ok {
				created = id
				return true, nil
			}
		}

		return false, nil
	})
	if err != nil {
		return uuid.Nil, err
	}

	return created, nil
}

func (c *SeedClient) getDealByID(ctx context.Context, token string, dealID uuid.UUID) (dealtypes.Deal, error) {
	var body dealtypes.Deal
	if err := c.doJSON(ctx, http.MethodGet, "/deals/"+dealID.String(), token, nil, &body, http.StatusOK); err != nil {
		return dealtypes.Deal{}, err
	}

	return body, nil
}

func (c *SeedClient) updateDeal(ctx context.Context, token string, dealID uuid.UUID, req dealtypes.UpdateDealRequest) (dealtypes.Deal, error) {
	var body dealtypes.Deal
	path := fmt.Sprintf("/deals/%s", dealID)
	if err := c.doJSON(ctx, http.MethodPatch, path, token, req, &body, http.StatusOK); err != nil {
		return dealtypes.Deal{}, err
	}

	return body, nil
}

func (c *SeedClient) updateDealItem(ctx context.Context, token string, dealID uuid.UUID, itemID uuid.UUID, req dealtypes.UpdateDealItemRequest) error {
	path := fmt.Sprintf("/deals/%s/items/%s", dealID, itemID)
	return c.doJSON(ctx, http.MethodPatch, path, token, req, nil, http.StatusOK)
}

func (c *SeedClient) getDealStatusVotes(ctx context.Context, token string, dealID uuid.UUID) (dealtypes.GetDealStatusVotesResponse, error) {
	var body dealtypes.GetDealStatusVotesResponse
	path := fmt.Sprintf("/deals/%s/status", dealID)
	if err := c.doJSON(ctx, http.MethodGet, path, token, nil, &body, http.StatusOK); err != nil {
		return nil, err
	}

	return body, nil
}

func (c *SeedClient) addDealItem(ctx context.Context, token string, dealID uuid.UUID, req dealtypes.AddDealItemRequest) (dealtypes.Deal, error) {
	var body dealtypes.Deal
	path := fmt.Sprintf("/deals/%s/items", dealID)
	if err := c.doJSON(ctx, http.MethodPost, path, token, req, &body, http.StatusOK); err != nil {
		return dealtypes.Deal{}, err
	}

	return body, nil
}

func (c *SeedClient) changeDealStatus(ctx context.Context, token string, dealID uuid.UUID, status dealtypes.DealStatus) error {
	return c.doJSON(ctx, http.MethodPatch, "/deals/"+dealID.String()+"/status", token, dealtypes.ChangeDealStatusRequest{
		ExpectedStatus: status,
	}, nil, http.StatusOK)
}

func (c *SeedClient) promoteDealToDiscussion(ctx context.Context, dealID uuid.UUID, userA *seededUser, userB *seededUser) (dealtypes.Deal, error) {
	deal, err := c.getDealByID(ctx, userA.Token, dealID)
	if err != nil {
		return dealtypes.Deal{}, err
	}

	itemByAuthor := make(map[uuid.UUID]uuid.UUID, len(deal.Items))
	for _, item := range deal.Items {
		itemByAuthor[item.AuthorId] = item.Id
	}

	itemA, ok := itemByAuthor[userA.UserID]
	if !ok {
		return dealtypes.Deal{}, fmt.Errorf("deal %s does not contain item authored by %s", dealID, userA.Key)
	}
	itemB, ok := itemByAuthor[userB.UserID]
	if !ok {
		return dealtypes.Deal{}, fmt.Errorf("deal %s does not contain item authored by %s", dealID, userB.Key)
	}

	if err := c.updateDealItem(ctx, userB.Token, dealID, itemA, dealtypes.UpdateDealItemRequest{
		ClaimReceiver: new(true),
	}); err != nil {
		return dealtypes.Deal{}, fmt.Errorf("claim receiver for %s item: %w", userA.Key, err)
	}

	if err := c.updateDealItem(ctx, userA.Token, dealID, itemB, dealtypes.UpdateDealItemRequest{
		ClaimReceiver: new(true),
	}); err != nil {
		return dealtypes.Deal{}, fmt.Errorf("claim receiver for %s item: %w", userB.Key, err)
	}

	if err := c.changeDealStatus(ctx, userA.Token, dealID, dealtypes.Discussion); err != nil {
		return dealtypes.Deal{}, fmt.Errorf("discussion vote by %s: %w", userA.Key, err)
	}
	if err := c.changeDealStatus(ctx, userB.Token, dealID, dealtypes.Discussion); err != nil {
		return dealtypes.Deal{}, fmt.Errorf("discussion vote by %s: %w", userB.Key, err)
	}

	return c.getDealByID(ctx, userA.Token, dealID)
}

func (c *SeedClient) completeTwoPartyDeal(ctx context.Context, dealID uuid.UUID, userA *seededUser, userB *seededUser) error {
	for _, step := range []struct {
		token  string
		status dealtypes.DealStatus
		label  string
	}{
		{token: userA.Token, status: dealtypes.Confirmed, label: userA.Key + " confirm"},
		{token: userB.Token, status: dealtypes.Confirmed, label: userB.Key + " confirm"},
		{token: userA.Token, status: dealtypes.Completed, label: userA.Key + " complete"},
		{token: userB.Token, status: dealtypes.Completed, label: userB.Key + " complete"},
	} {
		if err := c.changeDealStatus(ctx, step.token, dealID, step.status); err != nil {
			return fmt.Errorf("%s: %w", step.label, err)
		}
	}

	return nil
}

func (c *SeedClient) createDealItemReview(ctx context.Context, token string, dealID uuid.UUID, itemID uuid.UUID, req dealtypes.CreateReviewRequest) (dealtypes.Review, error) {
	path := fmt.Sprintf("/deals/%s/items/%s/reviews", dealID, itemID)
	var body dealtypes.Review
	if err := c.doJSON(ctx, http.MethodPost, path, token, req, &body, http.StatusCreated); err != nil {
		return dealtypes.Review{}, err
	}

	return body, nil
}

func (c *SeedClient) getDealItemReviewEligibility(ctx context.Context, token string, dealID uuid.UUID, itemID uuid.UUID) (dealtypes.ReviewEligibility, error) {
	var body dealtypes.ReviewEligibility
	path := fmt.Sprintf("/deals/%s/items/%s/reviews/eligibility", dealID, itemID)
	if err := c.doJSON(ctx, http.MethodGet, path, token, nil, &body, http.StatusOK); err != nil {
		return dealtypes.ReviewEligibility{}, err
	}

	return body, nil
}

func (c *SeedClient) listDealItemReviews(ctx context.Context, token string, dealID uuid.UUID, itemID uuid.UUID) (dealtypes.GetItemReviewsResponse, error) {
	var body dealtypes.GetItemReviewsResponse
	path := fmt.Sprintf("/deals/%s/items/%s/reviews", dealID, itemID)
	if err := c.doJSON(ctx, http.MethodGet, path, token, nil, &body, http.StatusOK); err != nil {
		return nil, err
	}

	return body, nil
}

func (c *SeedClient) listDealReviews(ctx context.Context, token string, dealID uuid.UUID) (dealtypes.GetDealReviewsResponse, error) {
	var body dealtypes.GetDealReviewsResponse
	path := fmt.Sprintf("/deals/%s/reviews", dealID)
	if err := c.doJSON(ctx, http.MethodGet, path, token, nil, &body, http.StatusOK); err != nil {
		return nil, err
	}

	return body, nil
}

func (c *SeedClient) listPendingDealReviews(ctx context.Context, token string, dealID uuid.UUID) (dealtypes.GetPendingDealReviewsResponse, error) {
	var body dealtypes.GetPendingDealReviewsResponse
	path := fmt.Sprintf("/deals/%s/reviews-pending", dealID)
	if err := c.doJSON(ctx, http.MethodGet, path, token, nil, &body, http.StatusOK); err != nil {
		return nil, err
	}

	return body, nil
}

func (c *SeedClient) listOfferReviews(ctx context.Context, token string, offerID uuid.UUID) (dealtypes.GetOfferReviewsResponse, error) {
	var body dealtypes.GetOfferReviewsResponse
	path := fmt.Sprintf("/offers/%s/reviews", offerID)
	if err := c.doJSON(ctx, http.MethodGet, path, token, nil, &body, http.StatusOK); err != nil {
		return nil, err
	}

	return body, nil
}

func (c *SeedClient) getOfferReviewsSummary(ctx context.Context, token string, offerID uuid.UUID) (dealtypes.ReviewSummary, error) {
	var body dealtypes.ReviewSummary
	path := fmt.Sprintf("/offers/%s/reviews-summary", offerID)
	if err := c.doJSON(ctx, http.MethodGet, path, token, nil, &body, http.StatusOK); err != nil {
		return dealtypes.ReviewSummary{}, err
	}

	return body, nil
}

func (c *SeedClient) listProviderReviews(ctx context.Context, token string, providerID uuid.UUID) (dealtypes.GetProviderReviewsResponse, error) {
	var body dealtypes.GetProviderReviewsResponse
	path := fmt.Sprintf("/providers/%s/reviews", providerID)
	if err := c.doJSON(ctx, http.MethodGet, path, token, nil, &body, http.StatusOK); err != nil {
		return nil, err
	}

	return body, nil
}

func (c *SeedClient) getProviderReviewsSummary(ctx context.Context, token string, providerID uuid.UUID) (dealtypes.ReviewSummary, error) {
	var body dealtypes.ReviewSummary
	path := fmt.Sprintf("/providers/%s/reviews-summary", providerID)
	if err := c.doJSON(ctx, http.MethodGet, path, token, nil, &body, http.StatusOK); err != nil {
		return dealtypes.ReviewSummary{}, err
	}

	return body, nil
}

func (c *SeedClient) listAuthorReviews(ctx context.Context, token string, authorID uuid.UUID) (dealtypes.GetAuthorReviewsResponse, error) {
	var body dealtypes.GetAuthorReviewsResponse
	path := fmt.Sprintf("/authors/%s/reviews", authorID)
	if err := c.doJSON(ctx, http.MethodGet, path, token, nil, &body, http.StatusOK); err != nil {
		return nil, err
	}

	return body, nil
}

func (c *SeedClient) getReviewByID(ctx context.Context, token string, reviewID uuid.UUID) (dealtypes.Review, error) {
	var body dealtypes.Review
	path := fmt.Sprintf("/reviews/%s", reviewID)
	if err := c.doJSON(ctx, http.MethodGet, path, token, nil, &body, http.StatusOK); err != nil {
		return dealtypes.Review{}, err
	}

	return body, nil
}

func (c *SeedClient) updateReview(ctx context.Context, token string, reviewID uuid.UUID, req dealtypes.UpdateReviewRequest) (dealtypes.Review, error) {
	var body dealtypes.Review
	path := fmt.Sprintf("/reviews/%s", reviewID)
	if err := c.doJSON(ctx, http.MethodPatch, path, token, req, &body, http.StatusOK); err != nil {
		return dealtypes.Review{}, err
	}

	return body, nil
}

func (c *SeedClient) deleteReview(ctx context.Context, token string, reviewID uuid.UUID) error {
	path := fmt.Sprintf("/reviews/%s", reviewID)
	return c.doJSON(ctx, http.MethodDelete, path, token, nil, nil, http.StatusNoContent)
}

func (c *SeedClient) subscribeToUser(ctx context.Context, token string, targetUserID uuid.UUID) error {
	return c.doJSON(ctx, http.MethodPost, "/users/subscriptions", token, usertypes.SubscribeRequest{
		TargetUserId: targetUserID,
	}, nil, http.StatusCreated, http.StatusConflict)
}

func (c *SeedClient) unsubscribeFromUser(ctx context.Context, token string, targetUserID uuid.UUID) error {
	return c.doJSON(ctx, http.MethodDelete, "/users/subscriptions", token, usertypes.SubscribeRequest{
		TargetUserId: targetUserID,
	}, nil, http.StatusNoContent)
}

func (c *SeedClient) ensureMutualSubscription(ctx context.Context, userA *seededUser, userB *seededUser) error {
	if err := c.subscribeToUser(ctx, userA.Token, userB.UserID); err != nil {
		return fmt.Errorf("subscribe %s -> %s: %w", userA.Key, userB.Key, err)
	}
	if err := c.subscribeToUser(ctx, userB.Token, userA.UserID); err != nil {
		return fmt.Errorf("subscribe %s -> %s: %w", userB.Key, userA.Key, err)
	}

	return nil
}

func (c *SeedClient) createDirectChat(ctx context.Context, token string, participantID uuid.UUID) (uuid.UUID, error) {
	var body chattypes.Chat
	if err := c.doJSON(ctx, http.MethodPost, "/chats", token, chattypes.CreateChatRequest{
		ParticipantId: participantID,
	}, &body, http.StatusCreated); err != nil {
		return uuid.Nil, err
	}

	return body.Id, nil
}

func (c *SeedClient) listChats(ctx context.Context, token string) (chattypes.ListChatsResponse, error) {
	var body chattypes.ListChatsResponse
	if err := c.doJSON(ctx, http.MethodGet, "/chats", token, nil, &body, http.StatusOK); err != nil {
		return nil, err
	}

	return body, nil
}

func (c *SeedClient) listChatUsers(ctx context.Context, token string) (chattypes.ListUsersResponse, error) {
	var body chattypes.ListUsersResponse
	if err := c.doJSON(ctx, http.MethodGet, "/chats/users", token, nil, &body, http.StatusOK); err != nil {
		return nil, err
	}

	return body, nil
}

func (c *SeedClient) waitForDealChat(ctx context.Context, token string, dealID uuid.UUID) (uuid.UUID, error) {
	var chatID uuid.UUID
	err := c.poll(ctx, func(ctx context.Context) (bool, error) {
		resp, err := c.do(ctx, http.MethodGet, "/chats/deals/"+dealID.String(), token, nil)
		if err != nil {
			return false, err
		}
		defer closeBody(resp.Body)

		if resp.StatusCode == http.StatusNotFound {
			return false, nil
		}
		if resp.StatusCode != http.StatusOK {
			return false, responseError(resp, http.StatusOK, http.StatusNotFound)
		}

		var body chattypes.Chat
		if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
			return false, fmt.Errorf("decode deal chat: %w", err)
		}

		chatID = body.Id
		return true, nil
	})
	if err != nil {
		return uuid.Nil, err
	}

	return chatID, nil
}

func (c *SeedClient) sendChatMessages(ctx context.Context, chatID uuid.UUID, messages []chatMessage) error {
	for _, message := range messages {
		path := fmt.Sprintf("/chats/%s/messages", chatID)
		if err := c.doJSON(ctx, http.MethodPost, path, message.Token, chattypes.SendMessageRequest{
			Content: message.Content,
		}, nil, http.StatusCreated); err != nil {
			return err
		}
	}

	return nil
}

func (c *SeedClient) getChatMessages(ctx context.Context, token string, chatID uuid.UUID, after *time.Time) (chattypes.GetMessagesResponse, error) {
	query := url.Values{}
	if after != nil {
		query.Set("after", after.Format(time.RFC3339Nano))
	}

	var body chattypes.GetMessagesResponse
	path := pathWithQuery(fmt.Sprintf("/chats/%s/messages", chatID), query)
	if err := c.doJSON(ctx, http.MethodGet, path, token, nil, &body, http.StatusOK); err != nil {
		return nil, err
	}

	return body, nil
}

func (c *SeedClient) createAndCancelDeal(ctx context.Context, userA, userB *seededUser, offerA, offerB uuid.UUID, name, description string) (uuid.UUID, error) {
	dealID, err := c.createTwoPartyDeal(ctx, userA, userB, offerA, offerB, name, description)
	if err != nil {
		return uuid.Nil, err
	}

	if err := c.changeDealStatus(ctx, userA.Token, dealID, dealtypes.Cancelled); err != nil {
		return uuid.Nil, fmt.Errorf("cancel deal by %s: %w", userA.Key, err)
	}
	// second vote may fail if already cancelled — tolerate it
	_ = c.changeDealStatus(ctx, userB.Token, dealID, dealtypes.Cancelled)

	return dealID, nil
}

func (c *SeedClient) createOfferReport(ctx context.Context, token string, offerID uuid.UUID, message string) (dealtypes.OfferReport, error) {
	var body dealtypes.OfferReport
	path := fmt.Sprintf("/offers/%s/reports", offerID)
	if err := c.doJSON(ctx, http.MethodPost, path, token, dealtypes.CreateOfferReportRequest{
		Message: message,
	}, &body, http.StatusCreated, http.StatusOK); err != nil {
		return dealtypes.OfferReport{}, err
	}

	return body, nil
}

func (c *SeedClient) resolveOfferReport(ctx context.Context, adminToken string, reportID uuid.UUID, accepted bool, comment *string) (dealtypes.OfferReport, error) {
	var body dealtypes.OfferReport
	path := fmt.Sprintf("/admin/offer-reports/%s/resolution", reportID)
	if err := c.doJSON(ctx, http.MethodPost, path, adminToken, dealtypes.ResolveOfferReportRequest{
		Accepted: accepted,
		Comment:  comment,
	}, &body, http.StatusOK, http.StatusConflict); err != nil {
		return dealtypes.OfferReport{}, err
	}

	return body, nil
}

func (c *SeedClient) voteForFailure(ctx context.Context, token string, dealID uuid.UUID, accusedUserID uuid.UUID) error {
	path := fmt.Sprintf("/deals/failures/%s/votes", dealID)
	err := c.doJSON(ctx, http.MethodPost, path, token, dealtypes.VoteForFailureRequest{
		UserId: accusedUserID,
	}, nil, http.StatusNoContent, http.StatusForbidden)
	return err
}

func (c *SeedClient) moderatorResolutionForFailure(ctx context.Context, adminToken string, dealID uuid.UUID, req dealtypes.ModeratorResolutionForFailureRequest) error {
	path := fmt.Sprintf("/deals/failures/%s/moderator-resolution", dealID)
	return c.doJSON(ctx, http.MethodPost, path, adminToken, req, nil, http.StatusOK, http.StatusConflict)
}

func (c *SeedClient) requestJoinDeal(ctx context.Context, token string, dealID uuid.UUID) error {
	path := fmt.Sprintf("/deals/%s/joins", dealID)
	return c.doJSON(ctx, http.MethodPost, path, token, nil, nil, http.StatusNoContent, http.StatusForbidden, http.StatusNotFound)
}

func (c *SeedClient) getDealJoinRequests(ctx context.Context, token string, dealID uuid.UUID) (dealtypes.GetDealJoinRequestsResponse, error) {
	var body dealtypes.GetDealJoinRequestsResponse
	path := fmt.Sprintf("/deals/%s/joins", dealID)
	if err := c.doJSON(ctx, http.MethodGet, path, token, nil, &body, http.StatusOK); err != nil {
		return nil, err
	}

	return body, nil
}

func (c *SeedClient) leaveDeal(ctx context.Context, token string, dealID uuid.UUID) error {
	path := fmt.Sprintf("/deals/%s/joins", dealID)
	return c.doJSON(ctx, http.MethodDelete, path, token, nil, nil, http.StatusNoContent)
}

func (c *SeedClient) processJoinRequest(ctx context.Context, token string, dealID uuid.UUID, applicantUserID uuid.UUID, accept bool) error {
	path := fmt.Sprintf("/deals/%s/joins/%s?accept=%v", dealID, applicantUserID, accept)
	resp, err := c.do(ctx, http.MethodPost, path, token, nil)
	if err != nil {
		return err
	}
	defer closeBody(resp.Body)

	switch resp.StatusCode {
	case http.StatusNoContent, http.StatusForbidden, http.StatusNotFound:
		return nil
	default:
		return responseError(resp, http.StatusNoContent)
	}
}

func (c *SeedClient) revokeVoteForFailure(ctx context.Context, token string, dealID uuid.UUID) error {
	path := fmt.Sprintf("/deals/failures/%s/votes", dealID)
	return c.doJSON(ctx, http.MethodDelete, path, token, nil, nil, http.StatusNoContent)
}

func (c *SeedClient) getFailureVotes(ctx context.Context, token string, dealID uuid.UUID) (dealtypes.FailureVotesResponse, error) {
	var body dealtypes.FailureVotesResponse
	path := fmt.Sprintf("/deals/failures/%s/votes", dealID)
	if err := c.doJSON(ctx, http.MethodGet, path, token, nil, &body, http.StatusOK); err != nil {
		return nil, err
	}

	return body, nil
}

func (c *SeedClient) listDealsForFailureReview(ctx context.Context, token string) (dealtypes.FailureModerationDealsResponse, error) {
	var body dealtypes.FailureModerationDealsResponse
	if err := c.doJSON(ctx, http.MethodGet, "/deals/failures/review", token, nil, &body, http.StatusOK); err != nil {
		return nil, err
	}

	return body, nil
}

func (c *SeedClient) getFailureMaterials(ctx context.Context, token string, dealID uuid.UUID) (dealtypes.FailureMaterialResponse, error) {
	var body dealtypes.FailureMaterialResponse
	path := fmt.Sprintf("/deals/failures/%s/materials", dealID)
	if err := c.doJSON(ctx, http.MethodGet, path, token, nil, &body, http.StatusOK); err != nil {
		return dealtypes.FailureMaterialResponse{}, err
	}

	return body, nil
}

func (c *SeedClient) getModeratorResolutionForFailure(ctx context.Context, token string, dealID uuid.UUID) (dealtypes.DealFailureModeratorResolution, error) {
	var body dealtypes.DealFailureModeratorResolution
	path := fmt.Sprintf("/deals/failures/%s/moderator-resolution", dealID)
	if err := c.doJSON(ctx, http.MethodGet, path, token, nil, &body, http.StatusOK); err != nil {
		return dealtypes.DealFailureModeratorResolution{}, err
	}

	return body, nil
}

func (c *SeedClient) getMyStatistics(ctx context.Context, token string) (dealtypes.MyStatistics, error) {
	var body dealtypes.MyStatistics
	if err := c.doJSON(ctx, http.MethodGet, "/me/statistics", token, nil, &body, http.StatusOK); err != nil {
		return dealtypes.MyStatistics{}, err
	}

	return body, nil
}

func (c *SeedClient) poll(ctx context.Context, fn func(context.Context) (bool, error)) error {
	ticker := time.NewTicker(c.PollInterval)
	defer ticker.Stop()

	for {
		done, err := fn(ctx)
		if err != nil {
			return err
		}
		if done {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

func (c *SeedClient) doJSON(
	ctx context.Context,
	method string,
	path string,
	token string,
	reqBody any,
	respBody any,
	expectedStatuses ...int,
) error {
	resp, err := c.do(ctx, method, path, token, reqBody)
	if err != nil {
		return err
	}
	defer closeBody(resp.Body)

	if !containsStatus(expectedStatuses, resp.StatusCode) {
		return responseError(resp, expectedStatuses...)
	}

	if respBody == nil {
		return nil
	}

	if err := json.NewDecoder(resp.Body).Decode(respBody); err != nil {
		return fmt.Errorf("decode response %s %s: %w", method, path, err)
	}

	return nil
}

func (c *SeedClient) do(ctx context.Context, method string, path string, token string, reqBody any) (*http.Response, error) {
	var body io.Reader
	if reqBody != nil {
		data, err := json.Marshal(reqBody)
		if err != nil {
			return nil, fmt.Errorf("marshal request %s %s: %w", method, path, err)
		}
		body = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.BaseURL+path, body)
	if err != nil {
		return nil, fmt.Errorf("build request %s %s: %w", method, path, err)
	}

	if reqBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("perform request %s %s: %w", method, path, err)
	}

	return resp, nil
}

func pathWithQuery(path string, query url.Values) string {
	if len(query) == 0 {
		return path
	}

	return path + "?" + query.Encode()
}

func responseError(resp *http.Response, expectedStatuses ...int) error {
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	message := strings.TrimSpace(string(body))
	if message == "" {
		message = http.StatusText(resp.StatusCode)
	}

	var errBody struct {
		Message *string `json:"message"`
	}
	if json.Unmarshal(body, &errBody) == nil && errBody.Message != nil && *errBody.Message != "" {
		message = *errBody.Message
	}

	return &httpStatusError{
		StatusCode:       resp.StatusCode,
		ExpectedStatuses: append([]int(nil), expectedStatuses...),
		Message:          message,
	}
}

type httpStatusError struct {
	StatusCode       int
	ExpectedStatuses []int
	Message          string
}

func (e *httpStatusError) Error() string {
	return fmt.Sprintf("unexpected status %d, expected %v: %s", e.StatusCode, e.ExpectedStatuses, e.Message)
}

func isMediaUploadFallbackable(err error) bool {
	var statusErr *httpStatusError
	return errors.As(err, &statusErr) && statusErr.StatusCode >= http.StatusInternalServerError
}

func retryMediaUpload[T any](ctx context.Context, delay time.Duration, attempts int, fn func() (T, error)) (T, error) {
	var zero T
	if attempts < 1 {
		attempts = 1
	}
	if delay <= 0 {
		delay = 500 * time.Millisecond
	}

	var lastErr error
	for attempt := 1; attempt <= attempts; attempt++ {
		result, err := fn()
		if err == nil {
			return result, nil
		}
		lastErr = err
		if !isMediaUploadFallbackable(err) || attempt == attempts {
			break
		}

		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return zero, ctx.Err()
		case <-timer.C:
		}
	}

	return zero, lastErr
}
