package main

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
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

func (c *seedClient) register(ctx context.Context, email, password string) (registerResponse, error) {
	var respBody registerResponse
	if err := c.doJSON(ctx, http.MethodPost, "/auth/register", "", registerRequest{
		Email:    email,
		Password: password,
	}, &respBody, http.StatusOK); err != nil {
		return registerResponse{}, err
	}

	return respBody, nil
}

func (c *seedClient) ensureUser(ctx context.Context, email, password string) (registerResponse, string, error) {
	registered, err := c.register(ctx, email, password)
	if err == nil {
		if err := c.waitForAuthProvisioning(ctx, registered.UserID); err != nil {
			return registerResponse{}, "", err
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
		UserID: uuid.UUID(me.Id),
		Email:  me.Email,
	}, token, nil
}

func (c *seedClient) waitForAuthProvisioning(ctx context.Context, userID uuid.UUID) error {
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

func (c *seedClient) login(ctx context.Context, email, password string) (string, error) {
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

func (c *seedClient) waitForUsersProjection(ctx context.Context, token string) error {
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

func (c *seedClient) updateMe(ctx context.Context, token string, req usertypes.UpdateUserRequest) (usertypes.Me, error) {
	var body usertypes.Me
	if err := c.doJSON(ctx, http.MethodPatch, "/users/me", token, req, &body, http.StatusOK); err != nil {
		return usertypes.Me{}, err
	}

	return body, nil
}

func (c *seedClient) getMe(ctx context.Context, token string) (usertypes.Me, error) {
	var body usertypes.Me
	if err := c.doJSON(ctx, http.MethodGet, "/users/me", token, nil, &body, http.StatusOK); err != nil {
		return usertypes.Me{}, err
	}

	return body, nil
}

func (c *seedClient) createOffers(ctx context.Context, user *seededUser, specs []offerSpec) (map[string]uuid.UUID, error) {
	result := make(map[string]uuid.UUID, len(specs))
	for _, spec := range specs {
		var offer dealtypes.Offer
		if err := c.doJSON(ctx, http.MethodPost, "/offers", user.Token, dealtypes.CreateOfferRequest{
			Name:        spec.Name,
			Description: spec.Description,
			Type:        spec.Type,
			Action:      spec.Action,
		}, &offer, http.StatusCreated); err != nil {
			return nil, fmt.Errorf("create offer %s for %s: %w", spec.Key, user.Key, err)
		}

		result[spec.Key] = uuid.UUID(offer.Id)
	}

	return result, nil
}

func (c *seedClient) createOfferGroup(ctx context.Context, token string, req offerGroupRequest) (uuid.UUID, error) {
	var body offerGroupResponse
	if err := c.doJSON(ctx, http.MethodPost, "/offer-groups", token, req, &body, http.StatusCreated); err != nil {
		return uuid.Nil, err
	}

	return body.ID, nil
}

func (c *seedClient) createDraftFromOfferGroup(ctx context.Context, token string, offerGroupID uuid.UUID, req offerGroupDraftRequest) (uuid.UUID, error) {
	var body dealtypes.CreateDraftDealResponse
	path := fmt.Sprintf("/offer-groups/%s/drafts", offerGroupID)
	if err := c.doJSON(ctx, http.MethodPost, path, token, req, &body, http.StatusCreated); err != nil {
		return uuid.Nil, err
	}

	return uuid.UUID(body.Id), nil
}

func (c *seedClient) listMyDeals(ctx context.Context, token string) (dealtypes.GetDealsResponse, error) {
	var deals dealtypes.GetDealsResponse
	if err := c.doJSON(ctx, http.MethodGet, "/deals?my=true", token, nil, &deals, http.StatusOK); err != nil {
		return nil, err
	}

	return deals, nil
}

func (c *seedClient) createDraft(ctx context.Context, token string, req dealtypes.CreateDraftDealRequest) (uuid.UUID, error) {
	var body dealtypes.CreateDraftDealResponse
	if err := c.doJSON(ctx, http.MethodPost, "/deals/drafts", token, req, &body, http.StatusCreated); err != nil {
		return uuid.Nil, err
	}

	return uuid.UUID(body.Id), nil
}

func (c *seedClient) confirmDraft(ctx context.Context, token string, draftID uuid.UUID) error {
	return c.doJSON(ctx, http.MethodPatch, "/deals/drafts/"+draftID.String(), token, nil, nil, http.StatusOK)
}

func (c *seedClient) createTwoPartyDeal(
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

func (c *seedClient) waitForNewDeal(ctx context.Context, token string, before dealtypes.GetDealsResponse) (uuid.UUID, error) {
	beforeSet := make(map[uuid.UUID]struct{}, len(before))
	for _, deal := range before {
		beforeSet[uuid.UUID(deal.Id)] = struct{}{}
	}

	var created uuid.UUID
	err := c.poll(ctx, func(ctx context.Context) (bool, error) {
		after, err := c.listMyDeals(ctx, token)
		if err != nil {
			return false, err
		}

		for _, deal := range after {
			id := uuid.UUID(deal.Id)
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

func (c *seedClient) getDealByID(ctx context.Context, token string, dealID uuid.UUID) (dealtypes.Deal, error) {
	var body dealtypes.Deal
	if err := c.doJSON(ctx, http.MethodGet, "/deals/"+dealID.String(), token, nil, &body, http.StatusOK); err != nil {
		return dealtypes.Deal{}, err
	}

	return body, nil
}

func (c *seedClient) updateDealItem(ctx context.Context, token string, dealID uuid.UUID, itemID uuid.UUID, req dealtypes.UpdateDealItemRequest) error {
	path := fmt.Sprintf("/deals/%s/items/%s", dealID, itemID)
	return c.doJSON(ctx, http.MethodPatch, path, token, req, nil, http.StatusOK)
}

func (c *seedClient) changeDealStatus(ctx context.Context, token string, dealID uuid.UUID, status dealtypes.DealStatus) error {
	return c.doJSON(ctx, http.MethodPatch, "/deals/"+dealID.String()+"/status", token, dealtypes.ChangeDealStatusRequest{
		ExpectedStatus: status,
	}, nil, http.StatusOK)
}

func (c *seedClient) promoteDealToDiscussion(ctx context.Context, dealID uuid.UUID, userA *seededUser, userB *seededUser) (dealtypes.Deal, error) {
	deal, err := c.getDealByID(ctx, userA.Token, dealID)
	if err != nil {
		return dealtypes.Deal{}, err
	}

	itemByAuthor := make(map[uuid.UUID]uuid.UUID, len(deal.Items))
	for _, item := range deal.Items {
		itemByAuthor[uuid.UUID(item.AuthorId)] = uuid.UUID(item.Id)
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

func (c *seedClient) completeTwoPartyDeal(ctx context.Context, dealID uuid.UUID, userA *seededUser, userB *seededUser) error {
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

func (c *seedClient) createMutualReviews(ctx context.Context, dealID uuid.UUID, deal dealtypes.Deal, userA *seededUser, userB *seededUser) error {
	itemByAuthor := make(map[uuid.UUID]uuid.UUID, len(deal.Items))
	for _, item := range deal.Items {
		itemByAuthor[uuid.UUID(item.AuthorId)] = uuid.UUID(item.Id)
	}

	itemA, ok := itemByAuthor[userA.UserID]
	if !ok {
		return fmt.Errorf("completed deal %s does not contain item for %s", dealID, userA.Key)
	}
	itemB, ok := itemByAuthor[userB.UserID]
	if !ok {
		return fmt.Errorf("completed deal %s does not contain item for %s", dealID, userB.Key)
	}

	commentA := "Все прошло четко: договорились быстро и получили именно то, что ожидали."
	if err := c.createDealItemReview(ctx, userB.Token, dealID, itemA, dealtypes.CreateReviewRequest{
		Rating:  5,
		Comment: &commentA,
	}); err != nil {
		return fmt.Errorf("review for %s item: %w", userA.Key, err)
	}

	commentB := "Хорошая коммуникация и удобная передача вещи."
	if err := c.createDealItemReview(ctx, userA.Token, dealID, itemB, dealtypes.CreateReviewRequest{
		Rating:  5,
		Comment: &commentB,
	}); err != nil {
		return fmt.Errorf("review for %s item: %w", userB.Key, err)
	}

	return nil
}

func (c *seedClient) createDealItemReview(ctx context.Context, token string, dealID uuid.UUID, itemID uuid.UUID, req dealtypes.CreateReviewRequest) error {
	path := fmt.Sprintf("/deals/%s/items/%s/reviews", dealID, itemID)
	return c.doJSON(ctx, http.MethodPost, path, token, req, nil, http.StatusCreated)
}

func (c *seedClient) subscribeToUser(ctx context.Context, token string, targetUserID uuid.UUID) error {
	return c.doJSON(ctx, http.MethodPost, "/users/subscriptions", token, usertypes.SubscribeRequest{
		TargetUserId: targetUserID,
	}, nil, http.StatusCreated, http.StatusConflict)
}

func (c *seedClient) ensureMutualSubscription(ctx context.Context, userA *seededUser, userB *seededUser) error {
	if err := c.subscribeToUser(ctx, userA.Token, userB.UserID); err != nil {
		return fmt.Errorf("subscribe %s -> %s: %w", userA.Key, userB.Key, err)
	}
	if err := c.subscribeToUser(ctx, userB.Token, userA.UserID); err != nil {
		return fmt.Errorf("subscribe %s -> %s: %w", userB.Key, userA.Key, err)
	}

	return nil
}

func (c *seedClient) createDirectChat(ctx context.Context, token string, participantID uuid.UUID) (uuid.UUID, error) {
	var body chattypes.Chat
	if err := c.doJSON(ctx, http.MethodPost, "/chats", token, chattypes.CreateChatRequest{
		ParticipantId: participantID,
	}, &body, http.StatusCreated); err != nil {
		return uuid.Nil, err
	}

	return uuid.UUID(body.Id), nil
}

func (c *seedClient) waitForDealChat(ctx context.Context, token string, dealID uuid.UUID) (uuid.UUID, error) {
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

		chatID = uuid.UUID(body.Id)
		return true, nil
	})
	if err != nil {
		return uuid.Nil, err
	}

	return chatID, nil
}

func (c *seedClient) sendChatMessages(ctx context.Context, chatID uuid.UUID, messages []chatMessage) error {
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

func (c *seedClient) poll(ctx context.Context, fn func(context.Context) (bool, error)) error {
	ticker := time.NewTicker(c.pollInterval)
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

func (c *seedClient) doJSON(
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

func (c *seedClient) do(ctx context.Context, method string, path string, token string, reqBody any) (*http.Response, error) {
	var body io.Reader
	if reqBody != nil {
		data, err := json.Marshal(reqBody)
		if err != nil {
			return nil, fmt.Errorf("marshal request %s %s: %w", method, path, err)
		}
		body = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return nil, fmt.Errorf("build request %s %s: %w", method, path, err)
	}

	if reqBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("perform request %s %s: %w", method, path, err)
	}

	return resp, nil
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

	return fmt.Errorf("unexpected status %d, expected %v: %s", resp.StatusCode, expectedStatuses, message)
}
