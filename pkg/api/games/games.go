package games

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"strconv"

	apiErrors "github.com/gamejolt/cli/pkg/api/errors"
	"github.com/gamejolt/cli/pkg/api/models"
	cliHttp "github.com/gamejolt/cli/pkg/http"
)

// GetResult is the payload from the `/games/:id` endpoint
type GetResult struct {
	Game  *models.Game  `json:"game"`
	Error *models.Error `json:"error,omitempty"`
}

// ListResult is the payload from the `/games` endpoint
type ListResult struct {
	Games *Games        `json:"games"`
	Error *models.Error `json:"error,omitempty"`
}

// Games is a list of games as returned by the /games endpoint
type Games struct {
	Games   []models.Game `json:"data"`
	Page    int           `json:"page"`
	PerPage int           `json:"per_page"`
	Total   int           `json:"total"`
}

// Get sends a new /games/:gameId request
func Get(client *cliHttp.SimpleClient, gameID int) (*models.Game, error) {
	_, res, err := client.Get("games/"+strconv.Itoa(gameID), nil)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, errors.New("Failed to fetch information about game: " + err.Error())
	}

	result := &GetResult{}
	if err = json.Unmarshal(body, result); err != nil {
		return nil, errors.New("Failed to fetch information about game, the server returned a weird looking response: " + string(body))
	}

	if result.Error != nil {
		return nil, apiErrors.New(result.Error)
	}
	return result.Game, nil
}

// List sends a new /games request
func List(client *cliHttp.SimpleClient) (*Games, error) {
	_, res, err := client.Get("games", nil)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, errors.New("Failed to list games: " + err.Error())
	}

	result := &ListResult{}
	if err = json.Unmarshal(body, result); err != nil {
		return nil, errors.New("Failed to list games, the server returned a weird looking response: " + string(body))
	}

	if result.Error != nil {
		return nil, apiErrors.New(result.Error)
	}
	return result.Games, nil
}
