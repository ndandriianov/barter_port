package offergroups

import (
	"barter-port/contracts/openapi/deals/types"
	"barter-port/internal/deals/domain"

	"github.com/google/uuid"
)

type offerGroupResponse struct {
	Id          uuid.UUID                `json:"id"`
	Name        string                   `json:"name"`
	Description *string                  `json:"description,omitempty"`
	Units       []offerGroupUnitResponse `json:"units"`
}

type offerGroupUnitResponse struct {
	Id     uuid.UUID     `json:"id"`
	Offers []types.Offer `json:"offers"`
}

func mapOfferGroupToDTO(item domain.OfferGroup) offerGroupResponse {
	units := make([]offerGroupUnitResponse, 0, len(item.Units))
	for _, unit := range item.Units {
		offers := make([]types.Offer, 0, len(unit.Offers))
		for _, offer := range unit.Offers {
			offers = append(offers, offer.ToDto())
		}

		units = append(units, offerGroupUnitResponse{
			Id:     unit.ID,
			Offers: offers,
		})
	}

	return offerGroupResponse{
		Id:          item.ID,
		Name:        item.Name,
		Description: item.Description,
		Units:       units,
	}
}

func mapOfferGroupsToDTO(items []domain.OfferGroup) []offerGroupResponse {
	result := make([]offerGroupResponse, 0, len(items))
	for _, item := range items {
		result = append(result, mapOfferGroupToDTO(item))
	}
	return result
}
