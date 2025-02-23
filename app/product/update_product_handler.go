package product

import (
	"context"
	"golang-fiber-poc/domain"
)

type UpdateProductRequest struct {
	ID   string `json:"id" param:"id"`
	Name string `json:"name"`
}

type UpdateProductResponse struct {
	ID string `json:"id"`
}

type UpdateProductHandler struct {
	repository Repository
}

func NewUpdateProductHandler(repository Repository) *UpdateProductHandler {
	return &UpdateProductHandler{repository: repository}
}

func (h *UpdateProductHandler) Handle(ctx context.Context, req *UpdateProductRequest) (*UpdateProductResponse, error) {

	product := domain.Product{
		ID:   req.ID,
		Name: req.Name,
	}

	err := h.repository.CreateProduct(ctx, &product)
	if err != nil {
		return nil, err
	}

	return &UpdateProductResponse{ID: product.ID}, nil
}
