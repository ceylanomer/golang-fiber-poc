package product

import (
	"context"
	"golang-fiber-poc/app/client"
	"golang-fiber-poc/pkg/circuitbreaker"
	"time"

	"github.com/sony/gobreaker"
)

type GetProductRequest struct {
	Id string `json:"id" param:"id"`
}

type GetProductResponse struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type GetProductHandler struct {
	repository    Repository
	client        client.CustomRetryableClient
	noRetryClient client.CustomHttpClient
	cb            *gobreaker.CircuitBreaker
}

func NewGetProductHandler(repository Repository, client client.CustomRetryableClient, noRetryClient client.CustomHttpClient) *GetProductHandler {
	cb := circuitbreaker.NewCircuitBreaker(circuitbreaker.CircuitBreakerConfig{
		Name:                    "get-product",
		MaxRequests:             3,
		Interval:                10 * time.Second,
		Timeout:                 5 * time.Second,
		RequestsVolumeThreshold: 10,
		FailureThreshold:        0.6,
	})

	return &GetProductHandler{
		repository:    repository,
		client:        client,
		noRetryClient: noRetryClient,
		cb:            cb,
	}
}

func (h *GetProductHandler) Handle(ctx context.Context, req *GetProductRequest) (*GetProductResponse, error) {
	// Execute GetError through circuit breaker
	_, err := h.cb.Execute(func() (interface{}, error) {
		return nil, h.noRetryClient.GetError(ctx)
	})
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
