package tunnel

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"text/template"
	"utopia-server/internal/config"
)

const frpsConfigTemplate = `
bindPort = {{ .BindPort }}
auth.token = "{{ .Token }}"
webServer.port = {{ .DashboardPort }}
webServer.addr = "{{ .DashboardAddr }}"
webServer.user = "{{ .DashboardUser }}"
webServer.password = "{{ .DashboardPwd }}"
`

// Service manages the frps subprocess.
type Service struct {
	config     config.FRPConfig
	cmd        *exec.Cmd
	configFile string
}

// NewService creates a new frps service.
func NewService(config config.FRPConfig) *Service {
	return &Service{
		config: config,
	}
}

// Start starts the frps subprocess.
func (s *Service) Start() error {
	tmpl, err := template.New("frpsConfig").Parse(frpsConfigTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse frps config template: %w", err)
	}

	var configContent bytes.Buffer
	if err := tmpl.Execute(&configContent, s.config); err != nil {
		return fmt.Errorf("failed to execute frps config template: %w", err)
	}

	configDir := "configs"
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	s.configFile = filepath.Join(configDir, "frps.toml")
	if err := os.WriteFile(s.configFile, configContent.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write frps config file: %w", err)
	}

	s.cmd = exec.Command("frps", "-c", s.configFile)
	s.cmd.Stdout = os.Stdout
	s.cmd.Stderr = os.Stderr

	if err := s.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start frps: %w", err)
	}

	log.Printf("frps started with PID %d, config: %s", s.cmd.Process.Pid, s.configFile)
	return nil
}

// Stop stops the frps subprocess.
func (s *Service) Stop() error {
	if s.cmd == nil || s.cmd.Process == nil {
		return nil
	}

	log.Printf("Stopping frps with PID %d...", s.cmd.Process.Pid)
	if err := s.cmd.Process.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("failed to send SIGTERM to frps: %w", err)
	}

	if err := os.Remove(s.configFile); err != nil {
		log.Printf("Warning: failed to remove frps config file: %v", err)
	}

	log.Println("frps stopped.")
	return nil
}
