// dispatcher.go – HTTP webhook dispatcher with HMAC-SHA256 request signing and configurable retry logic.
package webhooks

// Author: Benjamin Stonestreet
// Created: 2026-04-12

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	defaultTimeout     = 10 * time.Second
	defaultMaxAttempts = 3
	defaultBackoff     = 500 * time.Millisecond
	maxResponseBytes   = 8 * 1024
)

// DispatcherConfig controls timeout and retry behavior.
type DispatcherConfig struct {
	Timeout     time.Duration
	MaxAttempts int
	Backoff     time.Duration
	UserAgent   string
}

// Dispatcher sends signed webhook events over HTTP.
type Dispatcher struct {
	client      *http.Client
	maxAttempts int
	backoff     time.Duration
	userAgent   string
}

// DispatchResult reports the final HTTP delivery state.
type DispatchResult struct {
	StatusCode int
	Attempt    int
	Body       string
}

// EventEnvelope is the normalized webhook message sent to merchants.
type EventEnvelope struct {
	EventID   string      `json:"event_id"`
	EventType string      `json:"event_type"`
	CreatedAt time.Time   `json:"created_at"`
	Data      interface{} `json:"data"`
}

// NewDispatcher constructs a webhook dispatcher with sane defaults.
func NewDispatcher(config DispatcherConfig) *Dispatcher {
	timeout := config.Timeout
	if timeout <= 0 {
		timeout = defaultTimeout
	}

	maxAttempts := config.MaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = defaultMaxAttempts
	}

	backoff := config.Backoff
	if backoff <= 0 {
		backoff = defaultBackoff
	}

	userAgent := strings.TrimSpace(config.UserAgent)
	if userAgent == "" {
		userAgent = "XRPay-Webhook-Dispatcher/1.0"
	}

	return &Dispatcher{
		client:      &http.Client{Timeout: timeout},
		maxAttempts: maxAttempts,
		backoff:     backoff,
		userAgent:   userAgent,
	}
}

// Dispatch posts a signed event envelope to the target webhook URL.
// It retries on transient failures up to the configured MaxAttempts with linear back-off.
func (d *Dispatcher) Dispatch(ctx context.Context, webhookURL string, webhookKey string, eventType string, payload interface{}) (*DispatchResult, error) {
	webhookURL = strings.TrimSpace(webhookURL)
	webhookKey = strings.TrimSpace(webhookKey)
	eventType = strings.TrimSpace(eventType)

	if webhookURL == "" {
		return nil, errors.New("webhook url is required")
	}
	if webhookKey == "" {
		return nil, errors.New("webhook key is required")
	}
	if eventType == "" {
		return nil, errors.New("event type is required")
	}

	envelope := EventEnvelope{
		EventID:   strconv.FormatInt(time.Now().UTC().UnixNano(), 10),
		EventType: eventType,
		CreatedAt: time.Now().UTC(),
		Data:      payload,
	}

	requestBody, err := json.Marshal(envelope)
	if err != nil {
		return nil, err
	}

	timestamp := strconv.FormatInt(time.Now().UTC().Unix(), 10)
	signature := SignPayload(requestBody, webhookKey, timestamp)
	signatureHeader := fmt.Sprintf("t=%s,v1=%s", timestamp, signature)

	var lastErr error
	for attempt := 1; attempt <= d.maxAttempts; attempt++ {
		result, tryErr := d.dispatchAttempt(ctx, webhookURL, eventType, requestBody, timestamp, signatureHeader)
		if result != nil {
			result.Attempt = attempt
		}
		if tryErr == nil {
			return result, nil
		}

		lastErr = tryErr
		if !isRetryableError(result, tryErr) || attempt == d.maxAttempts {
			break
		}

		// Apply linear back-off between attempts.
		backoff := time.Duration(attempt) * d.backoff
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(backoff):
		}
	}

	if lastErr == nil {
		lastErr = errors.New("webhook dispatch failed")
	}

	return nil, lastErr
}

// dispatchAttempt performs a single HTTP POST attempt to the webhook endpoint.
func (d *Dispatcher) dispatchAttempt(ctx context.Context, webhookURL string, eventType string, body []byte, timestamp string, signatureHeader string) (*DispatchResult, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, webhookURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")
	request.Header.Set("User-Agent", d.userAgent)
	request.Header.Set("X-Webhook-Event", eventType)
	request.Header.Set("X-Webhook-Timestamp", timestamp)
	request.Header.Set("X-Webhook-Signature", signatureHeader)

	response, err := d.client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	bodyReader := io.LimitReader(response.Body, maxResponseBytes)
	responseBody, _ := io.ReadAll(bodyReader)
	result := &DispatchResult{
		StatusCode: response.StatusCode,
		Body:       strings.TrimSpace(string(responseBody)),
	}

	if response.StatusCode >= 200 && response.StatusCode < 300 {
		return result, nil
	}

	return result, fmt.Errorf("webhook endpoint returned status %d", response.StatusCode)
}

// isRetryableError returns true if the given result and error indicate a transient
// failure that is worth retrying (network errors, timeouts, 429, 5xx responses).
func isRetryableError(result *DispatchResult, err error) bool {
	if err == nil {
		return false
	}

	if result == nil {
		return true
	}

	return result.StatusCode == http.StatusRequestTimeout ||
		result.StatusCode == http.StatusTooManyRequests ||
		result.StatusCode >= 500
}

// SignPayload computes HMAC-SHA256 over "timestamp.payload".
func SignPayload(payload []byte, webhookKey string, timestamp string) string {
	h := hmac.New(sha256.New, []byte(webhookKey))
	h.Write([]byte(timestamp))
	h.Write([]byte("."))
	h.Write(payload)
	return hex.EncodeToString(h.Sum(nil))
}
