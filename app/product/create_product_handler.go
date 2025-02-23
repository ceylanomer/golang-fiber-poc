package product

import (
	"context"
	"github.com/google/uuid"
	"golang-fiber-poc/domain"
)

type CreateProductRequest struct {
	Name string `json:"name"`
}

type CreateProductResponse struct {
	ID string `json:"id"`
}

type CreateProductHandler struct {
	repository Repository
}

func NewCreateProductHandler(repository Repository) *CreateProductHandler {
	return &CreateProductHandler{repository: repository}
}

func (h *CreateProductHandler) Handle(ctx context.Context, req *CreateProductRequest) (*CreateProductResponse, error) {
	productId := uuid.New().String()

	product := domain.Product{
		ID:   productId,
		Name: req.Name,
	}

	err := h.repository.CreateProduct(ctx, &product)
	if err != nil {
		return nil, err
	}

	return &CreateProductResponse{ID: product.ID}, nil
}
