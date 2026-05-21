package ollama

import (
	"encoding/json"
	"net/http"
)

type Client struct {
	BaseURL string
	APIKey  string
}

func NewClient(baseURL string) *Client {
	return &Client{
		BaseURL: baseURL,
	}
}

type ModelPSResponse struct {
	Models []struct {
		Name string `json:"name"`
		Size int64  `json:"size"`
		Details struct {
			ParameterSize string `json:"parameter_size"`
		} `json:"details"`
		ExpiresAt     string `json:"expires_at"`
		SizeVRAM      int64  `json:"size_vram"`
		ContextLength int    `json:"context_length"`
	} `json:"models"`
}

func (c *Client) GetRunningModels() (*ModelPSResponse, error) {
	req, err := http.NewRequest("GET", c.BaseURL+"/api/ps", nil)
	if err != nil {
		return nil, err
	}

	if c.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var result ModelPSResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	return &result, err
}
