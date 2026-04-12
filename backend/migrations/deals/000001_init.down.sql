-- Индексы
DROP INDEX IF EXISTS deal_reviews_item_created_at_idx;
DROP INDEX IF EXISTS deal_reviews_offer_created_at_idx;
DROP INDEX IF EXISTS deal_reviews_unique_offer_item_context_idx;
DROP INDEX IF EXISTS deal_reviews_unique_item_only_context_idx;
DROP INDEX IF EXISTS deal_reviews_unique_offer_only_context_idx;

-- Таблицы (в порядке зависимостей)
DROP TABLE IF EXISTS deal_reviews;
DROP TABLE IF EXISTS deal_failures;
DROP TABLE IF EXISTS join_requests_votes;
DROP TABLE IF EXISTS join_requests;
DROP TABLE IF EXISTS participants;
DROP TABLE IF EXISTS items;
DROP TABLE IF EXISTS deals;

-- Тип (после удаления зависимых таблиц)
DROP TYPE IF EXISTS deal_status;

DROP TABLE IF EXISTS draft_deal_offers;
DROP TABLE IF EXISTS draft_deals;

DROP TABLE IF EXISTS unit_offers;
DROP TABLE IF EXISTS offer_group_units;
DROP TABLE IF EXISTS offer_groups;

DROP TABLE IF EXISTS offers;