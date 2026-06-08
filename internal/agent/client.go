package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"time"
)

// Client talks to the InfraBIOS central server.
type Client struct {
	baseURL    string
	token      string
	serverID   string
	httpClient *http.Client
}

func NewClient(baseURL, token, serverID string) *Client {
	return &Client{
		baseURL:  baseURL,
		token:    token,
		serverID: serverID,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// Register creates or verifies the server record on the central server.
func (c *Client) Register(info HardwareInfo) error {
	body, _ := json.Marshal(map[string]any{
		"hostname":      hostnameMust(),
		"ip_address":    localIP(),
		"vendor":        info.Vendor,
		"model":         info.Model,
		"serial_number": info.SerialNumber,
		"bios_version":  info.BIOSVersion,
		"bmc_version":   info.BMCVersion,
		"device_type":   info.DeviceType,
		"os":            info.OS,
	})
	resp, err := c.post("/api/v1/servers", body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusConflict || resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusOK {
		return nil
	}
	return fmt.Errorf("register: unexpected status %d", resp.StatusCode)
}

// SubmitInventory sends the latest BIOS settings to the central server.
func (c *Client) SubmitInventory(settings, firmware json.RawMessage, agentVersion string) error {
	body, _ := json.Marshal(map[string]any{
		"settings":      settings,
		"firmware":      firmware,
		"agent_version": agentVersion,
	})
	resp, err := c.post(fmt.Sprintf("/api/v1/agent/servers/%s/inventory", c.serverID), body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("submit inventory: status %d", resp.StatusCode)
	}
	return nil
}

func (c *Client) post(path string, body []byte) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPost, c.baseURL+path, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token)
	return c.httpClient.Do(req)
}

func hostnameMust() string {
	b, err := exec.Command("hostname").Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(b))
}

func localIP() string {
	b, _ := exec.Command("hostname", "-I").Output()
	fields := strings.Fields(string(b))
	if len(fields) > 0 {
		return fields[0]
	}
	return ""
}
