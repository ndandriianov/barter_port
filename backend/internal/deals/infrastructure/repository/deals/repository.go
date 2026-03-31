package deals

import (
	"barter-port/pkg/db"
	"context"
)

type Repository struct{}

func NewRepository() *Repository {
	return &Repository{}
}

func (r *Repository) CreateDeal(ctx context.Context, exec db.DB) {

}
