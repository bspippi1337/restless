package adapter

import (
	"context"
	"net/http"
	"time"

	"github.com/bspippi1337/restless/internal/core/app"
	"github.com/bspippi1337/restless/internal/core/types"
	"github.com/bspippi1337/restless/internal/modules/bench"
	"github.com/bspippi1337/restless/internal/modules/export"
	"github.com/bspippi1337/restless/internal/modules/openapi"
	"github.com/bspippi1337/restless/internal/modules/session"
)

type RequestConfig struct {
	URL     string
	Method  string
	Body    string
	Timeout int
}

func RunRequest(cfg RequestConfig) error {
	sess := session.New()
	mods := []app.Module{
		sess,
		openapi.New(),
		export.New(),
		bench.New(),
	}

	a, err := app.New(mods)
	if err != nil {
		return err
	}

	req := types.Request{
		Method:  cfg.Method,
		URL:     cfg.URL,
		Headers: http.Header{},
		Body:    []byte(cfg.Body),
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.Timeout)*time.Second)
	defer cancel()

	resp, err := a.RunOnce(ctx, req)
	if err != nil {
		return err
	}

	println("status:", resp.StatusCode)
	println(string(resp.Body))
	return nil
}

func RunProbe(cfg RequestConfig) error {
	// Probe er bare GET request med samme motor
	cfg.Method = "GET"
	return RunRequest(cfg)
}
