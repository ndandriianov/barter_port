package joins

import (
	"barter-port/contracts/openapi/deals/types"
	"barter-port/internal/deals/domain/htypes"

	openapitypes "github.com/oapi-codegen/runtime/types"
)

func mapJoinRequestToDTO(item htypes.JoinRequestWithVoters) types.GetDealJoinRequestsResponseItem {
	voters := make([]openapitypes.UUID, 0, len(item.VoterIDs))
	for _, voterID := range item.VoterIDs {
		voters = append(voters, voterID)
	}

	return types.GetDealJoinRequestsResponseItem{
		UserId: item.UserID,
		DealId: item.DealID,
		Voters: voters,
	}
}

func mapJoinRequestsToDTO(items []htypes.JoinRequestWithVoters) types.GetDealJoinRequestsResponse {
	result := make(types.GetDealJoinRequestsResponse, 0, len(items))
	for _, item := range items {
		result = append(result, mapJoinRequestToDTO(item))
	}
	return result
}
