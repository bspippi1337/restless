package store

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Endpoint struct {
	Path    string
	Methods []string
}

type API struct {
	BaseURL   string
	Endpoints []Endpoint
}

func DefaultRoot(custom string) (string, error) {
	if custom != "" {
		return custom, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".restless"), nil
}

func Write(root string, api *API) (string, error) {

	dir := filepath.Join(root, "apis")
	os.MkdirAll(dir, 0755)

	path := filepath.Join(dir, "last.json")

	b, _ := json.MarshalIndent(api, "", "  ")

	err := os.WriteFile(path, b, 0644)
	return path, err
}

func Read(root, name string) (*API, error) {

	path := filepath.Join(root, "apis", "last.json")

	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var api API
	json.Unmarshal(b, &api)

	return &api, nil
}

var last API

func Save(a API) {
	last = a
}

func Last() API {
	return last
}
