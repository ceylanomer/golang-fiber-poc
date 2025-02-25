package client

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.uber.org/zap"
)

type CustomRetryableClient struct {
	*retryablehttp.Client
}

// NewRetryableClient creates a new retryable client
func NewRetryableClient(transport *http.Transport) CustomRetryableClient {
	retryClient := retryablehttp.NewClient()
	retryClient.Logger = zap.NewStdLog(zap.L())
	retryClient.HTTPClient.Transport = otelhttp.NewTransport(transport)
	retryClient.RetryMax = 3
	retryClient.RetryWaitMin = 100 * time.Millisecond
	retryClient.RetryWaitMax = 10 * time.Second
	retryClient.Backoff = retryablehttp.LinearJitterBackoff

	retryClient.CheckRetry = func(ctx context.Context, resp *http.Response, err error) (bool, error) {
		if ctx.Err() != nil {
			return false, ctx.Err()
		}
		return retryablehttp.DefaultRetryPolicy(ctx, resp, err)
	}
	return CustomRetryableClient{retryClient}
}

func (c *CustomRetryableClient) GetTimeout(ctx context.Context) error {

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost:8081/timeout", nil)
	if err != nil {
		zap.L().Error("Failed to create request to google", zap.Error(err))
		return err
	}

	retryableRequest, err := retryablehttp.FromRequest(req)
	if err != nil {
		zap.L().Error("Failed to create retryable request", zap.Error(err))
		return err
	}

	resp, err := c.Client.Do(retryableRequest)
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

func (c *CustomRetryableClient) GetError(ctx context.Context) error {

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost:8081/error", nil)
	if err != nil {
		zap.L().Error("Failed to create request to google", zap.Error(err))
		return err
	}

	retryableRequest, err := retryablehttp.FromRequest(req)
	if err != nil {
		zap.L().Error("Failed to create retryable request", zap.Error(err))
		return err
	}

	resp, err := c.Client.Do(retryableRequest)
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

func (c *CustomRetryableClient) GetTest(ctx context.Context) error {

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost:8081/test", nil)
	if err != nil {
		zap.L().Error("Failed to create request to google", zap.Error(err))
		return err
	}

	retryableRequest, err := retryablehttp.FromRequest(req)
	if err != nil {
		zap.L().Error("Failed to create retryable request", zap.Error(err))
		return err
	}

	resp, err := c.Client.Do(retryableRequest)
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
