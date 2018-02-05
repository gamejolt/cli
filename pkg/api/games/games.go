package games

import (
	"encoding/json"
	"strconv"

	"github.com/gamejolt/cli/pkg/api/errors"
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

	decoder := json.NewDecoder(res.Body)
	result := &GetResult{}
	if err = decoder.Decode(result); err != nil {
		return nil, err
	}

	if result.Error != nil {
		return nil, errors.New(result.Error)
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

	decoder := json.NewDecoder(res.Body)
	result := &ListResult{}
	if err = decoder.Decode(result); err != nil {
		return nil, err
	}

	if result.Error != nil {
		return nil, errors.New(result.Error)
	}
	return result.Games, nil
}
