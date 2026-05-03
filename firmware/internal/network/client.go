package network

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client communicates with the Sentinel cloud platform.
type Client struct {
	baseURL    string
	deviceID   string
	httpClient *http.Client
}

func NewClient(baseURL, deviceID string) *Client {
	return &Client{
		baseURL:  baseURL,
		deviceID: deviceID,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// DeviceID returns the registered device ID.
func (c *Client) DeviceID() string { return c.deviceID }

// SetDeviceID sets the device ID after registration.
func (c *Client) SetDeviceID(id string) { c.deviceID = id }

// Register registers this device with the cloud platform.
func (c *Client) Register(ctx context.Context, payload map[string]interface{}) (map[string]interface{}, error) {
	return c.doPost(ctx, "/api/v1/devices", payload)
}

// SendHeartbeat sends a health/heartbeat update. Returns an error if device ID is empty.
func (c *Client) SendHeartbeat(ctx context.Context, health map[string]interface{}) error {
	if c.deviceID == "" {
		return errors.New("cannot send heartbeat: device not registered (empty device ID)")
	}
	_, err := c.doPost(ctx, fmt.Sprintf("/api/v1/devices/%s/heartbeat", c.deviceID), health)
	return err
}

// SendEvent sends a security or telemetry event to the analytics service.
// The analyticsURL should be the base URL of the Python analytics service.
func (c *Client) SendEvent(ctx context.Context, analyticsURL string, event map[string]interface{}) (map[string]interface{}, error) {
	return c.doPostAbsolute(ctx, analyticsURL+"/api/v1/events", event)
}

// GetConfig retrieves the current device configuration.
func (c *Client) GetConfig(ctx context.Context) (map[string]interface{}, error) {
	if c.deviceID == "" {
		return nil, errors.New("cannot get config: device not registered (empty device ID)")
	}
	return c.doGet(ctx, fmt.Sprintf("%s/api/v1/devices/%s", c.baseURL, c.deviceID))
}

// ReportFirmwareStatus reports the result of a firmware update.
func (c *Client) ReportFirmwareStatus(ctx context.Context, version, status string) error {
	if c.deviceID == "" {
		return errors.New("cannot report firmware status: device not registered (empty device ID)")
	}
	_, err := c.doPost(ctx, fmt.Sprintf("/api/v1/devices/%s/firmware-status", c.deviceID), map[string]interface{}{
		"version": version,
		"status":  status,
	})
	return err
}

func (c *Client) doGet(ctx context.Context, url string) (map[string]interface{}, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}
	return decodeResponse(resp.Body)
}

func (c *Client) doPost(ctx context.Context, path string, payload interface{}) (map[string]interface{}, error) {
	return c.doPostAbsolute(ctx, c.baseURL+path, payload)
}

func (c *Client) doPostAbsolute(ctx context.Context, url string, payload interface{}) (map[string]interface{}, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}
	return decodeResponse(resp.Body)
}

func decodeResponse(body io.Reader) (map[string]interface{}, error) {
	var result map[string]interface{}
	if err := json.NewDecoder(body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return result, nil
}
