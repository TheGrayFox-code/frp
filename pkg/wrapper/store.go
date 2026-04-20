package wrapper

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type Store struct {
	path string
	mu   sync.Mutex
}

func NewStore() (*Store, error) {
	base, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("get home directory: %w", err)
	}
	path := filepath.Join(base, ".frp-wrapper", "profiles.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create wrapper config dir: %w", err)
	}
	return &Store{path: path}, nil
}

func (s *Store) Load() ([]Profile, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.loadUnlocked()
}

func (s *Store) Save(profiles []Profile) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := json.MarshalIndent(profiles, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal profiles: %w", err)
	}
	if err := os.WriteFile(s.path, data, 0o644); err != nil {
		return fmt.Errorf("write profiles file: %w", err)
	}
	return nil
}

func (s *Store) Upsert(profile Profile) error {
	if profile.ProxyType == "" {
		profile.ProxyType = "tcp"
	}
	if err := profile.Validate(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	profiles, err := s.loadUnlocked()
	if err != nil {
		return err
	}

	updated := false
	for i, p := range profiles {
		if p.Name == profile.Name {
			profiles[i] = profile
			updated = true
			break
		}
	}
	if !updated {
		profiles = append(profiles, profile)
	}
	return s.saveUnlocked(profiles)
}

func (s *Store) loadUnlocked() ([]Profile, error) {
	data, err := os.ReadFile(s.path)
	if os.IsNotExist(err) {
		return []Profile{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read profiles file: %w", err)
	}
	if len(data) == 0 {
		return []Profile{}, nil
	}
	var profiles []Profile
	if err := json.Unmarshal(data, &profiles); err != nil {
		return nil, fmt.Errorf("unmarshal profiles file: %w", err)
	}
	return profiles, nil
}

func (s *Store) saveUnlocked(profiles []Profile) error {
	data, err := json.MarshalIndent(profiles, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal profiles: %w", err)
	}
	if err := os.WriteFile(s.path, data, 0o644); err != nil {
		return fmt.Errorf("write profiles file: %w", err)
	}
	return nil
}

func (s *Store) Path() string {
	return s.path
}
