package seed_demo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
)

const verificationEmailSubject = "Confirm your email"

var errSMTP4DevTemporary = errors.New("smtp4dev temporary response")

type smtp4devPagedResult struct {
	Results []smtp4devMessageSummary `json:"results"`
	Items   []smtp4devMessageSummary `json:"items"`
}

type smtp4devMessageSummary struct {
	ID           uuid.UUID          `json:"id"`
	Subject      string             `json:"subject"`
	To           smtp4devRecipients `json:"to"`
	ReceivedDate time.Time          `json:"receivedDate"`
}

type smtp4devRecipients []string

func (r *smtp4devRecipients) UnmarshalJSON(data []byte) error {
	var single string
	if err := json.Unmarshal(data, &single); err == nil {
		*r = smtp4devRecipients{single}
		return nil
	}

	var many []string
	if err := json.Unmarshal(data, &many); err == nil {
		*r = smtp4devRecipients(many)
		return nil
	}

	var rawItems []map[string]any
	if err := json.Unmarshal(data, &rawItems); err == nil {
		recipients := make([]string, 0, len(rawItems))
		for _, item := range rawItems {
			recipients = append(recipients, smtp4devRecipientStrings(item)...)
		}
		*r = smtp4devRecipients(recipients)
		return nil
	}

	return fmt.Errorf("unsupported smtp4dev recipients payload: %s", string(data))
}

func smtp4devRecipientStrings(item map[string]any) []string {
	result := make([]string, 0, 2)

	for _, key := range []string{"address", "email", "value"} {
		value, ok := item[key]
		if !ok {
			continue
		}
		if s, ok := value.(string); ok && strings.TrimSpace(s) != "" {
			result = append(result, s)
		}
	}

	if len(result) > 0 {
		return result
	}

	for _, value := range item {
		if s, ok := value.(string); ok && strings.TrimSpace(s) != "" {
			result = append(result, s)
		}
	}

	return result
}

func (c *SeedClient) smtp4devConfigured() bool {
	return strings.TrimSpace(c.SMTP4DevURL) != ""
}

func (c *SeedClient) smtp4devRequest(ctx context.Context, method, path string, query url.Values) (*http.Response, error) {
	if !c.smtp4devConfigured() {
		return nil, errors.New("smtp4dev URL is not configured")
	}

	base := strings.TrimRight(c.SMTP4DevURL, "/")
	if query != nil && len(query) > 0 {
		path += "?" + query.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, method, base+path, nil)
	if err != nil {
		return nil, fmt.Errorf("build smtp4dev request %s %s: %w", method, path, err)
	}
	if c.SMTP4DevUser != "" {
		req.SetBasicAuth(c.SMTP4DevUser, c.SMTP4DevPass)
	}

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("perform smtp4dev request %s %s: %w", method, path, err)
	}

	return resp, nil
}

func (c *SeedClient) getLatestSMTP4DevMessageID(ctx context.Context) (uuid.UUID, error) {
	query := url.Values{}
	query.Set("page", "1")
	query.Set("pageSize", "1")
	path := "/api/messages"

	resp, err := c.smtp4devRequest(ctx, http.MethodGet, path, query)
	if err != nil {
		return uuid.Nil, err
	}
	defer closeBody(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return uuid.Nil, smtp4devResponseError(resp, http.MethodGet, path, http.StatusOK)
	}

	var page smtp4devPagedResult
	if err := json.NewDecoder(resp.Body).Decode(&page); err != nil {
		return uuid.Nil, fmt.Errorf("decode smtp4dev messages page: %w", err)
	}

	messages := page.Results
	if len(messages) == 0 {
		messages = page.Items
	}
	if len(messages) == 0 {
		return uuid.Nil, nil
	}

	return messages[0].ID, nil
}

func (c *SeedClient) listNewSMTP4DevMessages(ctx context.Context, lastSeenID uuid.UUID) ([]smtp4devMessageSummary, error) {
	query := url.Values{}
	query.Set("pageSize", "50")
	path := "/api/messages/new"
	if lastSeenID != uuid.Nil {
		query.Set("lastSeenMessageId", lastSeenID.String())
	}

	resp, err := c.smtp4devRequest(ctx, http.MethodGet, path, query)
	if err != nil {
		return nil, err
	}
	defer closeBody(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, smtp4devResponseError(resp, http.MethodGet, path, http.StatusOK)
	}

	var messages []smtp4devMessageSummary
	if err := json.NewDecoder(resp.Body).Decode(&messages); err != nil {
		return nil, fmt.Errorf("decode smtp4dev new messages: %w", err)
	}

	return messages, nil
}

func (c *SeedClient) getSMTP4DevPlaintext(ctx context.Context, messageID uuid.UUID) (string, error) {
	path := "/api/messages/" + messageID.String() + "/plaintext"
	resp, err := c.smtp4devRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return "", err
	}
	defer closeBody(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return "", smtp4devResponseError(resp, http.MethodGet, path, http.StatusOK)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read smtp4dev plaintext body: %w", err)
	}

	return string(body), nil
}

func (c *SeedClient) waitForVerificationTokenFromEmail(ctx context.Context, email string, lastSeenID uuid.UUID) (string, error) {
	var token string
	err := c.poll(ctx, func(ctx context.Context) (bool, error) {
		messages, err := c.listNewSMTP4DevMessages(ctx, lastSeenID)
		if err != nil {
			if errors.Is(err, errSMTP4DevTemporary) {
				return false, nil
			}
			return false, err
		}

		sort.Slice(messages, func(i, j int) bool {
			return messages[i].ReceivedDate.After(messages[j].ReceivedDate)
		})

		for _, message := range messages {
			if !strings.EqualFold(strings.TrimSpace(message.Subject), verificationEmailSubject) {
				continue
			}
			if !message.To.contains(email) {
				continue
			}

			body, err := c.getSMTP4DevPlaintext(ctx, message.ID)
			if err != nil {
				if errors.Is(err, errSMTP4DevTemporary) {
					return false, nil
				}
				return false, err
			}

			token, err = extractVerificationToken(body)
			if err != nil {
				return false, nil
			}

			return true, nil
		}

		return false, nil
	})
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return "", fmt.Errorf("timed out waiting for verification email for %s via smtp4dev %s", email, c.SMTP4DevURL)
		}
		return "", err
	}

	return token, nil
}

func (c *SeedClient) waitForLatestSMTP4DevMessageID(ctx context.Context) (uuid.UUID, error) {
	var messageID uuid.UUID
	err := c.poll(ctx, func(ctx context.Context) (bool, error) {
		id, err := c.getLatestSMTP4DevMessageID(ctx)
		if err != nil {
			if errors.Is(err, errSMTP4DevTemporary) {
				return false, nil
			}
			return false, err
		}

		messageID = id
		return true, nil
	})
	if err != nil {
		return uuid.Nil, err
	}

	return messageID, nil
}

func extractVerificationToken(body string) (string, error) {
	for _, field := range strings.Fields(body) {
		if !strings.Contains(field, "token=") {
			continue
		}

		u, err := url.Parse(strings.TrimSpace(field))
		if err != nil {
			continue
		}

		token := strings.TrimSpace(u.Query().Get("token"))
		if token != "" {
			return token, nil
		}
	}

	return "", errors.New("verification token not found in email body")
}

func (r smtp4devRecipients) contains(email string) bool {
	needle := strings.ToLower(strings.TrimSpace(email))
	for _, recipient := range r {
		if strings.Contains(strings.ToLower(recipient), needle) {
			return true
		}
	}

	return false
}

func smtp4devResponseError(resp *http.Response, method, path string, expectedStatuses ...int) error {
	if isTemporarySMTP4DevStatus(resp.StatusCode) {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		message := strings.TrimSpace(string(body))
		if message == "" {
			message = http.StatusText(resp.StatusCode)
		}
		return fmt.Errorf("%w: %s %s returned %d: %s", errSMTP4DevTemporary, method, path, resp.StatusCode, message)
	}

	return fmt.Errorf("%s %s: %w", method, path, responseError(resp, expectedStatuses...))
}

func isTemporarySMTP4DevStatus(status int) bool {
	switch status {
	case http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return true
	default:
		return false
	}
}
