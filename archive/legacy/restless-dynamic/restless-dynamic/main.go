package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
)

type Config struct {
	User         string `json:"user"`
	Repo         string `json:"repo"`
	Radius       int    `json:"radius"`
	ShowReleases bool   `json:"show_releases"`
}

var cfg = Config{
	User:         "bspippi1337",
	Repo:         "restless",
	Radius:       220,
	ShowReleases: true,
}

func validate(c Config) error {
	if c.User == "" || c.Repo == "" {
		return fmt.Errorf("user and repo required")
	}
	if c.Radius < 100 || c.Radius > 600 {
		return fmt.Errorf("radius out of range (100-600)")
	}
	return nil
}

func svg() string {
	centerX := 450
	centerY := 300
	endpoints := []string{"issues", "pulls", "commits"}
	if cfg.ShowReleases {
		endpoints = append(endpoints, "releases")
	}

	out := `<svg xmlns="http://www.w3.org/2000/svg" width="900" height="600" style="background:#0b0f19;font-family:monospace">`
	out += fmt.Sprintf(`<text x="450" y="40" text-anchor="middle" fill="#a78bfa" font-size="22">GitHub API Map: %s/%s</text>`, cfg.User, cfg.Repo)
	out += fmt.Sprintf(`<circle cx="%d" cy="%d" r="60" fill="#1e293b" stroke="#a78bfa" stroke-width="2"/>`, centerX, centerY)
	out += fmt.Sprintf(`<text x="%d" y="%d" text-anchor="middle" fill="#e5e7eb">/repos/%s/%s</text>`, centerX, centerY, cfg.User, cfg.Repo)

	for i, ep := range endpoints {
		angle := (2 * math.Pi / float64(len(endpoints))) * float64(i)
		x := float64(centerX) + float64(cfg.Radius)*math.Cos(angle)
		y := float64(centerY) + float64(cfg.Radius)*math.Sin(angle)

		out += fmt.Sprintf(`<line x1="%d" y1="%d" x2="%.0f" y2="%.0f" stroke="#334155"/>`, centerX, centerY, x, y)
		out += fmt.Sprintf(`<circle cx="%.0f" cy="%.0f" r="50" fill="#0f172a" stroke="#22d3ee" stroke-width="2"/>`, x, y)
		out += fmt.Sprintf(`<text x="%.0f" y="%.0f" text-anchor="middle" fill="#e5e7eb" font-size="13">/%s</text>`, x, y, ep)
	}

	out += `</svg>`
	return out
}

func main() {
	http.HandleFunc("/config", func(w http.ResponseWriter, r *http.Request) {
		var newCfg Config
		if err := json.NewDecoder(r.Body).Decode(&newCfg); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		if err := validate(newCfg); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		cfg = newCfg
		w.WriteHeader(200)
	})

	http.HandleFunc("/graph.svg", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/svg+xml")
		fmt.Fprint(w, svg())
	})

	http.HandleFunc("/state", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(cfg)
	})

	log.Println("Restless dynamic server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
