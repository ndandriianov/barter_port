package joins

import (
	"barter-port/contracts/openapi/deals/types"
	"barter-port/internal/deals/domain/htypes"
	"barter-port/internal/deals/infrastructure/transport/http/common"
)

func mapJoinRequestToDTO(item htypes.JoinRequestWithVoters) types.GetDealJoinRequestsResponseItem {
	return types.GetDealJoinRequestsResponseItem{
		UserId: item.UserID,
		DealId: item.DealID,
		Voters: common.MapUUIDs(item.VoterIDs),
	}
}

func mapJoinRequestsToDTO(items []htypes.JoinRequestWithVoters) types.GetDealJoinRequestsResponse {
	result := make(types.GetDealJoinRequestsResponse, 0, len(items))
	for _, item := range items {
		result = append(result, mapJoinRequestToDTO(item))
	}
	return result
}
