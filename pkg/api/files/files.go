package files

import (
	"encoding/json"
	"errors"
	"io"
	"net/url"
	"strconv"

	"github.com/gamejolt/cli/config"
	modelErrors "github.com/gamejolt/cli/pkg/api/errors"
	"github.com/gamejolt/cli/pkg/api/models"
	cliHttp "github.com/gamejolt/cli/pkg/http"
	customIO "github.com/gamejolt/cli/pkg/io"

	semver "gopkg.in/blang/semver.v3"
	"gopkg.in/cheggaaa/pb.v1"
)

// GetResult is the result from the /files endpoint
type GetResult struct {
	Status string        `json:"status"`
	FileID int           `json:"file_id,omitempty"` // Returned for partial/in progress file uploads
	Start  int64         `json:"start,omitempty"`
	Error  *models.Error `json:"error,omitempty"`
}

// AddResult is the result from the /files/add endpoint
type AddResult struct {
	GetResult
	Build *models.GameBuild `json:"build,omitempty"` // Only returned once when the file has been fully uploaded
}

func formatBool(b bool) string {
	if b {
		return "1"
	}
	return "0"
}

// Get sends a new GET /files/add request
func Get(client *cliHttp.SimpleClient, gameID int, size int64, checksum string) (*GetResult, error) {
	getParams := url.Values(map[string][]string{
		"game_id":  []string{strconv.Itoa(gameID)},
		"size":     []string{strconv.FormatInt(size, 10)},
		"checksum": []string{checksum},
	})

	_, res, err := client.Get("files/add", getParams)

	if err != nil {
		return nil, errors.New("Failed to fetch current state of file on the server: " + err.Error())
	}
	defer res.Body.Close()

	decoder := json.NewDecoder(res.Body)
	result := &GetResult{}
	if err = decoder.Decode(result); err != nil {
		return nil, errors.New("Failed to fetch current state of file on the server, the server returned a weird looking response")
	}

	if result.Error != nil {
		return nil, modelErrors.New(result.Error)
	}
	return result, nil
}

// Add sends a new POST /files/add request
func Add(client *cliHttp.SimpleClient, gameID, packageID int, releaseVersion *semver.Version, isDownloadable bool, size int64, checksum string, forceRestart bool, filepath string, startByte int64, bar *pb.ProgressBar) (*AddResult, error) {
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
		offset, err := src.File.Seek(startByte, 0)
		if err != nil || offset != startByte {
			return 0, errors.New("Failed to seek the file, has it changed while I was running?")
		}

		// Read only the wanted chunk size
		reader := io.LimitReader(src.File, config.ChunkSize)

		// Limit the upload speed in development for testing
		reader = customIO.NewReader(reader)

		bar.ShowBar = true
		reader = bar.NewProxyReader(reader)

		return io.Copy(dst, reader)
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
