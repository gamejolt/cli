package games

import (
	"encoding/json"
	"strconv"

	"github.com/gamejolt/cli/pkg/api/models"
	cliHttp "github.com/gamejolt/cli/pkg/http"
)

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

	if res.StatusCode == 404 || res.StatusCode == 403 {
		return nil, nil
	}

	decoder := json.NewDecoder(res.Body)
	result := &models.Game{}
	if err = decoder.Decode(result); err != nil {
		return nil, err
	}

	return result, nil
}

// List sends a new /games request
func List(client *cliHttp.SimpleClient) (*Games, error) {
	_, res, err := client.Get("games", nil)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	decoder := json.NewDecoder(res.Body)
	result := &Games{}
	if err = decoder.Decode(result); err != nil {
		return nil, err
	}

	return result, nil
}
