package ghactions

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	Repo  string
	Token string
	HTTP  *http.Client
}

func New(repo, token string) *Client {
	return &Client{
		Repo:  repo,
		Token: token,
		HTTP:  &http.Client{Timeout: 20 * time.Second},
	}
}

type dispatchBody struct {
	Ref    string            `json:"ref"`
	Inputs map[string]string `json:"inputs"`
}

func (c *Client) Dispatch(workflow, ref string, inputs map[string]string) error {

	body := dispatchBody{Ref: ref, Inputs: inputs}
	b, _ := json.Marshal(body)

	url := fmt.Sprintf("https://api.github.com/repos/%s/actions/workflows/%s/dispatches", c.Repo, workflow)

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(b))
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		msg, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("dispatch failed: %s %s", resp.Status, string(msg))
	}

	return nil
}
