package files

import (
	"encoding/json"
	"errors"
	"io"
	"net/url"
	"strconv"

	modelErrors "github.com/gamejolt/cli/pkg/api/errors"
	"github.com/gamejolt/cli/pkg/api/models"
	cliHttp "github.com/gamejolt/cli/pkg/http"
	customIO "github.com/gamejolt/cli/pkg/io"
	semver "gopkg.in/blang/semver.v3"
)

// AddResult is the result from the /files/add endpoint
type AddResult struct {
	Status string            `json:"status"`
	FileID int               `json:"file_id,omitempty"` // Returned for partial/in progress file uploads
	Build  *models.GameBuild `json:"build,omitempty"`   // Only returned once when the file has been fully uploaded
	Start  int64             `json:"start,omitempty"`
	Error  *models.Error     `json:"error,omitempty"`
}

func formatBool(b bool) string {
	if b {
		return "1"
	}
	return "0"
}

// Add sends a new /files/add request
func Add(client *cliHttp.SimpleClient, gameID, packageID int, releaseVersion *semver.Version, isDownloadable bool, size int64, checksum string, forceRestart bool, filepath string) (*AddResult, error) {
	getParams := url.Values(map[string][]string{
		"game_id":         []string{strconv.Itoa(gameID)},
		"package_id":      []string{strconv.Itoa(packageID)},
		"release_version": []string{releaseVersion.String()},
		"downloadable":    []string{formatBool(isDownloadable)},
		"size":            []string{strconv.FormatInt(size, 10)},
		"checksum":        []string{checksum},
		"restart":         []string{formatBool(forceRestart)},
	})

	writeFileFunc := func(dst io.Writer, src cliHttp.MultipartFileEntry) (int64, error) {
		return io.Copy(dst, customIO.NewReader(src.File))
	}

	_, res, err := client.Multipart("files/add", map[string]string{"file": filepath}, getParams, nil, writeFileFunc)

	if err != nil {
		return nil, errors.New("Failed to upload file: " + err.Error())
	}
	defer res.Body.Close()

	decoder := json.NewDecoder(res.Body)
	result := &AddResult{}
	if err = decoder.Decode(result); err != nil {
		return nil, errors.New("Failed to upload file, the server returned a weird looking response")
	}

	if result.Error != nil {
		return nil, modelErrors.New(result.Error)
	}
	return result, nil
}
