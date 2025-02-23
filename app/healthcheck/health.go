package healthcheck

import "context"

type Request struct {
}

type Response struct {
	Status string `json:"status"`
}

type Handler struct {
}

func NewHealthCheckHandler() *Handler {
	return &Handler{}
}

func (h *Handler) Handle(ctx context.Context, req *Request) (*Response, error) {
	return &Response{Status: "OK"}, nil
}
