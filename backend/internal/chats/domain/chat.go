package domain

import (
	"time"

	"github.com/google/uuid"
)

type Chat struct {
	ID           uuid.UUID
	DealID       *uuid.UUID
	Participants []ChatParticipant
	CreatedAt    time.Time
	UpdatedAt    *time.Time
}

func (c Chat) GetParticipantIdsToString() []string {
	ids := make([]string, len(c.Participants))
	for i, p := range c.Participants {
		ids[i] = p.ID.String()
	}
	return ids
}

type ChatParticipant struct {
	ID   uuid.UUID
	Name *string
}

func NewChatParticipantsWithoutNames(ids []uuid.UUID) []ChatParticipant {
	participants := make([]ChatParticipant, len(ids))
	for i, id := range ids {
		participants[i] = ChatParticipant{
			ID: id,
		}
	}

	return participants
}

type Message struct {
	ID        uuid.UUID
	ChatID    uuid.UUID
	SenderID  uuid.UUID
	Content   string
	CreatedAt time.Time
	UpdatedAt *time.Time
}
