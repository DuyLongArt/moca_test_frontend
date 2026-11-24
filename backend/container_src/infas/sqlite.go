package infas

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type CloudflareD1Client struct {
	AccountID  string
	DatabaseID string
	APIToken   string
	HTTPClient *http.Client
}

type d1QueryPayload struct {
	SQL    string        `json:"sql"`
	Params []interface{} `json:"params"`
}

type d1APIResponse struct {
	Success bool `json:"success"`
	Errors  []struct {
		Message string `json:"message"`
	} `json:"errors"`
	Result []struct {
		Results []map[string]interface{} `json:"results"`
		Success bool                     `json:"success"`
	} `json:"result"`
}

func NewD1Client(accountID, databaseID, apiToken string) *CloudflareD1Client {
	return &CloudflareD1Client{
		AccountID:  accountID,
		DatabaseID: databaseID,
		APIToken:   apiToken,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *CloudflareD1Client) Query(sqlQuery string, params ...interface{}) ([]map[string]interface{}, error) {
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/accounts/%s/d1/database/%s/query", c.AccountID, c.DatabaseID)

	payload := d1QueryPayload{
		SQL:    sqlQuery,
		Params: params,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.APIToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var d1Resp d1APIResponse
	if err := json.Unmarshal(respBytes, &d1Resp); err != nil {
		return nil, err
	}

	if !d1Resp.Success || len(d1Resp.Errors) > 0 {
		return nil, fmt.Errorf("D1 error: %v", d1Resp.Errors)
	}

	if len(d1Resp.Result) > 0 {
		return d1Resp.Result[0].Results, nil
	}

	return nil, nil
}
