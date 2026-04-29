package seed_demo

import (
	dealtypes "barter-port/contracts/openapi/deals/types"
	usertypes "barter-port/contracts/openapi/users/types"
	"context"
	"fmt"
	"net/url"

	"github.com/google/uuid"
)

func RunSeed(ctx context.Context, client *SeedClient, cfg SeedConfig) (*SeedSummary, error) {
	warnings := make([]string, 0)

	users := []seededUser{
		{
			Key:      "alice",
			Name:     "Alice Morozova",
			Bio:      "Меняю декор для дома, книги и комнатные растения.",
			Email:    "alice.demo@barterport.local",
			Password: cfg.Password,
		},
		{
			Key:      "bob",
			Name:     "Bob Sokolov",
			Bio:      "Обмениваю велоаксессуары, инструменты и вещи для поездок.",
			Email:    "bob.demo@barterport.local",
			Password: cfg.Password,
		},
		{
			Key:      "clara",
			Name:     "Clara Lebedeva",
			Bio:      "Предлагаю занятия по английскому и ищу настольные игры.",
			Email:    "clara.demo@barterport.local",
			Password: cfg.Password,
		},
		{
			Key:      "dan",
			Name:     "Dan Volkov",
			Bio:      "Люблю ремонтировать мелкую технику и обменивать хобби-товары.",
			Email:    "dan.demo@barterport.local",
			Password: cfg.Password,
		},
		{
			Key:      "eva",
			Name:     "Eva Nikiforova",
			Bio:      "Увлекаюсь акварелью и меняю художественные материалы.",
			Email:    "eva.demo@barterport.local",
			Password: cfg.Password,
		},
		{
			Key:      "fedor",
			Name:     "Fedor Gromov",
			Bio:      "Собираю виниловые пластинки и советскую электронику.",
			Email:    "fedor.demo@barterport.local",
			Password: cfg.Password,
		},
	}

	for i := range users {
		registered, token, err := client.ensureUser(ctx, users[i].Email, users[i].Password)
		if err != nil {
			return nil, fmt.Errorf("ensure user %s: %w", users[i].Key, err)
		}

		users[i].UserID = registered.UserID
		users[i].Token = token

		avatarPath, err := resolveUserAvatarPath(users[i].Key)
		if err != nil {
			return nil, fmt.Errorf("resolve avatar for %s: %w", users[i].Key, err)
		}

		var avatarURL *usertypes.AvatarUrl
		uploadedAvatar, err := client.uploadMeAvatar(ctx, token, avatarPath)
		if err != nil {
			if !isMediaUploadFallbackable(err) {
				return nil, fmt.Errorf("upload avatar for %s: %w", users[i].Key, err)
			}
			warnings = append(warnings, fmt.Sprintf("avatar skipped for %s: %v", users[i].Key, err))
		} else {
			avatarURL = new(uploadedAvatar.AvatarUrl)
		}

		if _, err := client.updateMe(ctx, token, usertypes.UpdateUserRequest{
			Name:      new(users[i].Name),
			Bio:       new(users[i].Bio),
			AvatarUrl: avatarURL,
		}); err != nil {
			return nil, fmt.Errorf("update profile for %s: %w", users[i].Key, err)
		}
	}

	alice := &users[0]
	bob := &users[1]
	clara := &users[2]
	dan := &users[3]
	eva := &users[4]
	fedor := &users[5]

	adminToken, err := client.ensureAdminToken(ctx, cfg.AdminEmail, cfg.AdminPassword)
	if err != nil {
		return nil, fmt.Errorf("ensure admin session: %w", err)
	}

	// ── Auth lifecycle ───────────────────────────────────────────────────────

	_, refreshCookie, err := client.loginWithRefreshCookie(ctx, alice.Email, alice.Password)
	if err != nil {
		return nil, fmt.Errorf("login with refresh cookie for alice: %w", err)
	}

	refreshedToken, refreshedCookie, err := client.refresh(ctx, refreshCookie)
	if err != nil {
		return nil, fmt.Errorf("refresh auth session for alice: %w", err)
	}

	if _, err := client.getMe(ctx, refreshedToken); err != nil {
		return nil, fmt.Errorf("get me after refresh for alice: %w", err)
	}

	if err := client.logout(ctx, refreshedCookie); err != nil {
		return nil, fmt.Errorf("logout refreshed session for alice: %w", err)
	}

	// ── Offers ───────────────────────────────────────────────────────────────

	aliceOffers, offerWarnings, err := client.createOffersWithWarnings(ctx, alice, []offerSpec{
		{
			Key:         "lamp",
			Name:        "Винтажная лампа",
			Description: "Латунная настольная лампа в рабочем состоянии.",
			Type:        dealtypes.Good,
			Action:      dealtypes.Give,
			Tags:        []dealtypes.TagName{"декор", "винтаж"},
			Latitude:    new(55.7500),
			Longitude:   new(37.5960),
		},
		{
			Key:         "plant-consulting",
			Name:        "Консультация по комнатным растениям",
			Description: "Помогу подобрать уход и пересадку для домашних растений.",
			Type:        dealtypes.Service,
			Action:      dealtypes.Give,
			Tags:        []dealtypes.TagName{"растения", "уход"},
			Latitude:    new(55.7632),
			Longitude:   new(37.6452),
		},
		{
			Key:         "storage-boxes",
			Name:        "Набор коробов для хранения",
			Description: "Три тканевых короба для стеллажа и хранения мелочей.",
			Type:        dealtypes.Good,
			Action:      dealtypes.Give,
			Tags:        []dealtypes.TagName{"хранение", "дом"},
			Latitude:    new(55.7437),
			Longitude:   new(37.6547),
		},
		{
			Key:          "bookshelf",
			Name:         "Ищу узкий книжный стеллаж",
			Description:  "Нужен компактный стеллаж для прихожей.",
			Type:         dealtypes.Good,
			Action:       dealtypes.Take,
			Tags:         []dealtypes.TagName{"мебель"},
			PhotoAliases: []string{"узкий книжный стеллаж"},
		},
		{
			Key:         "magazines",
			Name:        "Журналы по дизайну интерьера",
			Description: "Подборка за 2022–2023 годы, 12 номеров.",
			Type:        dealtypes.Good,
			Action:      dealtypes.Give,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("create Alice offers: %w", err)
	}
	warnings = append(warnings, offerWarnings...)

	bobOffers, offerWarnings, err := client.createOffersWithWarnings(ctx, bob, []offerSpec{
		{
			Key:         "bike-bag",
			Name:        "Велосипедная сумка",
			Description: "Влагозащищенная сумка на багажник.",
			Type:        dealtypes.Good,
			Action:      dealtypes.Give,
			Tags:        []dealtypes.TagName{"велосипед", "поездки"},
			Latitude:    new(55.7312),
			Longitude:   new(37.5987),
		},
		{
			Key:         "tool-kit",
			Name:        "Набор инструментов",
			Description: "Компактный набор ключей и отверток для дома.",
			Type:        dealtypes.Good,
			Action:      dealtypes.Give,
			Tags:        []dealtypes.TagName{"инструменты", "дом"},
			Latitude:    new(55.8011),
			Longitude:   new(37.6234),
		},
		{
			Key:          "coffee-beans",
			Name:         "Ищу зерновой кофе",
			Description:  "Обменяю на аксессуары для поездок или инструменты.",
			Type:         dealtypes.Good,
			Action:       dealtypes.Take,
			Tags:         []dealtypes.TagName{"кофе"},
			PhotoAliases: []string{"зерновой кофе"},
		},
		{
			Key:         "thermos",
			Name:        "Термос туристический",
			Description: "0.9 л, нержавейка, держит тепло 12 часов.",
			Type:        dealtypes.Good,
			Action:      dealtypes.Give,
			Tags:        []dealtypes.TagName{"туризм", "поездки"},
			Latitude:    new(55.7909),
			Longitude:   new(37.6698),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("create Bob offers: %w", err)
	}
	warnings = append(warnings, offerWarnings...)

	claraOffers, offerWarnings, err := client.createOffersWithWarnings(ctx, clara, []offerSpec{
		{
			Key:         "english-session",
			Name:        "Разговорный английский",
			Description: "Час разговорной практики онлайн или офлайн.",
			Type:        dealtypes.Service,
			Action:      dealtypes.Give,
			Tags:        []dealtypes.TagName{"языки", "уроки"},
			Latitude:    new(55.7603),
			Longitude:   new(37.6254),
		},
		{
			Key:         "board-game",
			Name:        "Настольная игра Ticket to Ride",
			Description: "Коробка в хорошем состоянии, все компоненты на месте.",
			Type:        dealtypes.Good,
			Action:      dealtypes.Give,
			Tags:        []dealtypes.TagName{"игры", "настолки"},
			Latitude:    new(55.7295),
			Longitude:   new(37.6453),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("create Clara offers: %w", err)
	}
	warnings = append(warnings, offerWarnings...)

	danOffers, offerWarnings, err := client.createOffersWithWarnings(ctx, dan, []offerSpec{
		{
			Key:         "repair",
			Name:        "Мелкий ремонт техники",
			Description: "Помогу диагностировать и починить бытовые гаджеты.",
			Type:        dealtypes.Service,
			Action:      dealtypes.Give,
			Tags:        []dealtypes.TagName{"ремонт", "техника"},
			Latitude:    new(55.7845),
			Longitude:   new(37.4987),
		},
		{
			Key:         "board-game-sleeves",
			Name:        "Протекторы для карт",
			Description: "Набор протекторов для настольных игр.",
			Type:        dealtypes.Good,
			Action:      dealtypes.Give,
			Tags:        []dealtypes.TagName{"игры", "карты"},
			Latitude:    new(55.7389),
			Longitude:   new(37.6567),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("create Dan offers: %w", err)
	}
	warnings = append(warnings, offerWarnings...)

	evaOffers, offerWarnings, err := client.createOffersWithWarnings(ctx, eva, []offerSpec{
		{
			Key:         "watercolors",
			Name:        "Акварельные краски",
			Description: "Набор из 24 цветов, почти новый.",
			Type:        dealtypes.Good,
			Action:      dealtypes.Give,
			Tags:        []dealtypes.TagName{"арт", "акварель"},
			Latitude:    new(55.7653),
			Longitude:   new(37.5843),
		},
		{
			Key:         "canvas-set",
			Name:        "Набор холстов",
			Description: "5 холстов 30×40 см, загрунтованы.",
			Type:        dealtypes.Good,
			Action:      dealtypes.Give,
			Tags:        []dealtypes.TagName{"арт", "холсты"},
			Latitude:    new(55.7721),
			Longitude:   new(37.6143),
		},
		{
			Key:         "art-lesson",
			Name:        "Урок по акварели",
			Description: "Индивидуальный урок 1.5 часа онлайн или очно.",
			Type:        dealtypes.Service,
			Action:      dealtypes.Give,
			Tags:        []dealtypes.TagName{"арт", "уроки"},
			Latitude:    new(55.7578),
			Longitude:   new(37.6619),
		},
		{
			Key:          "want-brushes",
			Name:         "Ищу кисти для акрила",
			Description:  "Нужен набор кистей разных размеров.",
			Type:         dealtypes.Good,
			Action:       dealtypes.Take,
			Tags:         []dealtypes.TagName{"кисти"},
			PhotoAliases: []string{"кисти для акрила"},
		},
		{
			Key:         "sketchbooks",
			Name:        "Блокноты для скетчей",
			Description: "Три блокнота A5 с плотной бумагой.",
			Type:        dealtypes.Good,
			Action:      dealtypes.Give,
			Tags:        []dealtypes.TagName{"арт", "скетчи"},
			Latitude:    new(55.7765),
			Longitude:   new(37.5826),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("create Eva offers: %w", err)
	}
	warnings = append(warnings, offerWarnings...)

	fedorOffers, offerWarnings, err := client.createOffersWithWarnings(ctx, fedor, []offerSpec{
		{
			Key:          "vinyl-player",
			Name:         "Виниловый проигрыватель",
			Description:  "Рабочий, советского производства, с иглой.",
			Type:         dealtypes.Good,
			Action:       dealtypes.Give,
			Tags:         []dealtypes.TagName{"винтаж", "музыка"},
			PhotoAliases: []string{"vinyl_player"},
			Latitude:     new(55.7534),
			Longitude:    new(37.6342),
		},
		{
			Key:          "old-camera",
			Name:         "Плёночная фотокамера Зенит-11",
			Description:  "С объективом Helios-44, рабочее состояние.",
			Type:         dealtypes.Good,
			Action:       dealtypes.Give,
			Tags:         []dealtypes.TagName{"фото", "винтаж"},
			PhotoAliases: []string{"film_camera_zenit_11"},
			Latitude:     new(55.7714),
			Longitude:    new(37.6789),
		},
		{
			Key:         "want-records",
			Name:        "Ищу виниловые пластинки",
			Description: "Интересует jazz и классика, любое состояние.",
			Type:        dealtypes.Good,
			Action:      dealtypes.Take,
			Tags:        []dealtypes.TagName{"винил", "музыка"},
			SkipPhoto:   true,
		},
		{
			Key:          "headphones",
			Name:         "Советские наушники ТДС-3",
			Description:  "Изодинамические, рабочие, в оригинальном чехле.",
			Type:         dealtypes.Good,
			Action:       dealtypes.Give,
			Tags:         []dealtypes.TagName{"аудио", "винтаж"},
			PhotoAliases: []string{"soviet_headphones"},
			Latitude:     new(55.7150),
			Longitude:    new(37.5568),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("create Fedor offers: %w", err)
	}
	warnings = append(warnings, offerWarnings...)

	if _, err := client.updateOffer(ctx, alice.Token, aliceOffers["plant-consulting"], dealtypes.UpdateOfferRequest{
		Description: new("Помогу с подбором ухода, пересадкой и базовой диагностикой домашних растений."),
		Tags:        &[]dealtypes.TagName{"растения", "ботаника"},
	}); err != nil {
		return nil, fmt.Errorf("update Alice plant consulting offer: %w", err)
	}

	tempOfferIDs, offerWarnings, err := client.createOffersWithWarnings(ctx, eva, []offerSpec{
		{
			Key:         "cleanup-tag",
			Name:        "Временный оффер для чистки тега",
			Description: "Технический оффер для проверки удаления тега и удаления объявления.",
			Type:        dealtypes.Good,
			Action:      dealtypes.Give,
			Tags:        []dealtypes.TagName{"времятег"},
			SkipPhoto:   true,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("create temp cleanup offer: %w", err)
	}
	warnings = append(warnings, offerWarnings...)

	if err := client.deleteTag(ctx, adminToken, "времятег"); err != nil {
		return nil, fmt.Errorf("delete admin tag: %w", err)
	}
	if err := client.deleteOffer(ctx, eva.Token, tempOfferIDs["cleanup-tag"]); err != nil {
		return nil, fmt.Errorf("delete temp cleanup offer: %w", err)
	}

	for _, view := range []struct {
		token   string
		offerID uuid.UUID
	}{
		{alice.Token, bobOffers["tool-kit"]},
		{clara.Token, bobOffers["tool-kit"]},
		{dan.Token, bobOffers["tool-kit"]},
		{alice.Token, fedorOffers["vinyl-player"]},
		{bob.Token, fedorOffers["vinyl-player"]},
		{eva.Token, fedorOffers["vinyl-player"]},
		{clara.Token, evaOffers["canvas-set"]},
	} {
		if err := client.viewOffer(ctx, view.token, view.offerID); err != nil {
			return nil, fmt.Errorf("view offer %s: %w", view.offerID, err)
		}
	}

	if err := client.addOfferToFavorites(ctx, alice.Token, bobOffers["tool-kit"]); err != nil {
		return nil, fmt.Errorf("favorite bob tool-kit for alice: %w", err)
	}
	if err := client.addOfferToFavorites(ctx, alice.Token, evaOffers["sketchbooks"]); err != nil {
		return nil, fmt.Errorf("favorite eva sketchbooks for alice: %w", err)
	}
	if err := client.addOfferToFavorites(ctx, bob.Token, aliceOffers["plant-consulting"]); err != nil {
		return nil, fmt.Errorf("favorite alice plant consulting for bob: %w", err)
	}
	if err := client.removeOfferFromFavorites(ctx, bob.Token, aliceOffers["plant-consulting"]); err != nil {
		return nil, fmt.Errorf("remove favorite alice plant consulting for bob: %w", err)
	}
	if err := client.addOfferToFavorites(ctx, clara.Token, bobOffers["bike-bag"]); err != nil {
		return nil, fmt.Errorf("favorite bob bike bag for clara: %w", err)
	}

	offersByPopularity := url.Values{}
	offersByPopularity.Set("sort", string(dealtypes.SortTypeByPopularity))
	offersByPopularity.Set("cursor_limit", "20")
	if _, err := client.listOffers(ctx, alice.Token, offersByPopularity); err != nil {
		return nil, fmt.Errorf("list offers by popularity: %w", err)
	}

	myOffersQuery := url.Values{}
	myOffersQuery.Set("my", "true")
	myOffersQuery.Set("sort", string(dealtypes.SortTypeByTime))
	if _, err := client.listOffers(ctx, alice.Token, myOffersQuery); err != nil {
		return nil, fmt.Errorf("list my offers for alice: %w", err)
	}

	offersByTag := url.Values{}
	offersByTag.Set("sort", string(dealtypes.SortTypeByTime))
	offersByTag.Set("tags", "растения")
	if _, err := client.listOffers(ctx, alice.Token, offersByTag); err != nil {
		return nil, fmt.Errorf("list offers by tag: %w", err)
	}

	offersWithoutTags := url.Values{}
	offersWithoutTags.Set("sort", string(dealtypes.SortTypeByTime))
	offersWithoutTags.Set("withoutTags", "true")
	if _, err := client.listOffers(ctx, alice.Token, offersWithoutTags); err != nil {
		return nil, fmt.Errorf("list offers without tags: %w", err)
	}

	if _, err := client.getOfferByID(ctx, alice.Token, bobOffers["tool-kit"]); err != nil {
		return nil, fmt.Errorf("get bob tool-kit by alice: %w", err)
	}

	// ── Offer groups and drafts ──────────────────────────────────────────────

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

	multiUnitGroupID, err := client.createOfferGroup(ctx, eva.Token, offerGroupRequest{
		Name:        new("Художественные материалы или урок"),
		Description: new("Можно забрать краски или записаться на урок и получить блокноты"),
		Units: []offerGroupUnitRequest{
			{
				Offers: []offerGroupOfferRef{
					{OfferID: evaOffers["watercolors"]},
					{OfferID: evaOffers["art-lesson"]},
				},
			},
			{
				Offers: []offerGroupOfferRef{
					{OfferID: evaOffers["sketchbooks"]},
				},
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("create multi-unit offer group: %w", err)
	}

	if _, err := client.listOfferGroups(ctx, alice.Token); err != nil {
		return nil, fmt.Errorf("list offer groups: %w", err)
	}
	if _, err := client.getOfferGroupByID(ctx, bob.Token, offerGroupID); err != nil {
		return nil, fmt.Errorf("get offer group by id: %w", err)
	}
	if _, err := client.listDrafts(ctx, bob.Token, false, true); err != nil {
		return nil, fmt.Errorf("list drafts for bob: %w", err)
	}
	if _, err := client.getDraftByID(ctx, bob.Token, offerGroupDraftID); err != nil {
		return nil, fmt.Errorf("get offer-group draft by id: %w", err)
	}

	cancelDraftID, err := client.createDraft(ctx, alice.Token, dealtypes.CreateDraftDealRequest{
		Name:        new("Черновик для отмены"),
		Description: new("Технический черновик для проверки cancel draft."),
		Offers: []dealtypes.OfferIDAndQuantity{
			{OfferID: aliceOffers["magazines"], Quantity: 1},
			{OfferID: bobOffers["thermos"], Quantity: 1},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("create cancel draft: %w", err)
	}
	if err := client.cancelDraft(ctx, bob.Token, cancelDraftID); err != nil {
		return nil, fmt.Errorf("cancel draft by bob: %w", err)
	}

	deleteDraftID, err := client.createDraft(ctx, clara.Token, dealtypes.CreateDraftDealRequest{
		Name:        new("Черновик для удаления"),
		Description: new("Технический черновик для проверки delete draft."),
		Offers: []dealtypes.OfferIDAndQuantity{
			{OfferID: claraOffers["board-game"], Quantity: 1},
			{OfferID: danOffers["board-game-sleeves"], Quantity: 1},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("create delete draft: %w", err)
	}
	if err := client.deleteDraft(ctx, dan.Token, deleteDraftID); err != nil {
		return nil, fmt.Errorf("delete draft by dan: %w", err)
	}

	// ── Deals ────────────────────────────────────────────────────────────────

	lookingDealID, err := client.createTwoPartyDeal(ctx, fedor, eva, fedorOffers["vinyl-player"], evaOffers["sketchbooks"],
		"Проигрыватель на скетчбуки",
		"Fedor и Eva собрали открытую сделку и оставили её в поиске участников для уточнения условий обмена.")
	if err != nil {
		return nil, fmt.Errorf("create open deal: %w", err)
	}

	discussionDealID, err := client.createTwoPartyDeal(ctx, alice, bob, aliceOffers["storage-boxes"], bobOffers["tool-kit"],
		"Обмен для домашнего кабинета",
		"Alice и Bob обсуждают обмен коробов для хранения на набор инструментов.")
	if err != nil {
		return nil, fmt.Errorf("create discussion deal: %w", err)
	}

	if _, err := client.updateDeal(ctx, alice.Token, discussionDealID, dealtypes.UpdateDealRequest{
		Name: "Обновлённый обмен для домашнего кабинета",
	}); err != nil {
		return nil, fmt.Errorf("update discussion deal name: %w", err)
	}

	if _, err := client.promoteDealToDiscussion(ctx, discussionDealID, alice, bob); err != nil {
		return nil, fmt.Errorf("promote discussion deal: %w", err)
	}

	confirmedDealID, err := client.createTwoPartyDeal(ctx, clara, eva, claraOffers["english-session"], evaOffers["art-lesson"],
		"Обмен уроками: английский на акварель",
		"Clara и Eva меняются уроками и ждут встречи.")
	if err != nil {
		return nil, fmt.Errorf("create confirmed deal: %w", err)
	}

	if _, err := client.promoteDealToDiscussion(ctx, confirmedDealID, clara, eva); err != nil {
		return nil, fmt.Errorf("promote confirmed deal to discussion: %w", err)
	}

	if err := client.changeDealStatus(ctx, clara.Token, confirmedDealID, dealtypes.Confirmed); err != nil {
		return nil, fmt.Errorf("confirm deal by clara: %w", err)
	}
	if _, err := client.getDealStatusVotes(ctx, clara.Token, confirmedDealID); err != nil {
		return nil, fmt.Errorf("get status votes for confirmed deal: %w", err)
	}
	if err := client.changeDealStatus(ctx, eva.Token, confirmedDealID, dealtypes.Confirmed); err != nil {
		return nil, fmt.Errorf("confirm deal by eva: %w", err)
	}

	completedDealID, err := client.createTwoPartyDeal(ctx, clara, dan, claraOffers["board-game"], danOffers["repair"],
		"Игра в обмен на ремонт",
		"Clara и Dan договорились об обмене игры на помощь с ремонтом техники.")
	if err != nil {
		return nil, fmt.Errorf("create completed deal: %w", err)
	}

	completedDeal, err := client.promoteDealToDiscussion(ctx, completedDealID, clara, dan)
	if err != nil {
		return nil, fmt.Errorf("promote completed deal to discussion: %w", err)
	}

	completedDealChatID, err := client.waitForDealChat(ctx, clara.Token, completedDealID)
	if err != nil {
		return nil, fmt.Errorf("wait for completed deal chat: %w", err)
	}
	if err := client.sendChatMessages(ctx, completedDealChatID, []chatMessage{
		{Token: clara.Token, Content: "Все устраивает?"},
		{Token: dan.Token, Content: "Да, договорились. Когда и где встретимся?"},
	}); err != nil {
		return nil, fmt.Errorf("send messages for completed deal: %w", err)
	}

	if err := client.completeTwoPartyDeal(ctx, completedDealID, clara, dan); err != nil {
		return nil, fmt.Errorf("complete deal: %w", err)
	}

	danItemInCompleted, err := itemByAuthorID(completedDeal.Items, dan.UserID)
	if err != nil {
		return nil, fmt.Errorf("find Dan item in completed deal: %w", err)
	}

	if _, err := client.listPendingDealReviews(ctx, clara.Token, completedDealID); err != nil {
		return nil, fmt.Errorf("list pending reviews for clara: %w", err)
	}
	if _, err := client.getDealItemReviewEligibility(ctx, clara.Token, completedDealID, danItemInCompleted.Id); err != nil {
		return nil, fmt.Errorf("get review eligibility for clara on Dan item: %w", err)
	}

	_, claraReviewOnDan, err := client.createMutualReviews(ctx, completedDealID, completedDeal, clara, dan, 5, 5,
		"Все прошло четко: договорились быстро и получили именно то, что ожидали.",
		"Хорошая коммуникация и удобная передача вещи.",
	)
	if err != nil {
		return nil, fmt.Errorf("create reviews for completed deal: %w", err)
	}

	if _, err := client.listPendingDealReviews(ctx, clara.Token, completedDealID); err != nil {
		return nil, fmt.Errorf("list pending reviews for clara after review creation: %w", err)
	}
	if _, err := client.getReviewByID(ctx, clara.Token, claraReviewOnDan.Id); err != nil {
		return nil, fmt.Errorf("get clara review by id: %w", err)
	}
	if _, err := client.updateReview(ctx, clara.Token, claraReviewOnDan.Id, dealtypes.UpdateReviewRequest{
		Rating:  new(4),
		Comment: new("Хороший обмен, немного сдвинули время встречи, но всё прошло спокойно."),
	}); err != nil {
		return nil, fmt.Errorf("update clara review on Dan: %w", err)
	}
	if _, err := client.listDealReviews(ctx, alice.Token, completedDealID); err != nil {
		return nil, fmt.Errorf("list reviews for completed deal: %w", err)
	}
	if _, err := client.listDealItemReviews(ctx, alice.Token, completedDealID, danItemInCompleted.Id); err != nil {
		return nil, fmt.Errorf("list item reviews for Dan item: %w", err)
	}
	if _, err := client.listOfferReviews(ctx, alice.Token, danOffers["repair"]); err != nil {
		return nil, fmt.Errorf("list offer reviews for Dan repair: %w", err)
	}
	if _, err := client.getOfferReviewsSummary(ctx, alice.Token, danOffers["repair"]); err != nil {
		return nil, fmt.Errorf("get offer review summary for Dan repair: %w", err)
	}
	if _, err := client.listProviderReviews(ctx, alice.Token, dan.UserID); err != nil {
		return nil, fmt.Errorf("list provider reviews for Dan: %w", err)
	}
	if _, err := client.getProviderReviewsSummary(ctx, alice.Token, dan.UserID); err != nil {
		return nil, fmt.Errorf("get provider review summary for Dan: %w", err)
	}
	if _, err := client.listAuthorReviews(ctx, alice.Token, clara.UserID); err != nil {
		return nil, fmt.Errorf("list author reviews for Clara: %w", err)
	}

	completedDeal2ID, err := client.createTwoPartyDeal(ctx, alice, bob, aliceOffers["magazines"], bobOffers["thermos"],
		"Журналы на термос",
		"Alice и Bob обменялись журналами и термосом.")
	if err != nil {
		return nil, fmt.Errorf("create completed deal 2: %w", err)
	}

	completedDeal2, err := client.promoteDealToDiscussion(ctx, completedDeal2ID, alice, bob)
	if err != nil {
		return nil, fmt.Errorf("promote completed deal 2: %w", err)
	}
	if err := client.completeTwoPartyDeal(ctx, completedDeal2ID, alice, bob); err != nil {
		return nil, fmt.Errorf("complete deal 2: %w", err)
	}
	if _, _, err := client.createMutualReviews(ctx, completedDeal2ID, completedDeal2, alice, bob, 4, 4,
		"Всё как договаривались, чуть задержались с передачей.",
		"Хороший обмен, термос в отличном состоянии.",
	); err != nil {
		return nil, fmt.Errorf("create reviews for completed deal 2: %w", err)
	}

	tempCompletedDealID, err := client.createTwoPartyDeal(ctx, alice, eva, aliceOffers["plant-consulting"], evaOffers["watercolors"],
		"Технический обмен для review CRUD",
		"Временная завершенная сделка, чтобы покрыть update item и delete review.")
	if err != nil {
		return nil, fmt.Errorf("create temp completed deal: %w", err)
	}

	tempCompletedDeal, err := client.promoteDealToDiscussion(ctx, tempCompletedDealID, alice, eva)
	if err != nil {
		return nil, fmt.Errorf("promote temp completed deal to discussion: %w", err)
	}

	aliceItemInTempCompleted, err := itemByAuthorID(tempCompletedDeal.Items, alice.UserID)
	if err != nil {
		return nil, fmt.Errorf("find Alice item in temp completed deal: %w", err)
	}
	if err := client.updateDealItem(ctx, alice.Token, tempCompletedDealID, aliceItemInTempCompleted.Id, dealtypes.UpdateDealItemRequest{
		Name:        new("Расширенная консультация по растениям"),
		Description: new("Добавила подробный разбор освещения, грунта и графика полива."),
	}); err != nil {
		return nil, fmt.Errorf("update Alice item in temp completed deal: %w", err)
	}

	if err := client.completeTwoPartyDeal(ctx, tempCompletedDealID, alice, eva); err != nil {
		return nil, fmt.Errorf("complete temp review CRUD deal: %w", err)
	}

	evaReviewOnAlice, _, err := client.createMutualReviews(ctx, tempCompletedDealID, tempCompletedDeal, alice, eva, 5, 5,
		"Очень полезная консультация, получила конкретный план ухода.",
		"Краски пришли в отличном состоянии, всё соответствует описанию.",
	)
	if err != nil {
		return nil, fmt.Errorf("create reviews for temp completed deal: %w", err)
	}

	if _, err := client.getReviewByID(ctx, eva.Token, evaReviewOnAlice.Id); err != nil {
		return nil, fmt.Errorf("get temp offer+item review by id: %w", err)
	}
	if _, err := client.listDealItemReviews(ctx, bob.Token, tempCompletedDealID, aliceItemInTempCompleted.Id); err != nil {
		return nil, fmt.Errorf("list item reviews for updated Alice item: %w", err)
	}
	if err := client.deleteReview(ctx, eva.Token, evaReviewOnAlice.Id); err != nil {
		return nil, fmt.Errorf("delete temp offer+item review: %w", err)
	}

	completedDeal3ID, err := client.createTwoPartyDeal(ctx, eva, fedor, evaOffers["canvas-set"], fedorOffers["headphones"],
		"Холсты на наушники",
		"Eva и Fedor обменялись художественными материалами и винтажными наушниками.")
	if err != nil {
		return nil, fmt.Errorf("create completed deal 3: %w", err)
	}

	completedDeal3, err := client.promoteDealToDiscussion(ctx, completedDeal3ID, eva, fedor)
	if err != nil {
		return nil, fmt.Errorf("promote completed deal 3: %w", err)
	}
	if err := client.completeTwoPartyDeal(ctx, completedDeal3ID, eva, fedor); err != nil {
		return nil, fmt.Errorf("complete deal 3: %w", err)
	}
	if _, _, err := client.createMutualReviews(ctx, completedDeal3ID, completedDeal3, eva, fedor, 3, 5,
		"Наушники рабочие, но состояние хуже, чем ожидалось.",
		"Холсты отличного качества, очень доволен обменом.",
	); err != nil {
		return nil, fmt.Errorf("create reviews for completed deal 3: %w", err)
	}

	cancelledDealID, err := client.createAndCancelDeal(ctx, bob, dan, bobOffers["bike-bag"], danOffers["board-game-sleeves"],
		"Велосумка на протекторы",
		"Bob и Dan не смогли договориться о времени встречи.")
	if err != nil {
		return nil, fmt.Errorf("create cancelled deal: %w", err)
	}

	revokeVoteDealID, err := createMultiPartyDeal(
		ctx,
		client,
		alice,
		[]*seededUser{alice, bob, clara, dan},
		[]uuid.UUID{aliceOffers["lamp"], bobOffers["tool-kit"], claraOffers["board-game"], danOffers["board-game-sleeves"]},
		"Коллективный обмен для revoke vote",
		"Техническая сделка, чтобы покрыть revoke голосов по провалу.",
	)
	if err != nil {
		return nil, fmt.Errorf("create revoke-vote deal: %w", err)
	}

	if _, err := promoteMultiPartyDealToDiscussion(
		ctx,
		client,
		revokeVoteDealID,
		alice,
		[]*seededUser{alice, bob, clara, dan},
		map[uuid.UUID]*seededUser{
			alice.UserID: bob,
			bob.UserID:   clara,
			clara.UserID: dan,
			dan.UserID:   alice,
		},
	); err != nil {
		return nil, fmt.Errorf("promote revoke-vote deal to discussion: %w", err)
	}

	if err := client.voteForFailure(ctx, alice.Token, revokeVoteDealID, bob.UserID); err != nil {
		return nil, fmt.Errorf("vote for failure in revoke-vote deal: %w", err)
	}
	if _, err := client.getFailureVotes(ctx, alice.Token, revokeVoteDealID); err != nil {
		return nil, fmt.Errorf("get failure votes before revoke: %w", err)
	}
	if err := client.revokeVoteForFailure(ctx, alice.Token, revokeVoteDealID); err != nil {
		return nil, fmt.Errorf("revoke failure vote: %w", err)
	}
	if err := client.changeDealStatus(ctx, bob.Token, revokeVoteDealID, dealtypes.Cancelled); err != nil {
		return nil, fmt.Errorf("cancel revoke-vote deal: %w", err)
	}

	pendingFailedDealID, err := client.createTwoPartyDeal(ctx, bob, eva, bobOffers["thermos"], evaOffers["watercolors"],
		"Термос на акварель",
		"Bob и Eva договорились об обмене, но только Bob зафиксировал провал сделки.")
	if err != nil {
		return nil, fmt.Errorf("create pending failed deal: %w", err)
	}

	if _, err := client.promoteDealToDiscussion(ctx, pendingFailedDealID, bob, eva); err != nil {
		return nil, fmt.Errorf("promote pending failed deal to discussion: %w", err)
	}

	if err := client.changeDealStatus(ctx, bob.Token, pendingFailedDealID, dealtypes.Confirmed); err != nil {
		return nil, fmt.Errorf("confirm pending failed deal by bob: %w", err)
	}
	if err := client.changeDealStatus(ctx, eva.Token, pendingFailedDealID, dealtypes.Confirmed); err != nil {
		return nil, fmt.Errorf("confirm pending failed deal by eva: %w", err)
	}
	if err := client.voteForFailure(ctx, bob.Token, pendingFailedDealID, eva.UserID); err != nil {
		return nil, fmt.Errorf("bob vote for pending failure: %w", err)
	}
	if _, err := client.getFailureVotes(ctx, bob.Token, pendingFailedDealID); err != nil {
		return nil, fmt.Errorf("get failure votes for pending failed deal: %w", err)
	}

	failedDealID, err := client.createTwoPartyDeal(ctx, alice, fedor, aliceOffers["lamp"], fedorOffers["old-camera"],
		"Лампа на фотоаппарат",
		"Alice и Fedor пытались обменять лампу на фотокамеру.")
	if err != nil {
		return nil, fmt.Errorf("create failed deal: %w", err)
	}

	if _, err := client.promoteDealToDiscussion(ctx, failedDealID, alice, fedor); err != nil {
		return nil, fmt.Errorf("promote failed deal to discussion: %w", err)
	}

	failedDealChatID, err := client.waitForDealChat(ctx, alice.Token, failedDealID)
	if err != nil {
		return nil, fmt.Errorf("wait for failed deal chat: %w", err)
	}
	if err := client.sendChatMessages(ctx, failedDealChatID, []chatMessage{
		{Token: alice.Token, Content: "Кажется, состояние камеры хуже, чем в описании."},
		{Token: fedor.Token, Content: "Не согласен, но давай зафиксируем всё для модератора."},
	}); err != nil {
		return nil, fmt.Errorf("send failed deal chat messages: %w", err)
	}

	if err := client.changeDealStatus(ctx, alice.Token, failedDealID, dealtypes.Confirmed); err != nil {
		return nil, fmt.Errorf("confirm failed deal by alice: %w", err)
	}
	if err := client.changeDealStatus(ctx, fedor.Token, failedDealID, dealtypes.Confirmed); err != nil {
		return nil, fmt.Errorf("confirm failed deal by fedor: %w", err)
	}

	if err := client.voteForFailure(ctx, alice.Token, failedDealID, fedor.UserID); err != nil {
		return nil, fmt.Errorf("alice vote for failure: %w", err)
	}
	if _, err := client.getFailureVotes(ctx, adminToken, failedDealID); err != nil {
		return nil, fmt.Errorf("get failure votes for failed deal: %w", err)
	}
	if _, err := client.listDealsForFailureReview(ctx, adminToken); err != nil {
		return nil, fmt.Errorf("list deals for failure review: %w", err)
	}
	if _, err := client.getFailureMaterials(ctx, adminToken, failedDealID); err != nil {
		return nil, fmt.Errorf("get failure materials: %w", err)
	}
	if err := client.voteForFailure(ctx, fedor.Token, failedDealID, alice.UserID); err != nil {
		return nil, fmt.Errorf("fedor vote for failure: %w", err)
	}

	if err := client.moderatorResolutionForFailure(ctx, adminToken, failedDealID, dealtypes.ModeratorResolutionForFailureRequest{
		Confirmed:        true,
		UserId:           new(fedor.UserID),
		PunishmentPoints: new(10),
		Comment:          new("Сделка провалена по вине Fedor: товар не соответствовал описанию."),
	}); err != nil {
		return nil, fmt.Errorf("moderator resolution for failure: %w", err)
	}
	if _, err := client.getModeratorResolutionForFailure(ctx, alice.Token, failedDealID); err != nil {
		return nil, fmt.Errorf("get moderator resolution for failed deal: %w", err)
	}

	joinDealID, err := client.createTwoPartyDeal(ctx, bob, dan, bobOffers["coffee-beans"], danOffers["repair"],
		"Кофе на мелкий ремонт",
		"Bob и Dan оставили сделку открытой: кофе в обмен на помощь с техникой, позже Bob добавил термос.")
	if err != nil {
		return nil, fmt.Errorf("create join-ready deal: %w", err)
	}

	joinDeal, err := client.addDealItem(ctx, bob.Token, joinDealID, dealtypes.AddDealItemRequest{
		OfferId:  bobOffers["thermos"],
		Quantity: 1,
	})
	if err != nil {
		return nil, fmt.Errorf("add item to join deal: %w", err)
	}

	thermosItemInJoinDeal, err := itemByOfferID(joinDeal.Items, bobOffers["thermos"])
	if err != nil {
		return nil, fmt.Errorf("find thermos item in join deal: %w", err)
	}
	if err := client.updateDealItem(ctx, bob.Token, joinDealID, thermosItemInJoinDeal.Id, dealtypes.UpdateDealItemRequest{
		Name:        new("Термос туристический 0.9 л"),
		Description: new("Добавил детали по состоянию и комплектности, крышка без дефектов."),
		Quantity:    new(2),
	}); err != nil {
		return nil, fmt.Errorf("update thermos item in join deal: %w", err)
	}

	if err := client.requestJoinDeal(ctx, eva.Token, joinDealID); err != nil {
		return nil, fmt.Errorf("eva request join deal: %w", err)
	}
	if _, err := client.getDealJoinRequests(ctx, bob.Token, joinDealID); err != nil {
		return nil, fmt.Errorf("get join requests for join deal: %w", err)
	}
	if err := client.processJoinRequest(ctx, bob.Token, joinDealID, eva.UserID, true); err != nil {
		return nil, fmt.Errorf("bob process join request: %w", err)
	}
	if err := client.processJoinRequest(ctx, dan.Token, joinDealID, eva.UserID, true); err != nil {
		return nil, fmt.Errorf("dan process join request: %w", err)
	}

	leaveJoinDealID, err := client.createTwoPartyDeal(ctx, clara, alice, claraOffers["english-session"], aliceOffers["storage-boxes"],
		"Английский на короба для хранения",
		"Clara и Alice оставили сделку открытой, чтобы рассмотреть дополнительные предложения по обмену.")
	if err != nil {
		return nil, fmt.Errorf("create leave-join-ready deal: %w", err)
	}
	if err := client.requestJoinDeal(ctx, eva.Token, leaveJoinDealID); err != nil {
		return nil, fmt.Errorf("eva request leave-join deal: %w", err)
	}
	if _, err := client.getDealJoinRequests(ctx, clara.Token, leaveJoinDealID); err != nil {
		return nil, fmt.Errorf("get join requests for leave-join deal: %w", err)
	}
	if err := client.leaveDeal(ctx, eva.Token, leaveJoinDealID); err != nil {
		return nil, fmt.Errorf("eva leave join deal: %w", err)
	}

	openDealsQuery := url.Values{}
	openDealsQuery.Set("open", "true")
	if _, err := client.listDeals(ctx, alice.Token, openDealsQuery); err != nil {
		return nil, fmt.Errorf("list open deals: %w", err)
	}

	myDealsQuery := url.Values{}
	myDealsQuery.Set("my", "true")
	if _, err := client.listDeals(ctx, alice.Token, myDealsQuery); err != nil {
		return nil, fmt.Errorf("list my deals through generic endpoint: %w", err)
	}

	// ── Subscriptions ────────────────────────────────────────────────────────

	if err := client.ensureMutualSubscription(ctx, alice, bob); err != nil {
		return nil, fmt.Errorf("mutual subscription alice <-> bob: %w", err)
	}
	if err := client.ensureMutualSubscription(ctx, clara, dan); err != nil {
		return nil, fmt.Errorf("mutual subscription clara <-> dan: %w", err)
	}
	if err := client.ensureMutualSubscription(ctx, eva, fedor); err != nil {
		return nil, fmt.Errorf("mutual subscription eva <-> fedor: %w", err)
	}
	if err := client.subscribeToUser(ctx, alice.Token, eva.UserID); err != nil {
		return nil, fmt.Errorf("subscribe alice -> eva: %w", err)
	}
	if err := client.subscribeToUser(ctx, bob.Token, fedor.UserID); err != nil {
		return nil, fmt.Errorf("subscribe bob -> fedor: %w", err)
	}
	if err := client.subscribeToUser(ctx, clara.Token, alice.UserID); err != nil {
		return nil, fmt.Errorf("subscribe clara -> alice: %w", err)
	}
	if err := client.subscribeToUser(ctx, bob.Token, clara.UserID); err != nil {
		return nil, fmt.Errorf("subscribe bob -> clara temporary: %w", err)
	}
	if err := client.unsubscribeFromUser(ctx, bob.Token, clara.UserID); err != nil {
		return nil, fmt.Errorf("unsubscribe bob -> clara temporary: %w", err)
	}

	if _, err := client.listSubscriptions(ctx, alice.Token); err != nil {
		return nil, fmt.Errorf("list subscriptions for alice: %w", err)
	}
	if _, err := client.listSubscriptionsByUser(ctx, alice.Token, bob.UserID); err != nil {
		return nil, fmt.Errorf("list subscriptions by Bob ID: %w", err)
	}
	if _, err := client.listMySubscribers(ctx, eva.Token); err != nil {
		return nil, fmt.Errorf("list my subscribers for Eva: %w", err)
	}
	if _, err := client.listSubscribersByUser(ctx, alice.Token, alice.UserID); err != nil {
		return nil, fmt.Errorf("list subscribers by Alice ID: %w", err)
	}

	subscribedOffersQuery := url.Values{}
	subscribedOffersQuery.Set("sort", string(dealtypes.ByTime))
	subscribedOffersQuery.Set("cursor_limit", "20")
	if _, err := client.listSubscribedOffers(ctx, alice.Token, subscribedOffersQuery); err != nil {
		return nil, fmt.Errorf("list subscribed offers for alice: %w", err)
	}
	favoriteOffersQuery := url.Values{}
	favoriteOffersQuery.Set("cursor_limit", "20")
	if _, err := client.listFavoriteOffers(ctx, alice.Token, favoriteOffersQuery); err != nil {
		return nil, fmt.Errorf("list favorite offers for alice: %w", err)
	}

	// ── Chats ────────────────────────────────────────────────────────────────

	directChatID, err := client.createDirectChat(ctx, alice.Token, bob.UserID)
	if err != nil {
		return nil, fmt.Errorf("create direct chat alice-bob: %w", err)
	}
	if err := client.sendChatMessages(ctx, directChatID, []chatMessage{
		{Token: alice.Token, Content: "Привет! Я оставила черновик по твоему набору вещей."},
		{Token: bob.Token, Content: "Вижу, давай обсудим детали по времени и месту встречи."},
	}); err != nil {
		return nil, fmt.Errorf("send direct chat messages: %w", err)
	}

	if _, err := client.listChats(ctx, alice.Token); err != nil {
		return nil, fmt.Errorf("list chats for alice: %w", err)
	}
	if _, err := client.listChatUsers(ctx, alice.Token); err != nil {
		return nil, fmt.Errorf("list chat users for alice: %w", err)
	}
	directMessages, err := client.getChatMessages(ctx, alice.Token, directChatID, nil)
	if err != nil {
		return nil, fmt.Errorf("get direct chat messages: %w", err)
	}
	if len(directMessages) > 0 {
		if _, err := client.getChatMessages(ctx, alice.Token, directChatID, new(directMessages[0].CreatedAt)); err != nil {
			return nil, fmt.Errorf("get direct chat messages after cursor: %w", err)
		}
	}

	dealChatID, err := client.waitForDealChat(ctx, alice.Token, discussionDealID)
	if err != nil {
		return nil, fmt.Errorf("wait for deal chat: %w", err)
	}
	if err := client.sendChatMessages(ctx, dealChatID, []chatMessage{
		{Token: alice.Token, Content: "Добавила детали по сделке и предлагаю встретиться на выходных."},
		{Token: bob.Token, Content: "Отлично, беру инструмент и подтверждаю участие в сделке."},
	}); err != nil {
		return nil, fmt.Errorf("send deal chat messages: %w", err)
	}
	if _, err := client.getChatMessages(ctx, alice.Token, dealChatID, nil); err != nil {
		return nil, fmt.Errorf("get discussion deal chat messages: %w", err)
	}

	confirmedDealChatID, err := client.waitForDealChat(ctx, clara.Token, confirmedDealID)
	if err != nil {
		return nil, fmt.Errorf("wait for confirmed deal chat: %w", err)
	}
	if err := client.sendChatMessages(ctx, confirmedDealChatID, []chatMessage{
		{Token: clara.Token, Content: "Когда тебе удобно провести урок? Я свободна в субботу."},
		{Token: eva.Token, Content: "Отлично, договорились на субботу в 12:00."},
	}); err != nil {
		return nil, fmt.Errorf("send confirmed deal chat messages: %w", err)
	}

	claraDanChatID, err := client.createDirectChat(ctx, clara.Token, dan.UserID)
	if err != nil {
		return nil, fmt.Errorf("create direct chat clara-dan: %w", err)
	}
	if err := client.sendChatMessages(ctx, claraDanChatID, []chatMessage{
		{Token: clara.Token, Content: "Дан, у тебя случайно нет чего-то интересного для обмена?"},
		{Token: dan.Token, Content: "Есть пара вещей, посмотри мои объявления и напиши если что-то подойдёт."},
	}); err != nil {
		return nil, fmt.Errorf("send clara-dan chat messages: %w", err)
	}

	// ── Offer reports ────────────────────────────────────────────────────────

	pendingReport, err := client.createOfferReport(ctx, clara.Token, danOffers["board-game-sleeves"],
		"Объявление содержит недостоверное описание товара: количество протекторов не соответствует.")
	if err != nil {
		return nil, fmt.Errorf("create pending report: %w", err)
	}

	acceptedReport, err := client.createOfferReport(ctx, alice.Token, fedorOffers["want-records"],
		"Объявление размещено повторно и нарушает правила площадки: похоже на дубликат.")
	if err != nil {
		return nil, fmt.Errorf("create accepted report: %w", err)
	}
	if _, err := client.resolveOfferReport(ctx, adminToken, acceptedReport.Id, true, new("Повторная публикация, применён штраф.")); err != nil {
		return nil, fmt.Errorf("resolve accepted report: %w", err)
	}

	rejectedReport, err := client.createOfferReport(ctx, dan.Token, evaOffers["canvas-set"],
		"Подозрительная цена на обмен, похоже на коммерческое объявление.")
	if err != nil {
		return nil, fmt.Errorf("create rejected report: %w", err)
	}
	if _, err := client.resolveOfferReport(ctx, adminToken, rejectedReport.Id, false, new("Жалоба не подтверждена: объявление соответствует правилам.")); err != nil {
		return nil, fmt.Errorf("resolve rejected report: %w", err)
	}

	if _, err := client.getOfferReports(ctx, dan.Token, danOffers["board-game-sleeves"]); err != nil {
		return nil, fmt.Errorf("get offer reports by offer id: %w", err)
	}
	if _, err := client.listOfferReportsForAdmin(ctx, adminToken, nil); err != nil {
		return nil, fmt.Errorf("list all offer reports for admin: %w", err)
	}
	if _, err := client.listOfferReportsForAdmin(ctx, adminToken, new(dealtypes.Pending)); err != nil {
		return nil, fmt.Errorf("list pending offer reports for admin: %w", err)
	}
	if _, err := client.getOfferReportForAdmin(ctx, adminToken, acceptedReport.Id); err != nil {
		return nil, fmt.Errorf("get offer report details for admin: %w", err)
	}
	if _, err := client.getOfferByID(ctx, fedor.Token, fedorOffers["want-records"]); err != nil {
		return nil, fmt.Errorf("get hidden offer by author: %w", err)
	}
	if _, err := client.getOfferByID(ctx, adminToken, fedorOffers["want-records"]); err != nil {
		return nil, fmt.Errorf("get hidden offer by admin: %w", err)
	}

	// ── Users and statistics ────────────────────────────────────────────────

	if _, err := client.getUserByID(ctx, alice.Token, bob.UserID); err != nil {
		return nil, fmt.Errorf("get Bob by id: %w", err)
	}
	if _, err := client.getReputationEvents(ctx, alice.Token); err != nil {
		return nil, fmt.Errorf("get Alice reputation events: %w", err)
	}
	if _, err := client.getReputationEvents(ctx, fedor.Token); err != nil {
		return nil, fmt.Errorf("get Fedor reputation events: %w", err)
	}
	if _, err := client.getMyStatistics(ctx, alice.Token); err != nil {
		return nil, fmt.Errorf("get Alice statistics: %w", err)
	}
	if _, err := client.getMyStatistics(ctx, fedor.Token); err != nil {
		return nil, fmt.Errorf("get Fedor statistics: %w", err)
	}
	if _, err := client.listTags(ctx, alice.Token); err != nil {
		return nil, fmt.Errorf("list tags: %w", err)
	}

	summary := &SeedSummary{
		Warnings: warnings,

		OfferGroupID:      offerGroupID,
		OfferGroupDraftID: offerGroupDraftID,
		MultiUnitGroupID:  multiUnitGroupID,

		LookingDealID:       lookingDealID,
		DiscussionDealID:    discussionDealID,
		ConfirmedDealID:     confirmedDealID,
		CompletedDealID:     completedDealID,
		CompletedDeal2ID:    completedDeal2ID,
		CompletedDeal3ID:    completedDeal3ID,
		CancelledDealID:     cancelledDealID,
		PendingFailedDealID: pendingFailedDealID,
		FailedDealID:        failedDealID,
		JoinDealID:          joinDealID,

		DirectChatID: directChatID,
		DealChatID:   dealChatID,

		PendingReportID:  pendingReport.Id,
		AcceptedReportID: acceptedReport.Id,
		RejectedReportID: rejectedReport.Id,
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

func (c *SeedClient) createMutualReviews(
	ctx context.Context,
	dealID uuid.UUID,
	deal dealtypes.Deal,
	userA *seededUser,
	userB *seededUser,
	ratingForA int,
	ratingForB int,
	commentForA string,
	commentForB string,
) (dealtypes.Review, dealtypes.Review, error) {
	itemA, err := itemByAuthorID(deal.Items, userA.UserID)
	if err != nil {
		return dealtypes.Review{}, dealtypes.Review{}, fmt.Errorf("completed deal %s does not contain item for %s: %w", dealID, userA.Key, err)
	}

	itemB, err := itemByAuthorID(deal.Items, userB.UserID)
	if err != nil {
		return dealtypes.Review{}, dealtypes.Review{}, fmt.Errorf("completed deal %s does not contain item for %s: %w", dealID, userB.Key, err)
	}

	reviewForA, err := c.createDealItemReview(ctx, userB.Token, dealID, itemA.Id, dealtypes.CreateReviewRequest{
		Rating:  ratingForA,
		Comment: &commentForA,
	})
	if err != nil {
		return dealtypes.Review{}, dealtypes.Review{}, fmt.Errorf("review for %s item: %w", userA.Key, err)
	}

	reviewForB, err := c.createDealItemReview(ctx, userA.Token, dealID, itemB.Id, dealtypes.CreateReviewRequest{
		Rating:  ratingForB,
		Comment: &commentForB,
	})
	if err != nil {
		return dealtypes.Review{}, dealtypes.Review{}, fmt.Errorf("review for %s item: %w", userB.Key, err)
	}

	return reviewForA, reviewForB, nil
}

func createMultiPartyDeal(
	ctx context.Context,
	client *SeedClient,
	creator *seededUser,
	participants []*seededUser,
	offerIDs []uuid.UUID,
	name string,
	description string,
) (uuid.UUID, error) {
	if len(participants) != len(offerIDs) {
		return uuid.Nil, fmt.Errorf("participants count %d does not match offers count %d", len(participants), len(offerIDs))
	}

	before, err := client.listMyDeals(ctx, creator.Token)
	if err != nil {
		return uuid.Nil, fmt.Errorf("list deals before multi-party draft: %w", err)
	}

	offers := make([]dealtypes.OfferIDAndQuantity, len(offerIDs))
	for i, offerID := range offerIDs {
		offers[i] = dealtypes.OfferIDAndQuantity{OfferID: offerID, Quantity: 1}
	}

	draftID, err := client.createDraft(ctx, creator.Token, dealtypes.CreateDraftDealRequest{
		Name:        &name,
		Description: &description,
		Offers:      offers,
	})
	if err != nil {
		return uuid.Nil, err
	}

	for _, participant := range participants {
		if err := client.confirmDraft(ctx, participant.Token, draftID); err != nil {
			return uuid.Nil, fmt.Errorf("confirm multi-party draft by %s: %w", participant.Key, err)
		}
	}

	return client.waitForNewDeal(ctx, creator.Token, before)
}

func promoteMultiPartyDealToDiscussion(
	ctx context.Context,
	client *SeedClient,
	dealID uuid.UUID,
	viewer *seededUser,
	voters []*seededUser,
	receivers map[uuid.UUID]*seededUser,
) (dealtypes.Deal, error) {
	deal, err := client.getDealByID(ctx, viewer.Token, dealID)
	if err != nil {
		return dealtypes.Deal{}, err
	}

	for _, item := range deal.Items {
		receiver, ok := receivers[item.AuthorId]
		if !ok {
			return dealtypes.Deal{}, fmt.Errorf("receiver is not configured for item author %s", item.AuthorId)
		}

		if err := client.updateDealItem(ctx, receiver.Token, dealID, item.Id, dealtypes.UpdateDealItemRequest{
			ClaimReceiver: new(true),
		}); err != nil {
			return dealtypes.Deal{}, fmt.Errorf("claim receiver for item %s: %w", item.Id, err)
		}
	}

	for _, voter := range voters {
		if err := client.changeDealStatus(ctx, voter.Token, dealID, dealtypes.Discussion); err != nil {
			return dealtypes.Deal{}, fmt.Errorf("discussion vote by %s: %w", voter.Key, err)
		}
	}

	return client.getDealByID(ctx, viewer.Token, dealID)
}

func itemByAuthorID(items []dealtypes.Item, authorID uuid.UUID) (dealtypes.Item, error) {
	for _, item := range items {
		if item.AuthorId == authorID {
			return item, nil
		}
	}

	return dealtypes.Item{}, fmt.Errorf("item for author %s not found", authorID)
}

func itemByOfferID(items []dealtypes.Item, offerID uuid.UUID) (dealtypes.Item, error) {
	for _, item := range items {
		if item.OfferId != nil && *item.OfferId == offerID {
			return item, nil
		}
	}

	return dealtypes.Item{}, fmt.Errorf("item for offer %s not found", offerID)
}
