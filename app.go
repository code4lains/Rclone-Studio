package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// App struct
type App struct {
	ctx             context.Context
	instanceManager *InstanceManager
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		instanceManager: NewInstanceManager(),
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.instanceManager.SetContext(ctx)
	// Ignore load error for now, it's fine if file doesn't exist
	_ = a.instanceManager.Load()
}

// GetInstances returns all instances
func (a *App) GetInstances() []*Instance {
	return a.instanceManager.GetInstances()
}

// SaveInstance saves an instance
func (a *App) SaveInstance(inst Instance) error {
	return a.instanceManager.SaveInstance(inst)
}

// DeleteInstance deletes an instance
func (a *App) DeleteInstance(id string) error {
	return a.instanceManager.DeleteInstance(id)
}

// StartInstance starts an instance
func (a *App) StartInstance(id string) error {
	return a.instanceManager.StartInstance(id)
}

// StopInstance stops an instance
func (a *App) StopInstance(id string) error {
	return a.instanceManager.StopInstance(id)
}

// RestartInstance restarts an instance
func (a *App) RestartInstance(id string) error {
	return a.instanceManager.RestartInstance(id)
}

// CallRcloneAPI makes an HTTP request to the Rclone rc server
func (a *App) CallRcloneAPI(endpoint string, payload interface{}, ip string, port string, user string, pass string) (interface{}, error) {
	if ip == "" {
		return nil, fmt.Errorf("ip address is required")
	}
	if port == "" {
		return nil, fmt.Errorf("port is required")
	}
	url := fmt.Sprintf("http://%s:%s/%s", ip, port, endpoint)

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(a.ctx, http.MethodPost, url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	if user != "" || pass != "" {
		req.SetBasicAuth(user, pass)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("rclone api error: status %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var result interface{}
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return result, nil
}
