package client

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/remram44/vogon/internal/versioning"
)

type ClientOptions struct {
	Uri string
}

type Client struct {
	httpClient http.Client
	uri        string
}

func NewClient(options ClientOptions) (*Client, error) {
	// Remove trailing slash
	uri := options.Uri
	if len(uri) > 1 && uri[len(uri)-1] == '/' {
		uri = uri[:len(uri)-1]
	}
	client := &Client{
		uri: uri,
	}
	version, err := client.GetVersion()
	if err != nil {
		return nil, fmt.Errorf("Ping server: %w", err)
	}
	if version != versioning.NameAndVersionString() {
		return nil, fmt.Errorf("Unsupported version")
	}
	return client, nil
}

type serverVersion struct {
	Version string `json:"version"`
}

func (c *Client) GetVersion() (string, error) {
	request, err := http.NewRequest("GET", c.uri+"/_version", nil)
	if err != nil {
		return "", err
	}
	request.Header.Set("Accept", "application/json")
	response, err := c.httpClient.Do(request)
	if err != nil {
		return "", fmt.Errorf("getting version: %w", err)
	}
	decoder := json.NewDecoder(response.Body)
	var result serverVersion
	err = decoder.Decode(&result)
	if err != nil {
		return "", fmt.Errorf("parsing version response: %w", err)
	}

	return result.Version, nil
}
