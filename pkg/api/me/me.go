package me

import (
	"encoding/json"

	"github.com/gamejolt/cli/pkg/api/errors"
	"github.com/gamejolt/cli/pkg/api/models"
	cliHttp "github.com/gamejolt/cli/pkg/http"
)

// Result is the payload from the `/me` endpoint
type Result struct {
	User  *models.User  `json:"user"`
	Error *models.Error `json:"error,omitempty"`
}

// Send sends a new /me request
func Send(client *cliHttp.SimpleClient) (*models.User, error) {
	_, res, err := client.Get("me", nil)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	decoder := json.NewDecoder(res.Body)
	result := &Result{}
	if err = decoder.Decode(result); err != nil {
		return nil, err
	}

	if result.Error != nil {
		return nil, errors.New(result.Error)
	}

	return result.User, nil
}
