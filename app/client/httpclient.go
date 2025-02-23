package client

import (
	"context"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.uber.org/zap"
	"io"
	"net"
	"net/http"
	"time"
)

type CustomHttpClient struct {
	*http.Client
}

func NewHttpClient() CustomHttpClient {
	transport := &http.Transport{
		Dial: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).Dial,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	httpClient := &http.Client{
		Transport: otelhttp.NewTransport(transport),
	}

	return CustomHttpClient{httpClient}
}

func (c *CustomHttpClient) GetGoogle(ctx context.Context) error {

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://www.google.com", nil)
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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		zap.L().Error("Failed to read response from google", zap.Error(err))
		return err
	}

	zap.L().Info("Response from google", zap.String("body", string(body)))

	return nil
}
