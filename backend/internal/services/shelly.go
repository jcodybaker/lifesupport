package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// ShellyService handles communication with Shelly smart devices
type ShellyService struct {
	client *http.Client
}

// NewShellyService creates a new Shelly service instance
func NewShellyService() *ShellyService {
	return &ShellyService{
		client: &http.Client{},
	}
}

// ShellyStatus represents the status response from a Shelly device
type ShellyStatus struct {
	IsOn        bool    `json:"ison"`
	Power       float64 `json:"power"`
	Temperature float64 `json:"temperature"`
}

// TurnOn turns on a Shelly device
func (s *ShellyService) TurnOn(shellyID string) error {
	url := fmt.Sprintf("http://%s/relay/0?turn=on", shellyID)
	resp, err := s.client.Get(url)
	if err != nil {
		return fmt.Errorf("failed to turn on device: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// TurnOff turns off a Shelly device
func (s *ShellyService) TurnOff(shellyID string) error {
	url := fmt.Sprintf("http://%s/relay/0?turn=off", shellyID)
	resp, err := s.client.Get(url)
	if err != nil {
		return fmt.Errorf("failed to turn off device: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// Toggle toggles a Shelly device
func (s *ShellyService) Toggle(shellyID string) error {
	url := fmt.Sprintf("http://%s/relay/0?turn=toggle", shellyID)
	resp, err := s.client.Get(url)
	if err != nil {
		return fmt.Errorf("failed to toggle device: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// GetStatus retrieves the current status of a Shelly device
func (s *ShellyService) GetStatus(shellyID string) (*ShellyStatus, error) {
	url := fmt.Sprintf("http://%s/status", shellyID)
	resp, err := s.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get device status: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var status ShellyStatus
	if err := json.Unmarshal(body, &status); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &status, nil
}

// Example usage in handlers:
// shellyService := services.NewShellyService()
// err := shellyService.TurnOn(device.ShellyID)
