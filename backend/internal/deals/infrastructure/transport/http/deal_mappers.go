package http

import (
	"barter-port/contracts/openapi/deals/types"
	"barter-port/internal/deals/domain/htypes"

	openapi_types "github.com/oapi-codegen/runtime/types"
)

func mapDealIDWithParticipantIDsToDTO(d htypes.DealIDWithParticipantIDs) types.GetDealsResponseItem {
	participants := make([]openapi_types.UUID, 0, len(d.ParticipantIDs))
	for _, participantID := range d.ParticipantIDs {
		participants = append(participants, participantID)
	}

	return types.GetDealsResponseItem{
		Id:           d.ID,
		Participants: participants,
	}
}

func mapDealIDsWithParticipantIDsToDTO(deals []htypes.DealIDWithParticipantIDs) types.GetDealsResponse {
	result := make(types.GetDealsResponse, 0, len(deals))
	for _, deal := range deals {
		result = append(result, mapDealIDWithParticipantIDsToDTO(deal))
	}
	return result
}
