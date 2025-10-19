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

type TelemetryServiceClient struct {
	baseURL    string
	httpClient *http.Client
}

type TelemetryData struct {
	ID         int64     `json:"id"`
	DeviceID   int       `json:"device_id"`
	MetricName string    `json:"metric_name"`
	Value      float64   `json:"value"`
	Unit       string    `json:"unit"`
	Timestamp  time.Time `json:"timestamp"`
}

type AggregatedData struct {
	Period   time.Time `json:"period"`
	AvgValue float64   `json:"avg_value"`
	MinValue float64   `json:"min_value"`
	MaxValue float64   `json:"max_value"`
	Count    int       `json:"count"`
}

func NewTelemetryServiceClient() *TelemetryServiceClient {
	baseURL := os.Getenv("TELEMETRY_SERVICE_URL")
	if baseURL == "" {
		baseURL = "http://telemetry-service:8083"
	}

	return &TelemetryServiceClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *TelemetryServiceClient) GetTelemetry(deviceID int, metricName string, limit int) ([]TelemetryData, error) {
	url := fmt.Sprintf("%s/telemetry", c.baseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	if deviceID > 0 {
		q.Add("device_id", fmt.Sprintf("%d", deviceID))
	}
	if metricName != "" {
		q.Add("metric_name", metricName)
	}
	if limit > 0 {
		q.Add("limit", fmt.Sprintf("%d", limit))
	}
	req.URL.RawQuery = q.Encode()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("telemetry service returned %d: %s", resp.StatusCode, string(body))
	}

	var telemetry []TelemetryData
	if err := json.NewDecoder(resp.Body).Decode(&telemetry); err != nil {
		return nil, err
	}

	return telemetry, nil
}

func (c *TelemetryServiceClient) PostTelemetry(deviceID int, metricName string, value float64, unit string) (*TelemetryData, error) {
	payload := map[string]interface{}{
		"device_id":   deviceID,
		"metric_name": metricName,
		"value":       value,
		"unit":        unit,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/telemetry", c.baseURL)
	resp, err := c.httpClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("telemetry service returned %d: %s", resp.StatusCode, string(body))
	}

	var telemetry TelemetryData
	if err := json.NewDecoder(resp.Body).Decode(&telemetry); err != nil {
		return nil, err
	}

	return &telemetry, nil
}

func (c *TelemetryServiceClient) GetAggregatedTelemetry(deviceID int, metricName, interval string) ([]AggregatedData, error) {
	url := fmt.Sprintf("%s/telemetry/aggregated", c.baseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("device_id", fmt.Sprintf("%d", deviceID))
	q.Add("metric_name", metricName)
	if interval != "" {
		q.Add("interval", interval)
	}
	req.URL.RawQuery = q.Encode()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("telemetry service returned %d: %s", resp.StatusCode, string(body))
	}

	var aggregated []AggregatedData
	if err := json.NewDecoder(resp.Body).Decode(&aggregated); err != nil {
		return nil, err
	}

	return aggregated, nil
}
