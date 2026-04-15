package main

import (
	dealtypes "barter-port/contracts/openapi/deals/types"
	usertypes "barter-port/contracts/openapi/users/types"
	"context"
	"fmt"

	"github.com/google/uuid"
)

func runSeed(ctx context.Context, client *seedClient, cfg seedConfig) (*seedSummary, error) {
	users := []seededUser{
		{
			Key:      "alice",
			Name:     "Alice Morozova",
			Bio:      "Меняю декор для дома, книги и комнатные растения.",
			Email:    "alice.demo@barterport.local",
			Password: cfg.Password,
			Avatar:   cfg.AvatarBaseURL + "/alice.png",
		},
		{
			Key:      "bob",
			Name:     "Bob Sokolov",
			Bio:      "Обмениваю велоаксессуары, инструменты и вещи для поездок.",
			Email:    "bob.demo@barterport.local",
			Password: cfg.Password,
			Avatar:   cfg.AvatarBaseURL + "/bob.png",
		},
		{
			Key:      "clara",
			Name:     "Clara Lebedeva",
			Bio:      "Предлагаю занятия по английскому и ищу настольные игры.",
			Email:    "clara.demo@barterport.local",
			Password: cfg.Password,
			Avatar:   cfg.AvatarBaseURL + "/clara.png",
		},
		{
			Key:      "dan",
			Name:     "Dan Volkov",
			Bio:      "Люблю ремонтировать мелкую технику и обменивать хобби-товары.",
			Email:    "dan.demo@barterport.local",
			Password: cfg.Password,
			Avatar:   cfg.AvatarBaseURL + "/dan.png",
		},
	}

	for i := range users {
		registered, token, err := client.ensureUser(ctx, users[i].Email, users[i].Password)
		if err != nil {
			return nil, fmt.Errorf("ensure user %s: %w", users[i].Key, err)
		}

		users[i].UserID = registered.UserID
		users[i].Token = token

		if _, err := client.updateMe(ctx, token, usertypes.UpdateUserRequest{
			Name:      new(users[i].Name),
			Bio:       new(users[i].Bio),
			AvatarUrl: new(users[i].Avatar),
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
		ResponderOfferID: new(bobOffers["bike-bag"]),
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

	completedDealChatID, err := client.waitForDealChat(ctx, clara.Token, completedDealID)
	if err != nil {
		return nil, fmt.Errorf("wait for deal chat: %w", err)
	}

	if err := client.sendChatMessages(ctx, completedDealChatID, []chatMessage{
		{Token: clara.Token, Content: "Все устраивает?"},
		{Token: dan.Token, Content: "Да, договорились! Когда и где встретимся?"},
	}); err != nil {
		return nil, fmt.Errorf("send chat messages for completed deal: %w", err)
	}

	if err := client.completeTwoPartyDeal(ctx, completedDealID, clara, dan); err != nil {
		return nil, fmt.Errorf("complete deal: %w", err)
	}

	if err := client.createMutualReviews(ctx, completedDealID, completedDeal, clara, dan); err != nil {
		return nil, fmt.Errorf("create reviews: %w", err)
	}

	if err := client.ensureMutualSubscription(ctx, alice, bob); err != nil {
		return nil, fmt.Errorf("ensure mutual subscription for direct chat: %w", err)
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
