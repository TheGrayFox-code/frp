package wrapper

import "fmt"

// Profile defines a simplified frpc connection profile.
type Profile struct {
	Name       string `json:"name"`
	ServerAddr string `json:"serverAddr"`
	ServerPort int    `json:"serverPort"`
	AuthToken  string `json:"authToken"`
	ProxyName  string `json:"proxyName"`
	ProxyType  string `json:"proxyType"`
	LocalIP    string `json:"localIP"`
	LocalPort  int    `json:"localPort"`
	RemotePort int    `json:"remotePort"`
}

func (p Profile) Validate() error {
	if p.Name == "" {
		return fmt.Errorf("profile name is required")
	}
	if p.ServerAddr == "" {
		return fmt.Errorf("server address is required")
	}
	if p.ServerPort <= 0 {
		return fmt.Errorf("server port must be > 0")
	}
	if p.ProxyName == "" {
		return fmt.Errorf("proxy name is required")
	}
	if p.ProxyType == "" {
		return fmt.Errorf("proxy type is required")
	}
	if p.ProxyType != "tcp" {
		return fmt.Errorf("only tcp is supported by this wrapper right now")
	}
	if p.LocalIP == "" {
		return fmt.Errorf("local ip is required")
	}
	if p.LocalPort <= 0 {
		return fmt.Errorf("local port must be > 0")
	}
	if p.RemotePort <= 0 {
		return fmt.Errorf("remote port must be > 0")
	}
	return nil
}
