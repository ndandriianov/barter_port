package seed_demo

import (
	dealtypes "barter-port/contracts/openapi/deals/types"
	usertypes "barter-port/contracts/openapi/users/types"
	"context"
	"fmt"

	"github.com/google/uuid"
)

func RunSeed(ctx context.Context, client *SeedClient, cfg SeedConfig) (*SeedSummary, error) {
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
		{
			Key:      "eva",
			Name:     "Eva Nikiforova",
			Bio:      "Увлекаюсь акварелью и меняю художественные материалы.",
			Email:    "eva.demo@barterport.local",
			Password: cfg.Password,
			Avatar:   cfg.AvatarBaseURL + "/eva.png",
		},
		{
			Key:      "fedor",
			Name:     "Fedor Gromov",
			Bio:      "Собираю виниловые пластинки и советскую электронику.",
			Email:    "fedor.demo@barterport.local",
			Password: cfg.Password,
			Avatar:   cfg.AvatarBaseURL + "/fedor.png",
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
	eva := &users[4]
	fedor := &users[5]

	adminToken, err := client.login(ctx, cfg.AdminEmail, cfg.AdminPassword)
	if err != nil {
		return nil, fmt.Errorf("login as admin: %w", err)
	}

	// ── Offers ──────────────────────────────────────────────────────────────

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
		{
			Key:         "thermos",
			Name:        "Термос туристический",
			Description: "0.9 л, нержавейка, держит тепло 12 часов.",
			Type:        dealtypes.Good,
			Action:      dealtypes.Give,
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

	evaOffers, err := client.createOffers(ctx, eva, []offerSpec{
		{
			Key:         "watercolors",
			Name:        "Акварельные краски",
			Description: "Набор из 24 цветов, почти новый.",
			Type:        dealtypes.Good,
			Action:      dealtypes.Give,
		},
		{
			Key:         "canvas-set",
			Name:        "Набор холстов",
			Description: "5 холстов 30×40 см, загрунтованы.",
			Type:        dealtypes.Good,
			Action:      dealtypes.Give,
		},
		{
			Key:         "art-lesson",
			Name:        "Урок по акварели",
			Description: "Индивидуальный урок 1.5 часа онлайн или очно.",
			Type:        dealtypes.Service,
			Action:      dealtypes.Give,
		},
		{
			Key:         "want-brushes",
			Name:        "Ищу кисти для акрила",
			Description: "Нужен набор кистей разных размеров.",
			Type:        dealtypes.Good,
			Action:      dealtypes.Take,
		},
		{
			Key:         "sketchbooks",
			Name:        "Блокноты для скетчей",
			Description: "Три блокнота A5 с плотной бумагой.",
			Type:        dealtypes.Good,
			Action:      dealtypes.Give,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("create Eva offers: %w", err)
	}

	fedorOffers, err := client.createOffers(ctx, fedor, []offerSpec{
		{
			Key:         "vinyl-player",
			Name:        "Виниловый проигрыватель",
			Description: "Рабочий, советского производства, с иглой.",
			Type:        dealtypes.Good,
			Action:      dealtypes.Give,
		},
		{
			Key:         "old-camera",
			Name:        "Плёночная фотокамера Зенит-11",
			Description: "С объективом Helios-44, рабочее состояние.",
			Type:        dealtypes.Good,
			Action:      dealtypes.Give,
		},
		{
			Key:         "want-records",
			Name:        "Ищу виниловые пластинки",
			Description: "Интересует jazz и классика, любое состояние.",
			Type:        dealtypes.Good,
			Action:      dealtypes.Take,
		},
		{
			Key:         "headphones",
			Name:        "Советские наушники ТДС-3",
			Description: "Изодинамические, рабочие, в оригинальном чехле.",
			Type:        dealtypes.Good,
			Action:      dealtypes.Give,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("create Fedor offers: %w", err)
	}

	// ── Offer groups ─────────────────────────────────────────────────────────

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

	// Eva's multi-unit group: unit 1 has two offers (OR-choice), unit 2 is mandatory
	multiUnitGroupID, err := client.createOfferGroup(ctx, eva.Token, offerGroupRequest{
		Name:        new("Художественные материалы или урок"),
		Description: new("Можно забрать краски или записаться на урок — плюс блокноты в любом случае."),
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

	// ── Deals ────────────────────────────────────────────────────────────────

	// A — LookingForParticipants: Fedor's vinyl player, open deal
	lookingDealID, err := client.createLookingDeal(ctx, fedor, fedorOffers["vinyl-player"],
		"Виниловый проигрыватель ищет нового хозяина",
		"Открытая сделка — жду партнёра с интересным предложением.")
	if err != nil {
		return nil, fmt.Errorf("create looking deal: %w", err)
	}

	// B — Discussion: Alice ↔ Bob
	discussionDealID, err := client.createTwoPartyDeal(ctx, alice, bob, aliceOffers["storage-boxes"], bobOffers["tool-kit"],
		"Обмен для домашнего кабинета",
		"Alice и Bob обсуждают обмен коробов для хранения на набор инструментов.")
	if err != nil {
		return nil, fmt.Errorf("create discussion deal: %w", err)
	}

	if _, err := client.promoteDealToDiscussion(ctx, discussionDealID, alice, bob); err != nil {
		return nil, fmt.Errorf("promote discussion deal: %w", err)
	}

	// C — Confirmed: Clara ↔ Eva
	confirmedDealID, err := client.createTwoPartyDeal(ctx, clara, eva, claraOffers["english-session"], evaOffers["art-lesson"],
		"Обмен уроками: английский на акварель",
		"Clara и Eva меняются уроками — договорились, ждут встречи.")
	if err != nil {
		return nil, fmt.Errorf("create confirmed deal: %w", err)
	}

	if _, err := client.promoteDealToDiscussion(ctx, confirmedDealID, clara, eva); err != nil {
		return nil, fmt.Errorf("promote confirmed deal to discussion: %w", err)
	}

	if err := client.changeDealStatus(ctx, clara.Token, confirmedDealID, dealtypes.Confirmed); err != nil {
		return nil, fmt.Errorf("confirm deal by clara: %w", err)
	}
	if err := client.changeDealStatus(ctx, eva.Token, confirmedDealID, dealtypes.Confirmed); err != nil {
		return nil, fmt.Errorf("confirm deal by eva: %w", err)
	}

	// D — Completed: Clara ↔ Dan (5★/5★)
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
		{Token: dan.Token, Content: "Да, договорились! Когда и где встретимся?"},
	}); err != nil {
		return nil, fmt.Errorf("send messages for completed deal: %w", err)
	}

	if err := client.completeTwoPartyDeal(ctx, completedDealID, clara, dan); err != nil {
		return nil, fmt.Errorf("complete deal: %w", err)
	}

	if err := client.createMutualReviews(ctx, completedDealID, completedDeal, clara, dan, 5, 5,
		"Все прошло четко: договорились быстро и получили именно то, что ожидали.",
		"Хорошая коммуникация и удобная передача вещи.",
	); err != nil {
		return nil, fmt.Errorf("create reviews for completed deal: %w", err)
	}

	// E — Completed: Alice ↔ Bob, extra offers (4★/4★)
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

	if err := client.createMutualReviews(ctx, completedDeal2ID, completedDeal2, alice, bob, 4, 4,
		"Всё как договаривались, чуть задержались с передачей.",
		"Хороший обмен, термос в отличном состоянии.",
	); err != nil {
		return nil, fmt.Errorf("create reviews for completed deal 2: %w", err)
	}

	// F — Completed: Eva ↔ Fedor (3★/5★)
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

	// Eva rates Fedor's headphones 3★, Fedor rates Eva's canvas 5★
	if err := client.createMutualReviews(ctx, completedDeal3ID, completedDeal3, eva, fedor, 3, 5,
		"Наушники рабочие, но состояние хуже, чем на фото.",
		"Холсты отличного качества, очень доволен обменом!",
	); err != nil {
		return nil, fmt.Errorf("create reviews for completed deal 3: %w", err)
	}

	// G — Cancelled: Bob ↔ Dan
	cancelledDealID, err := client.createAndCancelDeal(ctx, bob, dan, bobOffers["bike-bag"], danOffers["board-game-sleeves"],
		"Велосумка на протекторы",
		"Bob и Dan не смогли договориться о времени встречи.")
	if err != nil {
		return nil, fmt.Errorf("create cancelled deal: %w", err)
	}

	// H — Failed: Alice ↔ Fedor
	failedDealID, err := client.createTwoPartyDeal(ctx, alice, fedor, aliceOffers["lamp"], fedorOffers["old-camera"],
		"Лампа на фотоаппарат",
		"Alice и Fedor пытались обменять лампу на фотокамеру.")
	if err != nil {
		return nil, fmt.Errorf("create failed deal: %w", err)
	}

	if _, err := client.promoteDealToDiscussion(ctx, failedDealID, alice, fedor); err != nil {
		return nil, fmt.Errorf("promote failed deal to discussion: %w", err)
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
	if err := client.voteForFailure(ctx, fedor.Token, failedDealID, alice.UserID); err != nil {
		return nil, fmt.Errorf("fedor vote for failure: %w", err)
	}

	punishmentPoints := 10
	comment := "Сделка провалена по вине Fedor: товар не соответствовал описанию."
	if err := client.moderatorResolutionForFailure(ctx, adminToken, failedDealID, dealtypes.ModeratorResolutionForFailureRequest{
		Confirmed:        true,
		UserId:           &fedor.UserID,
		PunishmentPoints: &punishmentPoints,
		Comment:          &comment,
	}); err != nil {
		return nil, fmt.Errorf("moderator resolution for failure: %w", err)
	}

	// I — LookingForParticipants + join request: Bob opens deal, Eva requests to join
	joinDealID, err := client.createLookingDeal(ctx, bob, bobOffers["coffee-beans"],
		"Ищу зерновой кофе",
		"Открытая сделка — готов обменять на что-то интересное.")
	if err != nil {
		return nil, fmt.Errorf("create join deal: %w", err)
	}

	if err := client.requestJoinDeal(ctx, eva.Token, joinDealID); err != nil {
		return nil, fmt.Errorf("eva request join deal: %w", err)
	}

	if err := client.processJoinRequest(ctx, bob.Token, joinDealID, eva.UserID, true); err != nil {
		return nil, fmt.Errorf("bob process join request: %w", err)
	}

	// ── Offer reports ─────────────────────────────────────────────────────────

	// Pending: Clara reports Dan's card sleeves
	pendingReport, err := client.createOfferReport(ctx, clara.Token, danOffers["board-game-sleeves"],
		"Объявление содержит недостоверное описание товара — количество протекторов не соответствует.")
	if err != nil {
		return nil, fmt.Errorf("create pending report: %w", err)
	}

	// Accepted: Alice reports Fedor's want-records → admin accepts → offer hidden, rep penalty
	acceptedReport, err := client.createOfferReport(ctx, alice.Token, fedorOffers["want-records"],
		"Объявление размещено повторно и нарушает правила площадки — дубликат другого оффера.")
	if err != nil {
		return nil, fmt.Errorf("create accepted report: %w", err)
	}

	acceptComment := "Повторная публикация — нарушение правил."
	if _, err := client.resolveOfferReport(ctx, adminToken, acceptedReport.Id, true, &acceptComment); err != nil {
		return nil, fmt.Errorf("resolve accepted report: %w", err)
	}

	// Rejected: Dan reports Eva's canvas-set → admin rejects
	rejectedReport, err := client.createOfferReport(ctx, dan.Token, evaOffers["canvas-set"],
		"Подозрительная цена на обмен — похоже на коммерческое объявление.")
	if err != nil {
		return nil, fmt.Errorf("create rejected report: %w", err)
	}

	rejectComment := "Жалоба не подтверждена: объявление соответствует правилам площадки."
	if _, err := client.resolveOfferReport(ctx, adminToken, rejectedReport.Id, false, &rejectComment); err != nil {
		return nil, fmt.Errorf("resolve rejected report: %w", err)
	}

	// ── Subscriptions ─────────────────────────────────────────────────────────

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

	// ── Chats ─────────────────────────────────────────────────────────────────

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

	// Chat for confirmed deal (Clara ↔ Eva)
	confirmedDealChatID, err := client.waitForDealChat(ctx, clara.Token, confirmedDealID)
	if err != nil {
		return nil, fmt.Errorf("wait for confirmed deal chat: %w", err)
	}

	if err := client.sendChatMessages(ctx, confirmedDealChatID, []chatMessage{
		{Token: clara.Token, Content: "Когда тебе удобно провести урок? Я свободна в субботу."},
		{Token: eva.Token, Content: "Отлично, договорились на субботу в 12:00!"},
	}); err != nil {
		return nil, fmt.Errorf("send confirmed deal chat messages: %w", err)
	}

	// Direct chat Clara ↔ Dan
	claraDanChatID, err := client.createDirectChat(ctx, clara.Token, dan.UserID)
	if err != nil {
		return nil, fmt.Errorf("create direct chat clara-dan: %w", err)
	}

	if err := client.sendChatMessages(ctx, claraDanChatID, []chatMessage{
		{Token: clara.Token, Content: "Дан, у тебя случайно нет чего-то интересного для обмена?"},
		{Token: dan.Token, Content: "Есть пара вещей, посмотри мои объявления — напиши если что-то подойдёт."},
	}); err != nil {
		return nil, fmt.Errorf("send clara-dan chat messages: %w", err)
	}

	_ = claraDanChatID

	// ── Summary ───────────────────────────────────────────────────────────────

	summary := &SeedSummary{
		OfferGroupID:      offerGroupID,
		OfferGroupDraftID: offerGroupDraftID,
		MultiUnitGroupID:  multiUnitGroupID,

		LookingDealID:    lookingDealID,
		DiscussionDealID: discussionDealID,
		ConfirmedDealID:  confirmedDealID,
		CompletedDealID:  completedDealID,
		CompletedDeal2ID: completedDeal2ID,
		CompletedDeal3ID: completedDeal3ID,
		CancelledDealID:  cancelledDealID,
		FailedDealID:     failedDealID,
		JoinDealID:       joinDealID,

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

// createMutualReviews creates cross-reviews between userA and userB for a completed deal.
// ratingForA is the rating userB gives to userA's item; ratingForB is the rating userA gives to userB's item.
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
) error {
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
		Rating:  ratingForA,
		Comment: &commentForA,
	}); err != nil {
		return fmt.Errorf("review for %s item: %w", userA.Key, err)
	}

	if err := c.createDealItemReview(ctx, userA.Token, dealID, itemB, dealtypes.CreateReviewRequest{
		Rating:  ratingForB,
		Comment: &commentForB,
	}); err != nil {
		return fmt.Errorf("review for %s item: %w", userB.Key, err)
	}

	return nil
}
