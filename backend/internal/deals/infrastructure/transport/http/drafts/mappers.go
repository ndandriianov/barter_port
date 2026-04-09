package drafts

import (
	"barter-port/contracts/openapi/deals/types"
	"barter-port/internal/deals/domain/htypes"

	openapitypes "github.com/oapi-codegen/runtime/types"
)

func mapDraftIDWithAuthorIDsToDTO(d htypes.DraftIDWithAuthorIDs) types.GetMyDraftDealsResponseItem {
	participants := make([]openapitypes.UUID, 0, len(d.ParticipantIDs))
	for _, participantID := range d.ParticipantIDs {
		participants = append(participants, participantID)
	}

	return types.GetMyDraftDealsResponseItem{
		Id:           d.ID,
		Name:         d.Name,
		Participants: participants,
	}
}

func mapDraftIDsWithAuthorIDsToDTO(drafts []htypes.DraftIDWithAuthorIDs) types.GetMyDraftDealsResponse {
	result := make(types.GetMyDraftDealsResponse, 0, len(drafts))
	for _, draft := range drafts {
		result = append(result, mapDraftIDWithAuthorIDsToDTO(draft))
	}
	return result
}
