package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/remram44/vogon/internal/database"
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

type jsonMessage struct {
	Message string `json:"message"`
}

func getError(response *http.Response) error {
	contentType := response.Header.Get("Content-type")
	log.WithFields(log.Fields{
		"status":          response.Status,
		"contentTypeJson": contentType == "application/json",
	}).Print("error from server")
	if contentType == "application/json" {
		decoder := json.NewDecoder(response.Body)
		var result jsonMessage
		err := decoder.Decode(&result)
		if err == nil {
			return fmt.Errorf("error from server: %v", result.Message)
		}
	}
	return fmt.Errorf("error: %v", response.Status)
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
	if response.StatusCode != 200 {
		return "", getError(response)
	}
	decoder := json.NewDecoder(response.Body)
	var result serverVersion
	err = decoder.Decode(&result)
	if err != nil {
		return "", fmt.Errorf("parsing version response: %w", err)
	}

	return result.Version, nil
}

func (c *Client) GetObject(name string) (*database.Object, error) {
	request, err := http.NewRequest("GET", c.uri+"/"+name, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Accept", "application/json")
	response, err := c.httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("getting object: %w", err)
	}
	if response.StatusCode == 404 {
		return nil, nil
	}
	if response.StatusCode != 200 {
		return nil, getError(response)
	}
	decoder := json.NewDecoder(response.Body)
	var result database.Object
	err = decoder.Decode(&result)
	if err != nil {
		return nil, fmt.Errorf("parsing object: %w", err)
	}

	return &result, nil
}

type WriteMode int

const (
	CreateOrReplace WriteMode = iota
	Create
	Replace
)

type bufferCloser struct {
	bytes.Buffer
}

func (b bufferCloser) Close() {
}

func (c *Client) WriteObject(object database.Object, mode WriteMode) (database.MetadataResponse, error) {
	var result database.MetadataResponse

	uri := c.uri + "/" + object.Metadata.Name
	switch mode {
	case CreateOrReplace:
	case Create:
		uri += "?replace=false"
	case Replace:
		uri += "?create=false"
	}
	request, err := http.NewRequest("PUT", uri, nil)
	if err != nil {
		return result, err
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-type", "application/json")

	var body bytes.Buffer
	encoder := json.NewEncoder(&body)
	encoder.Encode(object)
	request.Body = io.NopCloser(&body)

	response, err := c.httpClient.Do(request)
	if err != nil {
		return result, fmt.Errorf("sending object: %w", err)
	}
	if response.StatusCode != 200 {
		return result, getError(response)
	}
	decoder := json.NewDecoder(response.Body)
	err = decoder.Decode(&result)
	if err != nil {
		return result, fmt.Errorf("parsing response: %w", err)
	}

	return result, nil
}
