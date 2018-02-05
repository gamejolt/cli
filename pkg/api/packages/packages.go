package packages

import (
	"encoding/json"
	"fmt"

	"github.com/gamejolt/cli/pkg/api/errors"
	"github.com/gamejolt/cli/pkg/api/models"
	cliHttp "github.com/gamejolt/cli/pkg/http"
)

// GetResult is the payload from the `/packages/:id` endpoint
type GetResult struct {
	Package *models.GamePackage `json:"package"`
	Error   *models.Error       `json:"error,omitempty"`
}

// Get sends a new /packages/:packageId request
func Get(client *cliHttp.SimpleClient, packageID int) (*models.GamePackage, error) {
	_, res, err := client.Get(fmt.Sprintf("packages/%d", packageID), nil)
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
	return result.Package, nil
}
