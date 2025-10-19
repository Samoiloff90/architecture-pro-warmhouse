package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type DeviceServiceClient struct {
	baseURL    string
	httpClient *http.Client
}

type Device struct {
	ID         int       `json:"id"`
	Name       string    `json:"name"`
	DeviceType string    `json:"device_type"`
	Location   string    `json:"location"`
	Room       string    `json:"room"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func NewDeviceServiceClient() *DeviceServiceClient {
	baseURL := os.Getenv("DEVICE_SERVICE_URL")
	if baseURL == "" {
		baseURL = "http://device-service:8082"
	}

	return &DeviceServiceClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *DeviceServiceClient) GetDevices(location, status string) ([]Device, error) {
	url := fmt.Sprintf("%s/devices", c.baseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	if location != "" {
		q.Add("location", location)
	}
	if status != "" {
		q.Add("status", status)
	}
	req.URL.RawQuery = q.Encode()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("device service returned %d: %s", resp.StatusCode, string(body))
	}

	var devices []Device
	if err := json.NewDecoder(resp.Body).Decode(&devices); err != nil {
		return nil, err
	}

	return devices, nil
}

func (c *DeviceServiceClient) GetDevice(deviceID int) (*Device, error) {
	url := fmt.Sprintf("%s/devices/%d", c.baseURL, deviceID)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("device not found")
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("device service returned %d: %s", resp.StatusCode, string(body))
	}

	var device Device
	if err := json.NewDecoder(resp.Body).Decode(&device); err != nil {
		return nil, err
	}

	return &device, nil
}

func (c *DeviceServiceClient) CreateDevice(name, deviceType, location, room string) (*Device, error) {
	payload := map[string]string{
		"name":        name,
		"device_type": deviceType,
		"location":    location,
		"room":        room,
		"status":      "active",
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/devices", c.baseURL)
	resp, err := c.httpClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("device service returned %d: %s", resp.StatusCode, string(body))
	}

	var device Device
	if err := json.NewDecoder(resp.Body).Decode(&device); err != nil {
		return nil, err
	}

	return &device, nil
}
