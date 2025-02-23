package product

import "context"

type GetProductRequest struct {
	Id string `json:"id" param:"id"`
}

type GetProductResponse struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type GetProductHandler struct {
	repository Repository
}

func NewGetProductHandler(repository Repository) *GetProductHandler {
	return &GetProductHandler{
		repository: repository,
	}
}

func (h *GetProductHandler) Handle(ctx context.Context, req *GetProductRequest) (*GetProductResponse, error) {
	product, err := h.repository.GetProduct(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	return &GetProductResponse{
		Id:   product.ID,
		Name: product.Name,
	}, nil
}
