package main

import (
	"barter-port/contracts/openapi/chats/types"
	dealtypes "barter-port/contracts/openapi/deals/types"
	usertypes "barter-port/contracts/openapi/users/types"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand/v2"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

const (
	defaultBaseURL       = "http://localhost:80"
	defaultPassword      = "password123"
	defaultAvatarBaseURL = "http://localhost:8333/avatars"
	defaultTimeout       = 2 * time.Minute
	defaultPollInterval  = 500 * time.Millisecond
)

type seedConfig struct {
	BaseURL       string
	Password      string
	AvatarBaseURL string
	Timeout       time.Duration
	PollInterval  time.Duration
}

type seedClient struct {
	baseURL      string
	httpClient   *http.Client
	pollInterval time.Duration
}

type authStatusResponse struct {
	Status string `json:"status"`
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type registerResponse struct {
	UserID uuid.UUID `json:"userId"`
	Email  string    `json:"email"`
}

type loginResponse struct {
	AccessToken string `json:"accessToken"`
}

type offerGroupRequest struct {
	Name        *string                 `json:"name,omitempty"`
	Description *string                 `json:"description,omitempty"`
	Units       []offerGroupUnitRequest `json:"units"`
}

type offerGroupUnitRequest struct {
	Offers []offerGroupOfferRef `json:"offers"`
}

type offerGroupOfferRef struct {
	OfferID uuid.UUID `json:"offerId"`
}

type offerGroupDraftRequest struct {
	SelectedOfferIDs []uuid.UUID `json:"selectedOfferIds"`
	ResponderOfferID *uuid.UUID  `json:"responderOfferId,omitempty"`
	Name             *string     `json:"name,omitempty"`
	Description      *string     `json:"description,omitempty"`
}

type offerGroupResponse struct {
	ID uuid.UUID `json:"id"`
}

type seededUser struct {
	Key      string
	Name     string
	Bio      string
	Email    string
	Password string
	Avatar   string
	UserID   uuid.UUID
	Token    string
}

type offerSpec struct {
	Key         string
	Name        string
	Description string
	Type        dealtypes.ItemType
	Action      dealtypes.OfferAction
}

type seedSummary struct {
	Users             []seededUserSummary `json:"users"`
	OfferGroupID      uuid.UUID           `json:"offerGroupId"`
	OfferGroupDraftID uuid.UUID           `json:"offerGroupDraftId"`
	DiscussionDealID  uuid.UUID           `json:"discussionDealId"`
	CompletedDealID   uuid.UUID           `json:"completedDealId"`
	DirectChatID      uuid.UUID           `json:"directChatId"`
	DealChatID        uuid.UUID           `json:"dealChatId"`
}

type seededUserSummary struct {
	Key      string `json:"key"`
	UserID   string `json:"userId"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func main() {
	_ = godotenv.Load()

	cfg, err := parseConfig()
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	client := &seedClient{
		baseURL: strings.TrimRight(cfg.BaseURL, "/"),
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
		pollInterval: cfg.PollInterval,
	}

	summary, err := runSeed(ctx, client, cfg)
	if err != nil {
		log.Fatal(err)
	}

	data, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Seed completed against %s\n%s\n", cfg.BaseURL, data)
}

func parseConfig() (seedConfig, error) {
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	baseURL := fs.String("base-url", firstNonEmpty(os.Getenv("SEED_BASE_URL"), defaultBaseURL), "HTTP base URL for the app gateway")
	password := fs.String("password", firstNonEmpty(os.Getenv("SEED_PASSWORD"), defaultPassword), "Password for seeded demo users")
	avatarBaseURL := fs.String("avatar-base-url", firstNonEmpty(os.Getenv("SEED_AVATAR_BASE_URL"), defaultAvatarBaseURL), "Base URL for demo avatars")
	timeout := fs.Duration("timeout", durationFromEnv("SEED_TIMEOUT", defaultTimeout), "Overall seed timeout")
	pollInterval := fs.Duration("poll-interval", durationFromEnv("SEED_POLL_INTERVAL", defaultPollInterval), "Polling interval for async readiness checks")

	if err := fs.Parse(os.Args[1:]); err != nil {
		return seedConfig{}, err
	}

	if strings.TrimSpace(*baseURL) == "" {
		return seedConfig{}, errors.New("base URL must not be empty")
	}

	return seedConfig{
		BaseURL:       *baseURL,
		Password:      *password,
		AvatarBaseURL: strings.TrimRight(*avatarBaseURL, "/"),
		Timeout:       *timeout,
		PollInterval:  *pollInterval,
	}, nil
}

func runSeed(ctx context.Context, client *seedClient, cfg seedConfig) (*seedSummary, error) {
	suffix := fmt.Sprintf("%s-%04d", time.Now().UTC().Format("20060102-150405"), rand.IntN(10000))

	users := []seededUser{
		{
			Key:      "alice",
			Name:     "Alice Morozova",
			Bio:      "Меняю декор для дома, книги и комнатные растения.",
			Email:    fmt.Sprintf("alice.%s@demo.local", suffix),
			Password: cfg.Password,
			Avatar:   cfg.AvatarBaseURL + "/alice.png",
		},
		{
			Key:      "bob",
			Name:     "Bob Sokolov",
			Bio:      "Обмениваю велоаксессуары, инструменты и вещи для поездок.",
			Email:    fmt.Sprintf("bob.%s@demo.local", suffix),
			Password: cfg.Password,
			Avatar:   cfg.AvatarBaseURL + "/bob.png",
		},
		{
			Key:      "clara",
			Name:     "Clara Lebedeva",
			Bio:      "Предлагаю занятия по английскому и ищу настольные игры.",
			Email:    fmt.Sprintf("clara.%s@demo.local", suffix),
			Password: cfg.Password,
			Avatar:   cfg.AvatarBaseURL + "/clara.png",
		},
		{
			Key:      "dan",
			Name:     "Dan Volkov",
			Bio:      "Люблю ремонтировать мелкую технику и обменивать хобби-товары.",
			Email:    fmt.Sprintf("dan.%s@demo.local", suffix),
			Password: cfg.Password,
			Avatar:   cfg.AvatarBaseURL + "/dan.png",
		},
	}

	for i := range users {
		registered, err := client.register(ctx, users[i].Email, users[i].Password)
		if err != nil {
			return nil, fmt.Errorf("register %s: %w", users[i].Key, err)
		}

		users[i].UserID = registered.UserID

		if err := client.waitForAuthProvisioning(ctx, registered.UserID); err != nil {
			return nil, fmt.Errorf("wait auth provisioning for %s: %w", users[i].Key, err)
		}

		token, err := client.login(ctx, users[i].Email, users[i].Password)
		if err != nil {
			return nil, fmt.Errorf("login %s: %w", users[i].Key, err)
		}
		users[i].Token = token

		if err := client.waitForUsersProjection(ctx, token); err != nil {
			return nil, fmt.Errorf("wait users projection for %s: %w", users[i].Key, err)
		}

		if _, err := client.updateMe(ctx, token, usertypes.UpdateUserRequest{
			Name:      stringPtr(users[i].Name),
			Bio:       stringPtr(users[i].Bio),
			AvatarUrl: stringPtr(users[i].Avatar),
		}); err != nil {
			return nil, fmt.Errorf("update profile for %s: %w", users[i].Key, err)
		}
	}

	alice := &users[0]
	bob := &users[1]
	clara := &users[2]
	dan := &users[3]

	aliceOffers, err := client.createOffers(ctx, alice, []offerSpec{
		{
			Key:         "lamp",
			Name:        "Винтажная лампа",
			Description: "Латунная настольная лампа в рабочем состоянии.",
			Type:        dealtypes.Good,
			Action:      dealtypes.Give,
		},
		{
			Key:         "plant-consulting",
			Name:        "Консультация по комнатным растениям",
			Description: "Помогу подобрать уход и пересадку для домашних растений.",
			Type:        dealtypes.Service,
			Action:      dealtypes.Give,
		},
		{
			Key:         "storage-boxes",
			Name:        "Набор коробов для хранения",
			Description: "Три тканевых короба для стеллажа и хранения мелочей.",
			Type:        dealtypes.Good,
			Action:      dealtypes.Give,
		},
		{
			Key:         "bookshelf",
			Name:        "Ищу узкий книжный стеллаж",
			Description: "Нужен компактный стеллаж для прихожей.",
			Type:        dealtypes.Good,
			Action:      dealtypes.Take,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("create Alice offers: %w", err)
	}

	bobOffers, err := client.createOffers(ctx, bob, []offerSpec{
		{
			Key:         "bike-bag",
			Name:        "Велосипедная сумка",
			Description: "Влагозащищенная сумка на багажник.",
			Type:        dealtypes.Good,
			Action:      dealtypes.Give,
		},
		{
			Key:         "tool-kit",
			Name:        "Набор инструментов",
			Description: "Компактный набор ключей и отверток для дома.",
			Type:        dealtypes.Good,
			Action:      dealtypes.Give,
		},
		{
			Key:         "coffee-beans",
			Name:        "Ищу зерновой кофе",
			Description: "Обменяю на аксессуары для поездок или инструменты.",
			Type:        dealtypes.Good,
			Action:      dealtypes.Take,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("create Bob offers: %w", err)
	}

	claraOffers, err := client.createOffers(ctx, clara, []offerSpec{
		{
			Key:         "english-session",
			Name:        "Разговорный английский",
			Description: "Час разговорной практики онлайн или офлайн.",
			Type:        dealtypes.Service,
			Action:      dealtypes.Give,
		},
		{
			Key:         "board-game",
			Name:        "Настольная игра Ticket to Ride",
			Description: "Коробка в хорошем состоянии, все компоненты на месте.",
			Type:        dealtypes.Good,
			Action:      dealtypes.Give,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("create Clara offers: %w", err)
	}

	danOffers, err := client.createOffers(ctx, dan, []offerSpec{
		{
			Key:         "repair",
			Name:        "Мелкий ремонт техники",
			Description: "Помогу диагностировать и починить бытовые гаджеты.",
			Type:        dealtypes.Service,
			Action:      dealtypes.Give,
		},
		{
			Key:         "board-game-sleeves",
			Name:        "Протекторы для карт",
			Description: "Набор протекторов для настольных игр.",
			Type:        dealtypes.Good,
			Action:      dealtypes.Give,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("create Dan offers: %w", err)
	}

	offerGroupID, err := client.createOfferGroup(ctx, alice.Token, offerGroupRequest{
		Name:        new("Дом и уют"),
		Description: new("Набор для обмена на вещи для дома и организации пространства."),
		Units: []offerGroupUnitRequest{
			{
				Offers: []offerGroupOfferRef{
					{OfferID: aliceOffers["lamp"]},
				},
			},
			{
				Offers: []offerGroupOfferRef{
					{OfferID: aliceOffers["plant-consulting"]},
				},
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("create offer group: %w", err)
	}

	offerGroupDraftID, err := client.createDraftFromOfferGroup(ctx, bob.Token, offerGroupID, offerGroupDraftRequest{
		SelectedOfferIDs: []uuid.UUID{aliceOffers["lamp"], aliceOffers["plant-consulting"]},
		ResponderOfferID: uuidPtr(bobOffers["bike-bag"]),
		Name:             new("Отклик на набор Alice"),
		Description:      new("Черновик, созданный из набора офферов Alice."),
	})
	if err != nil {
		return nil, fmt.Errorf("create offer-group draft: %w", err)
	}

	discussionDealID, err := client.createTwoPartyDeal(ctx, alice, bob, aliceOffers["storage-boxes"], bobOffers["tool-kit"], "Обмен для домашнего кабинета", "Alice и Bob обсуждают обмен коробов для хранения на набор инструментов.")
	if err != nil {
		return nil, fmt.Errorf("create discussion deal: %w", err)
	}

	if _, err := client.promoteDealToDiscussion(ctx, discussionDealID, alice, bob); err != nil {
		return nil, fmt.Errorf("promote discussion deal: %w", err)
	}

	completedDealID, err := client.createTwoPartyDeal(ctx, clara, dan, claraOffers["board-game"], danOffers["repair"], "Игра в обмен на ремонт", "Clara и Dan договорились об обмене игры на помощь с ремонтом техники.")
	if err != nil {
		return nil, fmt.Errorf("create completed deal: %w", err)
	}

	completedDeal, err := client.promoteDealToDiscussion(ctx, completedDealID, clara, dan)
	if err != nil {
		return nil, fmt.Errorf("promote completed deal to discussion: %w", err)
	}

	if err := client.completeTwoPartyDeal(ctx, completedDealID, clara, dan); err != nil {
		return nil, fmt.Errorf("complete deal: %w", err)
	}

	if err := client.createMutualReviews(ctx, completedDealID, completedDeal, clara, dan); err != nil {
		return nil, fmt.Errorf("create reviews: %w", err)
	}

	directChatID, err := client.createDirectChat(ctx, alice.Token, bob.UserID)
	if err != nil {
		return nil, fmt.Errorf("create direct chat: %w", err)
	}

	if err := client.sendChatMessages(ctx, directChatID, []chatMessage{
		{Token: alice.Token, Content: "Привет! Я оставила черновик по твоему набору вещей."},
		{Token: bob.Token, Content: "Вижу, давай обсудим детали по времени и месту встречи."},
	}); err != nil {
		return nil, fmt.Errorf("send direct chat messages: %w", err)
	}

	dealChatID, err := client.waitForDealChat(ctx, alice.Token, discussionDealID)
	if err != nil {
		return nil, fmt.Errorf("wait for deal chat: %w", err)
	}

	if err := client.sendChatMessages(ctx, dealChatID, []chatMessage{
		{Token: alice.Token, Content: "Добавила детали по сделке и забронировала встречу на выходные."},
		{Token: bob.Token, Content: "Отлично, я беру инструмент и подтверждаю участие в сделке."},
	}); err != nil {
		return nil, fmt.Errorf("send deal chat messages: %w", err)
	}

	summary := &seedSummary{
		OfferGroupID:      offerGroupID,
		OfferGroupDraftID: offerGroupDraftID,
		DiscussionDealID:  discussionDealID,
		CompletedDealID:   completedDealID,
		DirectChatID:      directChatID,
		DealChatID:        dealChatID,
	}

	for _, user := range users {
		summary.Users = append(summary.Users, seededUserSummary{
			Key:      user.Key,
			UserID:   user.UserID.String(),
			Email:    user.Email,
			Password: user.Password,
		})
	}

	return summary, nil
}

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

		result[spec.Key] = offer.Id
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

	return body.Id, nil
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

	return body.Id, nil
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
		ClaimReceiver: boolPtr(true),
	}); err != nil {
		return dealtypes.Deal{}, fmt.Errorf("claim receiver for %s item: %w", userA.Key, err)
	}

	if err := c.updateDealItem(ctx, userA.Token, dealID, itemB, dealtypes.UpdateDealItemRequest{
		ClaimReceiver: boolPtr(true),
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
		itemByAuthor[item.AuthorId] = item.Id
	}

	itemA, ok := itemByAuthor[userA.UserID]
	if !ok {
		return fmt.Errorf("completed deal %s does not contain item for %s", dealID, userA.Key)
	}
	itemB, ok := itemByAuthor[userB.UserID]
	if !ok {
		return fmt.Errorf("completed deal %s does not contain item for %s", dealID, userB.Key)
	}

	if err := c.createDealItemReview(ctx, userB.Token, dealID, itemA, dealtypes.CreateReviewRequest{
		Rating:  5,
		Comment: new("Все прошло четко: договорились быстро и получили именно то, что ожидали."),
	}); err != nil {
		return fmt.Errorf("review for %s item: %w", userA.Key, err)
	}

	if err := c.createDealItemReview(ctx, userA.Token, dealID, itemB, dealtypes.CreateReviewRequest{
		Rating:  5,
		Comment: new("Хорошая коммуникация и удобная передача вещи."),
	}); err != nil {
		return fmt.Errorf("review for %s item: %w", userB.Key, err)
	}

	return nil
}

func (c *seedClient) createDealItemReview(ctx context.Context, token string, dealID uuid.UUID, itemID uuid.UUID, req dealtypes.CreateReviewRequest) error {
	path := fmt.Sprintf("/deals/%s/items/%s/reviews", dealID, itemID)
	return c.doJSON(ctx, http.MethodPost, path, token, req, nil, http.StatusCreated)
}

func (c *seedClient) createDirectChat(ctx context.Context, token string, participantID uuid.UUID) (uuid.UUID, error) {
	var body types.Chat
	if err := c.doJSON(ctx, http.MethodPost, "/chats", token, types.CreateChatRequest{
		ParticipantId: participantID,
	}, &body, http.StatusCreated); err != nil {
		return uuid.Nil, err
	}

	return body.Id, nil
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

		var body types.Chat
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

type chatMessage struct {
	Token   string
	Content string
}

func (c *seedClient) sendChatMessages(ctx context.Context, chatID uuid.UUID, messages []chatMessage) error {
	for _, message := range messages {
		path := fmt.Sprintf("/chats/%s/messages", chatID)
		if err := c.doJSON(ctx, http.MethodPost, path, message.Token, types.SendMessageRequest{
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

func containsStatus(statuses []int, status int) bool {
	for _, expected := range statuses {
		if expected == status {
			return true
		}
	}

	return false
}

func closeBody(closer io.Closer) {
	_ = closer.Close()
}

func boolPtr(v bool) *bool {
	return &v
}

func stringPtr(v string) *string {
	return &v
}

func uuidPtr(v uuid.UUID) *uuid.UUID {
	return &v
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}

	return ""
}

func durationFromEnv(key string, fallback time.Duration) time.Duration {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}

	value, err := time.ParseDuration(raw)
	if err != nil {
		return fallback
	}

	return value
}
