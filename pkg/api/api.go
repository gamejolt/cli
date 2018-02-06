package api

import (
	"io"
	"net/http"

	"github.com/gamejolt/cli/config"
	"github.com/gamejolt/cli/pkg/api/files"
	"github.com/gamejolt/cli/pkg/api/games"
	"github.com/gamejolt/cli/pkg/api/me"
	"github.com/gamejolt/cli/pkg/api/models"
	"github.com/gamejolt/cli/pkg/api/packages"
	cliHttp "github.com/gamejolt/cli/pkg/http"
	"github.com/gamejolt/cli/pkg/project"

	semver "gopkg.in/blang/semver.v3"
	"gopkg.in/cheggaaa/pb.v1"
)

// Client is a client through which http requests for the service api endpoints are made
type Client struct {
	token  string
	client *cliHttp.SimpleClient
}

// NewClient creates a new simple http client that has the given token as the authorization and a user agent to identify the cli version
func NewClient(token string) *Client {
	client := cliHttp.NewSimpleClient()
	client.Base = config.BaseURL + "/service-api/push/"
	client.NewRequest = func(method, urlStr string, body io.Reader) (*http.Request, error) {
		req, err := http.NewRequest(method, urlStr, body)
		if err != nil {
			return nil, err
		}

		req.Header.Set("Authorization", token)
		req.Header.Set("User-Agent", "GJPush/"+project.Version)

		return req, nil
	}

	return &Client{
		token,
		client,
	}
}

// Token returns the token this api client was created with
func (c *Client) Token() string {
	return c.token
}

// Me does a /me call
func (c *Client) Me() (*models.User, error) {
	return me.Send(c.client)
}

// Game does a /games/:gameId call
func (c *Client) Game(gameID int) (*models.Game, error) {
	return games.Get(c.client, gameID)
}

// Games does a /games call
func (c *Client) Games() (*games.Games, error) {
	return games.List(c.client)
}

// GamePackage does a /packages/:packageId call
func (c *Client) GamePackage(packageID int, options *packages.GetOptions) (*models.GamePackage, error) {
	return packages.Get(c.client, packageID, options)
}

// FileStatus does a GET /files/add call
func (c *Client) FileStatus(gameID int, size int64, checksum string) (*files.GetResult, error) {
	return files.Get(c.client, gameID, size, checksum)
}

// FileAdd does a POST /files/add call
func (c *Client) FileAdd(gameID, packageID int, releaseVersion *semver.Version, isDownloadable bool, size int64, checksum string, forceRestart bool, filepath string, startByte int64, bar *pb.ProgressBar) (*files.AddResult, error) {
	return files.Add(c.client, gameID, packageID, releaseVersion, isDownloadable, size, checksum, forceRestart, filepath, startByte, bar)
}
