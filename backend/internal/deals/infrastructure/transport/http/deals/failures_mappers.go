package deals

import (
	"barter-port/contracts/openapi/deals/types"
	"barter-port/internal/deals/domain/htypes"
)

func mapFailureVotesToDTO(votes []htypes.FailureVote) types.FailureVotesResponse {
	result := make(types.FailureVotesResponse, 0, len(votes))
	for _, vote := range votes {
		result = append(result, types.FailureVotesResponseItem{
			UserId: vote.UserID,
			Vote:   vote.Vote,
		})
	}

	return result
}

func mapFailureMaterialsToDTO(materials htypes.FailureMaterials) types.FailureMaterialResponse {
	return types.FailureMaterialResponse{
		Deal:   mapDealToDTO(materials.Deal),
		ChatId: materials.ChatID,
	}
}

func mapFailureRecordToDTO(record htypes.FailureRecord) types.DealFailureModeratorResolution {
	return types.DealFailureModeratorResolution{
		UserId:           record.UserID,
		Confirmed:        record.ConfirmedByAdmin,
		PunishmentPoints: record.PunishmentPoints,
		Comment:          record.AdminComment,
	}
}
