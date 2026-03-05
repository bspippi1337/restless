package gui

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"strings"
	"time"
)

type ExecBridge struct {
	Binary  string
	Timeout time.Duration
}

func NewExecBridge() *ExecBridge {
	return &ExecBridge{
		Binary:  "",
		Timeout: 25 * time.Second,
	}
}

func (b *ExecBridge) resolveBinary() string {
	if strings.TrimSpace(b.Binary) != "" {
		return b.Binary
	}
	if _, err := os.Stat("./restless"); err == nil {
		return "./restless"
	}
	return "restless"
}

func (b *ExecBridge) Do(ctx context.Context, req Request) (Result, error) {
	bin := b.resolveBinary()

	if strings.TrimSpace(req.URL) == "" {
		return Result{}, errors.New("missing URL")
	}

	args := []string{}
	m := strings.ToUpper(strings.TrimSpace(req.Method))
	if m == "" || m == "GET" {
		args = []string{req.URL}
	} else {
		args = []string{m, req.URL}
	}

	cctx, cancel := context.WithTimeout(ctx, b.Timeout)
	defer cancel()

	cmd := exec.CommandContext(cctx, bin, args...)
	out, err := cmd.CombinedOutput()

	res := Result{
		StatusText: "OK",
		Stdout:     string(out),
	}

	if cctx.Err() == context.DeadlineExceeded {
		return res, errors.New("request timed out")
	}
	if err != nil {
		res.StatusText = "ERROR"
		return res, err
	}
	return res, nil
}
