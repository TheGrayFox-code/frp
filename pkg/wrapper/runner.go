package wrapper

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"text/template"
)

type Runner struct {
	mu      sync.Mutex
	cmd     *exec.Cmd
	profile *Profile
	lastLog bytes.Buffer
}

func NewRunner() *Runner {
	return &Runner{}
}

func (r *Runner) Start(profile Profile) error {
	if profile.ProxyType == "" {
		profile.ProxyType = "tcp"
	}
	if err := profile.Validate(); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	if r.cmd != nil {
		return fmt.Errorf("frpc is already running with profile %q", r.profile.Name)
	}

	configPath, err := writeTempConfig(profile)
	if err != nil {
		return err
	}

	cmd := exec.CommandContext(context.Background(), "frpc", "-c", configPath)
	cmd.Stdout = &r.lastLog
	cmd.Stderr = &r.lastLog
	if err := cmd.Start(); err != nil {
		_ = os.Remove(configPath)
		return fmt.Errorf("start frpc: %w", err)
	}

	r.profile = &profile
	r.cmd = cmd
	go func(cfgPath string, c *exec.Cmd) {
		_ = c.Wait()
		_ = os.Remove(cfgPath)
		r.mu.Lock()
		defer r.mu.Unlock()
		if r.cmd == c {
			r.cmd = nil
			r.profile = nil
		}
	}(configPath, cmd)

	return nil
}

func (r *Runner) Stop() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.cmd == nil {
		return nil
	}
	if err := r.cmd.Process.Kill(); err != nil {
		return fmt.Errorf("stop frpc process: %w", err)
	}
	r.cmd = nil
	r.profile = nil
	return nil
}

func (r *Runner) Status() (running bool, profileName string, logs string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.cmd == nil {
		return false, "", r.lastLog.String()
	}
	return true, r.profile.Name, r.lastLog.String()
}

func writeTempConfig(profile Profile) (string, error) {
	cfgDir := os.TempDir()
	cfgPath := filepath.Join(cfgDir, "frpc-wrapper-"+profile.Name+".toml")

	tpl := `serverAddr = "{{ .ServerAddr }}"
serverPort = {{ .ServerPort }}

[auth]
method = "token"
token = "{{ .AuthToken }}"

[[proxies]]
name = "{{ .ProxyName }}"
type = "{{ .ProxyType }}"
localIP = "{{ .LocalIP }}"
localPort = {{ .LocalPort }}
remotePort = {{ .RemotePort }}
`

	t, err := template.New("frpcCfg").Parse(tpl)
	if err != nil {
		return "", fmt.Errorf("parse frpc config template: %w", err)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, profile); err != nil {
		return "", fmt.Errorf("execute frpc config template: %w", err)
	}
	if err := os.WriteFile(cfgPath, buf.Bytes(), 0o600); err != nil {
		return "", fmt.Errorf("write temp frpc config: %w", err)
	}
	return cfgPath, nil
}

func DefaultMode() string {
	if runtime.GOOS == "linux" {
		return "tui"
	}
	return "gui"
}
