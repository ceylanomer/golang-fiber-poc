package client

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.uber.org/zap"
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

func (c *CustomHttpClient) GetTimeout(ctx context.Context) error {

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost:8081/timeout", nil)
	if err != nil {
		zap.L().Error("Failed to create request to timeout", zap.Error(err))
		return err
	}

	resp, doErr := c.Client.Do(req)
	if doErr != nil {
		zap.L().Error("Failed to make request to timeout", zap.Error(doErr))
		return doErr
	}

	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		zap.L().Error("Server returned error status",
			zap.Int("statusCode", resp.StatusCode),
			zap.String("status", resp.Status),
		)
		return fmt.Errorf("server returned status code: %d", resp.StatusCode)
	}

	_, err = io.ReadAll(resp.Body)
	if err != nil {
		zap.L().Error("Failed to read response from timeout", zap.Error(err))
		return err
	}

	return nil
}

func (c *CustomHttpClient) GetError(ctx context.Context) error {

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost:8081/error", nil)
	if err != nil {
		zap.L().Error("Failed to create request to error", zap.Error(err))
		return err
	}

	resp, doErr := c.Client.Do(req)
	if doErr != nil {
		zap.L().Error("Failed to make request to error", zap.Error(doErr))
		return doErr
	}

	defer resp.Body.Close()

	// Additional status code check is needed here because doErr returns nil even when status code is 500
	if resp.StatusCode >= 400 {
		zap.L().Error("Server returned error status",
			zap.Int("statusCode", resp.StatusCode),
			zap.String("status", resp.Status),
		)
		return fmt.Errorf("server returned status code: %d", resp.StatusCode)
	}

	_, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		zap.L().Error("Failed to read response from error", zap.Error(readErr))
		return readErr
	}

	return nil
}

func (c *CustomHttpClient) GetTest(ctx context.Context) error {

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost:8081/test", nil)
	if err != nil {
		zap.L().Error("Failed to create request to test", zap.Error(err))
		return err
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		zap.L().Error("Failed to make request to test", zap.Error(err))
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		zap.L().Error("Server returned error status",
			zap.Int("statusCode", resp.StatusCode),
			zap.String("status", resp.Status),
		)
		return fmt.Errorf("server returned status code: %d", resp.StatusCode)
	}

	_, err = io.ReadAll(resp.Body)
	if err != nil {
		zap.L().Error("Failed to read response from test", zap.Error(err))
		return err
	}

	return nil
}
