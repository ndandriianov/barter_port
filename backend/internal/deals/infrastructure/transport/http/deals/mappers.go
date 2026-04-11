package deals

import (
	"barter-port/contracts/openapi/deals/types"
	"barter-port/internal/deals/domain/enums"
	"barter-port/internal/deals/infrastructure/transport/http/common"
	"fmt"
	"sort"

	"github.com/google/uuid"
)

func mapDealStatusFromDTO(s types.DealStatus) (enums.DealStatus, error) {
	switch s {
	case types.LookingForParticipants:
		return enums.DealStatusLookingForParticipants, nil
	case types.Discussion:
		return enums.DealStatusDiscussion, nil
	case types.Confirmed:
		return enums.DealStatusConfirmed, nil
	case types.Completed:
		return enums.DealStatusCompleted, nil
	case types.Cancelled:
		return enums.DealStatusCancelled, nil
	case types.Failed:
		return enums.DealStatusFailed, nil
	default:
		return 0, fmt.Errorf("unknown deal status: %s", s)
	}
}

func mapStatusVotesToDTO(votes map[uuid.UUID]enums.DealStatus) types.GetDealStatusVotesResponse {
	userIDs := make([]uuid.UUID, 0, len(votes))
	for userID := range votes {
		userIDs = append(userIDs, userID)
	}

	sort.Slice(userIDs, func(i, j int) bool {
		return userIDs[i].String() < userIDs[j].String()
	})

	result := make(types.GetDealStatusVotesResponse, 0, len(votes))
	for _, userID := range userIDs {
		result = append(result, types.GetDealStatusVotesResponseItem{
			UserId: userID,
			Vote:   common.MapDealStatusToDTO(votes[userID]),
		})
	}

	return result
}
