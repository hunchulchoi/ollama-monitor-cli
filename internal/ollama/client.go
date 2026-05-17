package ollama

import (
	"encoding/json"
	"net/http"
)

type Client struct {
	BaseURL string
}

type ModelPSResponse struct {
	Models []struct {
		Name string `json:"name"`
		Size int64  `json:"size"`
	} `json:"models"`
}

func (c *Client) GetRunningModels() (*ModelPSResponse, error) {
	resp, err := http.Get(c.BaseURL + "/api/ps")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var result ModelPSResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	return &result, err
}
