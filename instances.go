package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"sync"
	"bytes"
	"io"
	"time"
)

type Instance struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Type    string   `json:"type"` // "managed" or "external"
	Command []string `json:"command"` // used for managed
	URL     string   `json:"url"`
	User    string   `json:"user"`
	Pass    string   `json:"pass"`
	PID     int      `json:"pid"`
	Status  string   `json:"status"` // "running" or "stopped"

	cmd *exec.Cmd `json:"-"`
}

type InstanceManager struct {
	instances map[string]*Instance
	mu        sync.Mutex
	ctx       context.Context
	dataFile  string
}

func NewInstanceManager() *InstanceManager {
	return &InstanceManager{
		instances: make(map[string]*Instance),
		dataFile:  "instances.json",
	}
}

func (m *InstanceManager) SetContext(ctx context.Context) {
	m.ctx = ctx
}

func (m *InstanceManager) Load() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, err := os.ReadFile(m.dataFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var list []*Instance
	if err := json.Unmarshal(data, &list); err != nil {
		return err
	}

	m.instances = make(map[string]*Instance)
	for _, inst := range list {
		// Reset ephemeral state
		inst.PID = 0
		inst.Status = "stopped"
		inst.cmd = nil
		m.instances[inst.ID] = inst
	}
	return nil
}

func (m *InstanceManager) Save() error {
	list := make([]*Instance, 0, len(m.instances))
	for _, inst := range m.instances {
		list = append(list, inst)
	}
	data, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(m.dataFile, data, 0644)
}

func (m *InstanceManager) GetInstances() []*Instance {
	m.mu.Lock()
	defer m.mu.Unlock()
	list := make([]*Instance, 0, len(m.instances))
	for _, inst := range m.instances {
		// clone to avoid data race when serializing to frontend
		clone := *inst
		list = append(list, &clone)
	}
	return list
}

func (m *InstanceManager) SaveInstance(inst Instance) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if existing, ok := m.instances[inst.ID]; ok {
		// Update fields, keep runtime state
		existing.Name = inst.Name
		existing.Type = inst.Type
		existing.Command = inst.Command
		existing.URL = inst.URL
		existing.User = inst.User
		existing.Pass = inst.Pass
	} else {
		inst.Status = "stopped"
		m.instances[inst.ID] = &inst
	}
	return m.Save()
}

func (m *InstanceManager) DeleteInstance(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if inst, ok := m.instances[id]; ok {
		if inst.Status == "running" {
			return fmt.Errorf("cannot delete a running instance")
		}
		delete(m.instances, id)
		return m.Save()
	}
	return fmt.Errorf("instance not found")
}

func (m *InstanceManager) StartInstance(id string) error {
	m.mu.Lock()
	inst, ok := m.instances[id]
	m.mu.Unlock()

	if !ok {
		return fmt.Errorf("instance not found")
	}

	if inst.Status == "running" {
		return fmt.Errorf("instance already running")
	}

	if inst.Type == "external" {
		return fmt.Errorf("cannot start an external instance")
	}

	if len(inst.Command) == 0 {
		return fmt.Errorf("no command provided for managed instance")
	}

	args := append([]string{}, inst.Command...)
	hasAllowOrigin := false
	for _, arg := range args {
		if arg == "--rc-allow-origin" || arg == "--rc-allow-origin=*" {
			hasAllowOrigin = true
			break
		}
	}
	if !hasAllowOrigin {
		args = append(args, "--rc-allow-origin", "*")
	}

	cmd := exec.CommandContext(m.ctx, "rclone", args...)
	
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start process: %w", err)
	}

	m.mu.Lock()
	inst.cmd = cmd
	inst.PID = cmd.Process.Pid
	inst.Status = "running"
	m.mu.Unlock()

	// Wait for process in background to update status on exit
	go func(i *Instance) {
		i.cmd.Wait()
		m.mu.Lock()
		i.PID = 0
		i.Status = "stopped"
		i.cmd = nil
		m.mu.Unlock()
	}(inst)

	return nil
}

func (m *InstanceManager) StopInstance(id string) error {
	m.mu.Lock()
	inst, ok := m.instances[id]
	m.mu.Unlock()

	if !ok {
		return fmt.Errorf("instance not found")
	}
	if inst.Status != "running" {
		return fmt.Errorf("instance is not running")
	}

	if inst.Type == "managed" {
		if inst.cmd != nil && inst.cmd.Process != nil {
			err := inst.cmd.Process.Kill()
			if err != nil {
				return fmt.Errorf("failed to kill process: %w", err)
			}
			// status will be updated by the goroutine waiting in StartInstance
			return nil
		}
		return fmt.Errorf("process reference lost")
	} else if inst.Type == "external" {
		// Send core/quit
		return m.sendCoreQuit(inst)
	}

	return fmt.Errorf("unknown instance type")
}

func (m *InstanceManager) sendCoreQuit(inst *Instance) error {
	url := fmt.Sprintf("%s/core/quit", inst.URL)
	req, err := http.NewRequestWithContext(m.ctx, http.MethodPost, url, bytes.NewBufferString("{}"))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if inst.User != "" || inst.Pass != "" {
		req.SetBasicAuth(inst.User, inst.Pass)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send quit request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("quit failed: status %d, %s", resp.StatusCode, string(body))
	}
	
	m.mu.Lock()
	inst.Status = "stopped"
	m.mu.Unlock()
	
	return nil
}

func (m *InstanceManager) RestartInstance(id string) error {
	if err := m.StopInstance(id); err != nil {
		// Ignore if it was already stopped
		if err.Error() != "instance is not running" {
			return fmt.Errorf("stop failed: %w", err)
		}
	}

	// Wait up to 5 seconds for the goroutine to mark it as stopped
	for i := 0; i < 50; i++ {
		m.mu.Lock()
		inst, ok := m.instances[id]
		m.mu.Unlock()
		
		if !ok || inst.Status == "stopped" {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	return m.StartInstance(id)
}
