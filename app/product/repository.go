package product

import (
	"context"
	"golang-fiber-poc/domain"
)

type Repository interface {
	GetProduct(ctx context.Context, id string) (*domain.Product, error)
	CreateProduct(ctx context.Context, product *domain.Product) error
}
