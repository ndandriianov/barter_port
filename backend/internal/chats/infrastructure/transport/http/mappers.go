package http

import (
	"barter-port/contracts/openapi/chats/types"
	"barter-port/internal/chats/domain"
)

func mapChatToResp(c *domain.Chat) types.Chat {
	resp := types.Chat{
		Id:           c.ID,
		Participants: c.Participants,
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
