package deals

import (
	"barter-port/contracts/openapi/deals/types"
	"barter-port/internal/deals/domain"
	"barter-port/internal/deals/domain/enums"
	"barter-port/internal/deals/domain/htypes"
	"fmt"

	openapitypes "github.com/oapi-codegen/runtime/types"
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

func mapDealIDWithParticipantIDsToDTO(d htypes.DealIDWithParticipantIDs) types.GetDealsResponseItem {
	participants := make([]openapitypes.UUID, 0, len(d.ParticipantIDs))
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

func mapDealToDTO(deal domain.Deal) types.Deal {
	itemsDTO := make([]types.Item, len(deal.Items))
	for i, item := range deal.Items {
		itemsDTO[i] = item.ToDTO()
	}

	var status types.DealStatus
	switch deal.Status {
	case enums.DealStatusLookingForParticipants:
		status = types.LookingForParticipants
	case enums.DealStatusDiscussion:
		status = types.Discussion
	case enums.DealStatusConfirmed:
		status = types.Confirmed
	case enums.DealStatusCompleted:
		status = types.Completed
	case enums.DealStatusCancelled:
		status = types.Cancelled
	case enums.DealStatusFailed:
		status = types.Failed
	}

	return types.Deal{
		Id:          deal.ID,
		Name:        deal.Name,
		Description: deal.Description,
		CreatedAt:   deal.CreatedAt,
		UpdatedAt:   deal.UpdatedAt,
		Status:      status,
		Items:       itemsDTO,
	}
}
