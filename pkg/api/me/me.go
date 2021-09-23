package me

import (
	"encoding/json"
	"errors"
	"io/ioutil"

	apiErrors "github.com/gamejolt/cli/pkg/api/errors"
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
		return nil, errors.New("Failed to authenticate: " + err.Error())
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, errors.New("Failed to authenticate: " + err.Error())
	}

	result := &Result{}
	if err = json.Unmarshal(body, result); err != nil {
		return nil, errors.New("Failed to authenticate, the server returned a weird looking response: " + string(body))
	}

	if result.Error != nil {
		return nil, apiErrors.New(result.Error)
	}

	return result.User, nil
}
