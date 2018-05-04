package releases

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/gamejolt/cli/pkg/api/errors"
	"github.com/gamejolt/cli/pkg/api/models"
	cliHttp "github.com/gamejolt/cli/pkg/http"
)

// ListBuildsResult is the payload from the `/releases/builds/:id` endpoint
type ListBuildsResult struct {
	Builds *Builds       `json:"builds"`
	Error  *models.Error `json:"error,omitempty"`
}

// ListBuildsOptions are additional optional parameters for the `/releases/builds/:id` endpoint`
type ListBuildsOptions struct {
	GameID    int
	PackageID int
}

// Builds is a list of builds as returned by the /releases/builds/:id endpoint
type Builds struct {
	Builds  []models.GameBuild `json:"data"`
	Page    int                `json:"page"`
	PerPage int                `json:"per_page"`
	Total   int                `json:"total"`
}

// List sends a new /releases/builds/:id request
func List(client *cliHttp.SimpleClient, releaseID int, options *ListBuildsOptions) (*Builds, error) {
	var getParams url.Values
	if options != nil {
		getMap := map[string][]string{}
		if options.GameID != 0 {
			getMap["game_id"] = []string{strconv.Itoa(options.GameID)}
			getMap["package_id"] = []string{strconv.Itoa(options.PackageID)}
		}
		getParams = url.Values(getMap)
	}

	_, res, err := client.Get(fmt.Sprintf("releases/builds/%d", releaseID), getParams)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	decoder := json.NewDecoder(res.Body)
	result := &ListBuildsResult{}
	if err = decoder.Decode(result); err != nil {
		return nil, err
	}

	if result.Error != nil {
		return nil, errors.New(result.Error)
	}
	return result.Builds, nil
}
