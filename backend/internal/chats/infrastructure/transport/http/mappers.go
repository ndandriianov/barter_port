package http

import (
	"barter-port/contracts/openapi/chats/types"
	"barter-port/internal/chats/domain"
)

func mapChatToResp(c *domain.Chat) types.Chat {
	participants := make([]types.Participant, len(c.Participants))
	for i, p := range c.Participants {
		participants[i] = types.Participant{
			UserId:   p.ID,
			UserName: p.Name,
		}
	}

	resp := types.Chat{
		Id:           c.ID,
		Participants: participants,
		CreatedAt:    c.CreatedAt,
		DealId:       c.DealID,
	}

	return resp
}

func mapMessageToResp(m *domain.Message) types.Message {
	return types.Message{
		Id:        m.ID,
		ChatId:    m.ChatID,
		SenderId:  m.SenderID,
		Content:   m.Content,
		CreatedAt: m.CreatedAt,
	}
}
