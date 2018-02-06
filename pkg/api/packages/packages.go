package packages

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"

	modelErrors "github.com/gamejolt/cli/pkg/api/errors"
	"github.com/gamejolt/cli/pkg/api/models"
	cliHttp "github.com/gamejolt/cli/pkg/http"
)

// GetResult is the payload from the `/packages/:id` endpoint
type GetResult struct {
	Package *models.GamePackage `json:"package"`
	Error   *models.Error       `json:"error,omitempty"`
}

// GetOptions are additional optional parameters for the `/packages/:id` endpoint`
type GetOptions struct {
	GameID int
}

// Get sends a new /packages/:packageId request
func Get(client *cliHttp.SimpleClient, packageID int, options *GetOptions) (*models.GamePackage, error) {
	var getParams url.Values
	if options != nil {
		getMap := map[string][]string{}
		if options.GameID != 0 {
			getMap["game_id"] = []string{strconv.Itoa(options.GameID)}
		}
		getParams = url.Values(getMap)
	}

	_, res, err := client.Get(fmt.Sprintf("packages/%d", packageID), getParams)
	if err != nil {
		return nil, errors.New("Failed to get package: " + err.Error())
	}
	defer res.Body.Close()

	decoder := json.NewDecoder(res.Body)
	result := &GetResult{}
	if err = decoder.Decode(result); err != nil {
		return nil, errors.New("Failed to get package, the server returned a weird looking response")
	}

	if result.Error != nil {
		return nil, modelErrors.New(result.Error)
	}
	return result.Package, nil
}
