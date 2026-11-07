package gui

import "context"

type Request struct {
	Method      string
	URL         string
	HeadersJSON string
}

type Result struct {
	StatusText string
	Stdout     string
	Stderr     string
}

type Bridge interface {
	Do(ctx context.Context, req Request) (Result, error)
}
