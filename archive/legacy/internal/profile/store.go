package profile

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Profile struct {
	Name string `json:"name"`
	Base string `json:"base"`
}

type Config struct {
	Active   string             `json:"active"`
	Profiles map[string]Profile `json:"profiles"`
}

func configPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".restless", "profiles.json"), nil
}

func Load() (Config, error) {
	p, err := configPath()
	if err != nil {
		return Config{}, err
	}
	b, err := os.ReadFile(p)
	if err != nil {
		return Config{Profiles: map[string]Profile{}}, nil
	}
	var c Config
	if err := json.Unmarshal(b, &c); err != nil {
		return Config{}, err
	}
	if c.Profiles == nil {
		c.Profiles = map[string]Profile{}
	}
	return c, nil
}

func Save(c Config) error {
	p, err := configPath()
	if err != nil {
		return err
	}
	os.MkdirAll(filepath.Dir(p), 0755)
	b, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, b, 0644)
}
