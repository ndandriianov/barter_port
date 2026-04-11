package failures

import (
	"barter-port/contracts/openapi/deals/types"
	"barter-port/internal/deals/domain"
	"barter-port/internal/deals/domain/enums"
	"barter-port/internal/deals/domain/htypes"

	openapitypes "github.com/oapi-codegen/runtime/types"
)

func mapDealIDWithParticipantIDsToDTO(d htypes.DealIDWithParticipantIDs) types.GetDealsResponseItem {
	participants := make([]openapitypes.UUID, 0, len(d.ParticipantIDs))
	for _, participantID := range d.ParticipantIDs {
		participants = append(participants, participantID)
	}

	s := mapDealStatusToDTO(d.Status)
	return types.GetDealsResponseItem{
		Id:           d.ID,
		Name:         d.Name,
		Status:       &s,
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

func mapDealStatusToDTO(status enums.DealStatus) types.DealStatus {
	switch status {
	case enums.DealStatusLookingForParticipants:
		return types.LookingForParticipants
	case enums.DealStatusDiscussion:
		return types.Discussion
	case enums.DealStatusConfirmed:
		return types.Confirmed
	case enums.DealStatusCompleted:
		return types.Completed
	case enums.DealStatusCancelled:
		return types.Cancelled
	case enums.DealStatusFailed:
		return types.Failed
	default:
		return ""
	}
}

func mapDealToDTO(deal domain.Deal) types.Deal {
	itemsDTO := make([]types.Item, len(deal.Items))
	for i, item := range deal.Items {
		itemsDTO[i] = item.ToDTO()
	}

	return types.Deal{
		Id:           deal.ID,
		Name:         deal.Name,
		Description:  deal.Description,
		CreatedAt:    deal.CreatedAt,
		UpdatedAt:    deal.UpdatedAt,
		Status:       mapDealStatusToDTO(deal.Status),
		Items:        itemsDTO,
		Participants: deal.Participants,
	}
}

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
