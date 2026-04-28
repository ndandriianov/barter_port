package repository

import (
	"barter-port/internal/chats/domain"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/net/context"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// CreateChat creates a new chat with participants. If deal_id is provided, prevents duplicate deal chats.
func (r *Repository) CreateChat(ctx context.Context, dealID *uuid.UUID, participantIDs []uuid.UUID) (*domain.Chat, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	var chat domain.Chat

	if dealID != nil {
		// Check if chat for this deal already exists
		var existingID uuid.UUID
		err = tx.QueryRow(ctx, `SELECT id FROM chats WHERE deal_id = $1`, dealID).Scan(&existingID)
		if err == nil {
			return r.GetChatByID(ctx, existingID)
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("check existing deal chat: %w", err)
		}

		err = tx.QueryRow(ctx,
			`INSERT INTO chats (deal_id) VALUES ($1) RETURNING id, deal_id, created_at, updated_at`,
			dealID,
		).Scan(&chat.ID, &chat.DealID, &chat.CreatedAt, &chat.UpdatedAt)
	} else {
		if len(participantIDs) == 2 {
			var existingID uuid.UUID
			err = tx.QueryRow(ctx, `
				SELECT c.id
				FROM chats c
				JOIN chat_participants cp ON cp.chat_id = c.id
				WHERE c.deal_id IS NULL
				GROUP BY c.id
				HAVING COUNT(*) = 2
				   AND COUNT(*) FILTER (WHERE cp.user_id = $1 OR cp.user_id = $2) = 2
			`, participantIDs[0], participantIDs[1]).Scan(&existingID)
			if err == nil {
				return nil, domain.ErrChatAlreadyExists
			}
			if !errors.Is(err, pgx.ErrNoRows) {
				return nil, fmt.Errorf("check existing direct chat: %w", err)
			}
		}

		err = tx.QueryRow(ctx,
			`INSERT INTO chats DEFAULT VALUES RETURNING id, deal_id, created_at, updated_at`,
		).Scan(&chat.ID, &chat.DealID, &chat.CreatedAt, &chat.UpdatedAt)
	}
	if err != nil {
		return nil, fmt.Errorf("insert chat: %w", err)
	}

	for _, uid := range participantIDs {
		_, err = tx.Exec(ctx,
			`INSERT INTO chat_participants (chat_id, user_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
			chat.ID, uid,
		)
		if err != nil {
			return nil, fmt.Errorf("insert participant: %w", err)
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	chat.Participants = domain.NewChatParticipantsWithoutNames(participantIDs)
	return &chat, nil
}

// GetChatByID returns a chat by its ID including participants.
func (r *Repository) GetChatByID(ctx context.Context, chatID uuid.UUID) (*domain.Chat, error) {
	var chat domain.Chat
	err := r.db.QueryRow(ctx,
		`SELECT id, deal_id, created_at, updated_at FROM chats WHERE id = $1`,
		chatID,
	).Scan(&chat.ID, &chat.DealID, &chat.CreatedAt, &chat.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrChatNotFound
		}
		return nil, fmt.Errorf("query chat: %w", err)
	}

	participants, err := r.getChatParticipants(ctx, chatID)
	if err != nil {
		return nil, err
	}
	chat.Participants = domain.NewChatParticipantsWithoutNames(participants)

	return &chat, nil
}

// GetDealChatID returns the ID of the chat associated with the deal.
func (r *Repository) GetDealChatID(ctx context.Context, dealID uuid.UUID) (uuid.UUID, error) {
	var chatID uuid.UUID
	err := r.db.QueryRow(ctx, `SELECT id FROM chats WHERE deal_id = $1`, dealID).Scan(&chatID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, domain.ErrChatNotFound
		}
		return uuid.Nil, fmt.Errorf("query deal chat id: %w", err)
	}

	return chatID, nil
}

// ListChatsForUser returns all chats where the given user is a participant.
func (r *Repository) ListChatsForUser(ctx context.Context, userID uuid.UUID) ([]domain.Chat, error) {
	rows, err := r.db.Query(ctx, `
		SELECT c.id, c.deal_id, c.created_at, c.updated_at
		FROM chats c
		JOIN chat_participants cp ON cp.chat_id = c.id
		WHERE cp.user_id = $1
		ORDER BY c.created_at DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("query chats: %w", err)
	}
	defer rows.Close()

	var chats []domain.Chat
	for rows.Next() {
		var c domain.Chat
		if err = rows.Scan(&c.ID, &c.DealID, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan chat: %w", err)
		}
		chats = append(chats, c)
	}

	for i := range chats {
		participants, err := r.getChatParticipants(ctx, chats[i].ID)
		if err != nil {
			return nil, err
		}
		chats[i].Participants = domain.NewChatParticipantsWithoutNames(participants)
	}

	return chats, nil
}

// IsParticipant checks if a user is a participant of a chat.
func (r *Repository) IsParticipant(ctx context.Context, chatID, userID uuid.UUID) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM chat_participants WHERE chat_id = $1 AND user_id = $2)`,
		chatID, userID,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check participant: %w", err)
	}
	return exists, nil
}

// SendMessage inserts a new message into a chat.
func (r *Repository) SendMessage(ctx context.Context, chatID, senderID uuid.UUID, content string) (*domain.Message, error) {
	var msg domain.Message
	err := r.db.QueryRow(ctx,
		`INSERT INTO messages (chat_id, sender_id, content) VALUES ($1, $2, $3)
		 RETURNING id, chat_id, sender_id, content, created_at, updated_at`,
		chatID, senderID, content,
	).Scan(&msg.ID, &msg.ChatID, &msg.SenderID, &msg.Content, &msg.CreatedAt, &msg.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("insert message: %w", err)
	}
	return &msg, nil
}

// GetMessages returns messages in a chat. If after is non-zero, returns only messages created after that time.
func (r *Repository) GetMessages(ctx context.Context, chatID uuid.UUID, after *time.Time) ([]domain.Message, error) {
	var (
		rows pgx.Rows
		err  error
	)

	if after != nil {
		rows, err = r.db.Query(ctx, `
			SELECT id, chat_id, sender_id, content, created_at, updated_at
			FROM messages
			WHERE chat_id = $1 AND created_at > $2
			ORDER BY created_at ASC
		`, chatID, after)
	} else {
		rows, err = r.db.Query(ctx, `
			SELECT id, chat_id, sender_id, content, created_at, updated_at
			FROM messages
			WHERE chat_id = $1
			ORDER BY created_at ASC
		`, chatID)
	}
	if err != nil {
		return nil, fmt.Errorf("query messages: %w", err)
	}
	defer rows.Close()

	var msgs []domain.Message
	for rows.Next() {
		var m domain.Message
		if err = rows.Scan(&m.ID, &m.ChatID, &m.SenderID, &m.Content, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan message: %w", err)
		}
		msgs = append(msgs, m)
	}

	return msgs, nil
}

func (r *Repository) getChatParticipants(ctx context.Context, chatID uuid.UUID) ([]uuid.UUID, error) {
	rows, err := r.db.Query(ctx,
		`SELECT user_id FROM chat_participants WHERE chat_id = $1`,
		chatID,
	)
	if err != nil {
		return nil, fmt.Errorf("query participants: %w", err)
	}
	defer rows.Close()

	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err = rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan participant: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, nil
}
