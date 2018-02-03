package packages

import (
	"encoding/json"
	"fmt"

	"github.com/gamejolt/cli/pkg/api/models"
	cliHttp "github.com/gamejolt/cli/pkg/http"
)

// Get sends a new /packages/:packageId request
func Get(client *cliHttp.SimpleClient, packageID int) (*models.GamePackage, error) {
	_, res, err := client.Get(fmt.Sprintf("packages/%d", packageID), nil)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode == 404 || res.StatusCode == 403 {
		return nil, nil
	}

	decoder := json.NewDecoder(res.Body)
	result := &models.GamePackage{}
	if err = decoder.Decode(result); err != nil {
		return nil, err
	}

	return result, nil
}
