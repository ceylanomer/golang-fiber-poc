package product

import (
	"context"
	"golang-fiber-poc/app/client"
)

type GetProductRequest struct {
	Id string `json:"id" param:"id"`
}

type GetProductResponse struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type GetProductHandler struct {
	repository Repository
	client     client.CustomRetryableClient
}

func NewGetProductHandler(repository Repository, client client.CustomRetryableClient) *GetProductHandler {
	return &GetProductHandler{
		repository: repository,
		client:     client,
	}
}

func (h *GetProductHandler) Handle(ctx context.Context, req *GetProductRequest) (*GetProductResponse, error) {

	err := h.client.GetError(ctx)
	if err != nil {
		return nil, err
	}

	product, err := h.repository.GetProduct(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	return &GetProductResponse{
		Id:   product.ID,
		Name: product.Name,
	}, nil
}
