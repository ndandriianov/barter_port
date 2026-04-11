package common

import (
	"barter-port/contracts/openapi/deals/types"
	"barter-port/internal/deals/domain"
	"barter-port/internal/deals/domain/enums"
	"barter-port/internal/deals/domain/htypes"

	"github.com/google/uuid"
	openapitypes "github.com/oapi-codegen/runtime/types"
)

func MapUUIDs(ids []uuid.UUID) []openapitypes.UUID {
	result := make([]openapitypes.UUID, 0, len(ids))
	for _, id := range ids {
		result = append(result, id)
	}

	return result
}

func MapDealIDWithParticipantIDsToDTO(d htypes.DealIDWithParticipantIDs) types.GetDealsResponseItem {
	participants := MapUUIDs(d.ParticipantIDs)

	return types.GetDealsResponseItem{
		Id:           d.ID,
		Name:         d.Name,
		Status:       new(MapDealStatusToDTO(d.Status)),
		Participants: participants,
	}
}

func MapDealIDsWithParticipantIDsToDTO(deals []htypes.DealIDWithParticipantIDs) types.GetDealsResponse {
	result := make(types.GetDealsResponse, 0, len(deals))
	for _, deal := range deals {
		result = append(result, MapDealIDWithParticipantIDsToDTO(deal))
	}
	return result
}

func MapDealStatusToDTO(status enums.DealStatus) types.DealStatus {
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

func MapDealToDTO(deal domain.Deal) types.Deal {
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
		Status:       MapDealStatusToDTO(deal.Status),
		Items:        itemsDTO,
		Participants: deal.Participants,
	}
}
