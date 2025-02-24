package client

import (
	"context"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.uber.org/zap"
	"io"
	"net/http"
)

type CustomHttpClient struct {
	*http.Client
}

func NewHttpClient(transport *http.Transport) CustomHttpClient {

	httpClient := &http.Client{
		Transport: otelhttp.NewTransport(transport),
	}

	return CustomHttpClient{httpClient}
}

func (c *CustomHttpClient) GetGoogle(ctx context.Context) error {

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost:8081/timeout", nil)
	if err != nil {
		zap.L().Error("Failed to create request to google", zap.Error(err))
		return err
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		zap.L().Error("Failed to make request to google", zap.Error(err))
		return err
	}

	defer resp.Body.Close()

	_, err = io.ReadAll(resp.Body)
	if err != nil {
		zap.L().Error("Failed to read response from google", zap.Error(err))
		return err
	}

	return nil
}
