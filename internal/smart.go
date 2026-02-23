package internal

import (
"encoding/json"
"fmt"
"net/http"
"time"
)

func RunSmart(url string) error {
client := &http.Client{Timeout: 5 * time.Second}
resp, err := client.Get(url)
if err != nil {
return err
}
defer resp.Body.Close()

result := map[string]any{
"url":    url,
"status": resp.StatusCode,
"server": resp.Header.Get("Server"),
"type":   resp.Header.Get("Content-Type"),
}

out, _ := json.MarshalIndent(result, "", "  ")
fmt.Println(string(out))
return nil
}
