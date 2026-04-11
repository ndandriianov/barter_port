package drafts

import (
	"barter-port/contracts/openapi/deals/types"
	"barter-port/internal/deals/domain/htypes"
	"barter-port/internal/deals/infrastructure/transport/http/common"
)

func mapDraftIDWithAuthorIDsToDTO(d htypes.DraftIDWithAuthorIDs) types.GetMyDraftDealsResponseItem {
	return types.GetMyDraftDealsResponseItem{
		Id:           d.ID,
		Name:         d.Name,
		Participants: common.MapUUIDs(d.ParticipantIDs),
	}
}

func mapDraftIDsWithAuthorIDsToDTO(drafts []htypes.DraftIDWithAuthorIDs) types.GetMyDraftDealsResponse {
	result := make(types.GetMyDraftDealsResponse, 0, len(drafts))
	for _, draft := range drafts {
		result = append(result, mapDraftIDWithAuthorIDsToDTO(draft))
	}
	return result
}
