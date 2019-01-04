package main

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/howeyc/gopass"
	homedir "github.com/mitchellh/go-homedir"

	"github.com/gamejolt/cli/config"
	"github.com/gamejolt/cli/pkg/api"
	"github.com/gamejolt/cli/pkg/api/files"
	"github.com/gamejolt/cli/pkg/api/models"
	"github.com/gamejolt/cli/pkg/api/packages"
	"github.com/gamejolt/cli/pkg/fs"
	_io "github.com/gamejolt/cli/pkg/io"
	"github.com/gamejolt/cli/pkg/project"
	"github.com/gamejolt/cli/pkg/ui"

	semver "gopkg.in/blang/semver.v3"
	pb "gopkg.in/cheggaaa/pb.v1"
	color "gopkg.in/fatih/color.v1"
	flags "gopkg.in/jessevdk/go-flags.v1"
)

var inReader = bufio.NewReader(os.Stdin)

func main() {
	color.Unset()
	defer color.Unset()

	opts, err := ParseOptions()
	if err != nil {
		ErrorAndExit("%s\n", err.Error())
	}

	// Help and version are printed in the parse options command, we should be able to just quit here
	if opts.Help || opts.Version {
		Exit(0)
	}

	apiClient, game, gamePackage, releaseSemver, filepath, filesize, checksum, fileStatus, chunkSize, err := GetParams(opts)
	if err != nil {
		ErrorAndExit("%s\n", err.Error())
	}

	if fileStatus.Status == "new" {
		ui.Info("Starting a new upload ...\n")
	} else if fileStatus.Status == "partial" {
		ui.Info("Resuming the upload (File ID: %d) ...\n", fileStatus.FileID)
	} else if fileStatus.Status == "error" {
		ui.Warn("There was an issue with the previous upload chunk, we have to start over :(\n")
		ui.Info("Starting a new upload ...\n")
	}

	err = Upload(apiClient, game, gamePackage, releaseSemver, opts.IsBrowser, filepath, filesize, checksum, fileStatus.Start, chunkSize)
	if err != nil {
		ErrorAndExit("%s\n", err.Error())
	}
	ui.Success("Upload complete :D\n")
}

// Options is the command line options struct
type Options struct {
	Token          string `short:"t" long:"token" value-name:"TOKEN" description:"Your service API authentication token"`
	GameID         int
	GameIDStr      string `short:"g" long:"game" value-name:"GAME-ID" description:"The game ID"`
	PackageID      int
	PackageIDStr   string `short:"p" long:"package" value-name:"PACKAGE" description:"The package ID"`
	ReleaseVersion string `short:"r" long:"release" value-name:"VERSION" description:"The release version to attach the build file to[1]"`
	IsBrowser      bool   `short:"b" long:"browser" description:"Upload a browser build. By default uploads a desktop build."`
	Advanced       struct {
		ChunkSize int `long:"chunk-size" value-name:"MB" description:"How big should the chunks the CLI uploads be."`
	} `group:"Advanced Options"`
	Help    bool `short:"h" long:"help" description:"Show this help message"`
	Version bool `short:"v" long:"version" description:"Display the version"`
	Args    struct {
		File string `positional-arg-name:"FILE" description:"The file to upload"`
	} `positional-args:"1" required:"1"`
}

// Credentials is the structure of the credentials file used to fetch the token from if none is specified
type Credentials struct {
	Token string `json:"token"`
}

// ParseOptions parses the command line options
// If the program should stop after
func ParseOptions() (*Options, error) {
	opts := &Options{}
	parser := flags.NewParser(opts, flags.PassDoubleDash)
	parser.Usage += "[OPTIONS]"

	optStrings, err := parser.Parse()
	err = findHelpOrVersionFlags(opts, optStrings, err)

	if err != nil {
		ui.Error("Oh no, %s!\n\n", err.Error())
		opts.Help = true
	}

	// If we got passed a help/version flag, we dont care if the rest of the arguments are invalid,
	// because we'll only print the help/version - so we early out here.
	if opts.Help {
		PrintHelp(parser)
		return opts, nil
	}

	if opts.Version {
		PrintVersion()
		return opts, nil
	}

	if len(optStrings) > 0 {
		return nil, errors.New("Too many arguments! Maybe you need to escape the file name if it contains spaces?")
	}

	// If token is not specified, attempt getting it from an environment variable or a credentials file
	if opts.Token == "" {
		opts.Token = getTokenFallback()
	}

	if opts.GameIDStr != "" {
		gameID, err := strconv.Atoi(opts.GameIDStr)
		if err != nil || gameID < 1 {
			return nil, errors.New("Oh no, invalid game ID - expected a positive integer")
		}
		opts.GameID = gameID
	}

	if opts.PackageIDStr != "" {
		packageID, err := strconv.Atoi(opts.PackageIDStr)
		if err != nil || packageID < 1 {
			return nil, errors.New("Oh no, invalid package ID - expected a positive integer")
		}
		opts.PackageID = packageID
	}

	return opts, nil
}

func findHelpOrVersionFlags(opts *Options, optStrings []string, err error) error {
	if err == nil || optStrings == nil {
		return nil
	}

	// Even if there are errors, see if something resembling a help or version flag was passed in.
	// In these cases we want to silence the error and just output them right away.
	for _, opt := range optStrings {
		if opt == "-h" || opt == "--help" || opt == "/h" || opt == "/help" || opt == "/?" {
			opts.Help = true
			return nil
		}
		if opt == "-v" || opt == "--version" || opt == "/v" || opt == "/version" {
			opts.Version = true
			return nil
		}
	}

	return err
}

// PrintVersion prints the version
func PrintVersion() {
	fmt.Printf("%s %s\n", project.Name, project.Version)
}

// PrintHelp prints the help
func PrintHelp(parser *flags.Parser) {
	parser.WriteHelp(os.Stdout)
	fmt.Println("\n" +
		"Notes:\n" +
		"  [1] Semver compatible release version. If the specified game doesn't have this release yet, it will be created.")
}

// ErrorAndExit prints an string with error formatting and exits with code 1
func ErrorAndExit(str string, a ...interface{}) {
	ui.Error(str, a...)
	Exit(1)
}

// Exit exits the program
func Exit(code int) {
	color.Unset()
	os.Exit(code)
}

func getTokenFallback() (token string) {
	// Attempt to get the token from the GJPUSH_TOKEN environment variable
	token = os.Getenv("GJPUSH_TOKEN")
	if token != "" {
		ui.Info("Using token from `GJPUSH_TOKEN` environment variable\n")
		return
	}

	// Attempt to get the token from the ~/.gj/credentials.json file
	if dir, err := homedir.Dir(); err == nil {
		credentialsFile := filepath.Join(dir, ".gj", "credentials.json")
		if bytes, err := ioutil.ReadFile(credentialsFile); err == nil {
			creds := &Credentials{}
			if err = json.Unmarshal(bytes, creds); err == nil {
				ui.Info("Using token from credentials file\n")
				token = creds.Token
				return
			}

			ui.Warn("Attempted to get token credentials file (%s), but the file is malformed: %s\n", credentialsFile, err.Error())
		}
	}

	return
}

func getFileData(path string) (int64, string, error) {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			err = errors.New("File doesn't exist")
		} else if os.IsPermission(err) {
			err = errors.New("No permission to read the file")
		}
		return 0, "", err
	}
	file.Close()

	filesize, err := fs.Filesize(path)
	if err != nil {
		return 0, "", errors.New("Failed to determine filesize for some reason")
	}

	checksum, err := md5File(path, filesize)
	if err != nil {
		return 0, "", errors.New("Failed to calculate checksum for the file.\nHas it changed while I was running?")
	}

	return filesize, checksum, nil
}

func md5File(path string, filesize int64) (string, error) {
	hash := md5.New()

	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	isSlow := false
	_, err = _io.CopyWithSlowBar(hash, file, 2*time.Second, func() *pb.ProgressBar {
		isSlow = true
		ui.Info("Calculating checksum...\n")
		return pb.New64(filesize).SetMaxWidth(80).SetUnits(pb.U_BYTES_DEC)
	})

	if isSlow {
		ui.Info("\n")
	}

	if err != nil {
		return "", err
	}

	var result []byte
	result = hash.Sum(result)
	return hex.EncodeToString(result), nil
}

// GetParams gets the parsed parameters, prompts for missing ones, validates, and returns them if they are valid
func GetParams(opts *Options) (*api.Client, *models.Game, *models.GamePackage, *semver.Version, string, int64, string, *files.GetResult, int64, error) {
	path := opts.Args.File
	filesize, checksum, err := getFileData(path)
	if err != nil {
		return nil, nil, nil, nil, "", 0, "", nil, 0, err
	}

	apiClient, user, err := Authenticate(opts.Token)
	if err != nil {
		return nil, nil, nil, nil, "", 0, "", nil, 0, err
	}

	ui.Success("Hello, %s\n\n", user.Username)
	game, err := GetGame(apiClient, opts.GameID)
	if err != nil {
		return nil, nil, nil, nil, "", 0, "", nil, 0, err
	}

	gamePackage, err := GetGamePackage(apiClient, game.ID, opts.PackageID)
	if err != nil {
		return nil, nil, nil, nil, "", 0, "", nil, 0, err
	}

	releaseSemver, err := GetGameRelease(apiClient, opts.ReleaseVersion)
	if err != nil {
		return nil, nil, nil, nil, "", 0, "", nil, 0, err
	}

	fileStatus, err := apiClient.FileStatus(game.ID, filesize, checksum)
	if err != nil {
		return nil, nil, nil, nil, "", 0, "", nil, 0, err
	}

	chunkSize := int64(opts.Advanced.ChunkSize) * 1024 * 1024
	if chunkSize <= 0 {
		chunkSize = config.ChunkSize
	}

	return apiClient, game, gamePackage, releaseSemver, path, filesize, checksum, fileStatus, chunkSize, nil
}

// Authenticate uses the given token to authenticate the user.
// On a successful authentication, an API client will be returned for use in the rest of the lifetime of the program.
// If auth token is not given, it will be prompted.
func Authenticate(token string) (*api.Client, *models.User, error) {
	// Prompt for the auth token if not given
	if token == "" {
		ui.Prompt("Enter your authentication token: ")
		color.Unset()
		var err error

		tokenBytes, err := gopass.GetPasswd()
		if err != nil {
			return nil, nil, err
		}
		token = strings.TrimSpace(string(tokenBytes))
	}

	// Validate it
	apiClient := api.NewClient(token)
	user, err := apiClient.Me()
	if err != nil {
		return nil, nil, err
	}

	return apiClient, user, nil
}

// GetGame gets and validates a game by a given id. If the id is not given, it will be prompted.
func GetGame(apiClient *api.Client, gameID int) (*models.Game, error) {
	if gameID == 0 {
		ui.Prompt("Enter a game ID: ")
		gameIDStr, err := inReader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		gameID, err = strconv.Atoi(strings.TrimSpace(gameIDStr))
		if err != nil || gameID < 1 {
			return nil, errors.New("Invalid game ID - expected a positive integer")
		}
	}

	game, err := apiClient.Game(gameID)
	if err != nil {
		return nil, err
	}

	return game, nil
}

// GetGamePackage gets and validates a game package by a given id. If the id is not given, it will be prompted.
func GetGamePackage(apiClient *api.Client, gameID, packageID int) (*models.GamePackage, error) {
	if gameID == 0 {
		return nil, errors.New("Game ID must be provided")
	}

	if packageID == 0 {
		ui.Prompt("Enter a game package ID: ")
		packageIDStr, err := inReader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		packageID, err = strconv.Atoi(strings.TrimSpace(packageIDStr))
		if err != nil || packageID < 1 {
			return nil, errors.New("Invalid package ID - expected a positive integer")
		}
	}

	gamePackage, err := apiClient.GamePackage(packageID, &packages.GetOptions{GameID: gameID})
	if err != nil {
		return nil, err
	}

	return gamePackage, nil
}

// GetGameRelease gets and validates a game release by a given release version.
// If the release version is not given, it will be prompted.
func GetGameRelease(apiClient *api.Client, releaseVersion string) (*semver.Version, error) {
	if releaseVersion == "" {
		ui.Prompt("Enter a release version (1.2.3): ")
		var err error
		releaseVersion, err = inReader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		releaseVersion = strings.TrimSpace(releaseVersion)
	}

	semver, err := semver.Make(releaseVersion)
	if err != nil {
		return nil, errors.New("Invalid semver. Check out https://semver.org")
	}
	return &semver, nil
}

// Upload uploads a file to a game
func Upload(apiClient *api.Client, game *models.Game, gamePackage *models.GamePackage, releaseSemver *semver.Version, browserBuild bool, filepath string, filesize int64, checksum string, startByte, chunkSize int64) error {
	// Create a new progress bar that starts from the given start byte
	bar := pb.New64(filesize).SetMaxWidth(80).SetUnits(pb.U_BYTES_DEC)
	bar.Add64(startByte)

	// The bar will be set to visible by the apiClient as soon as it knows it wouldn't print any errors right off the bat
	bar.ShowBar = false
	bar.ShowSpeed = true
	bar.Start()
	defer bar.Finish()

	for {
		result, err := uploadChunk(apiClient, game, gamePackage, releaseSemver, browserBuild, filepath, filesize, checksum, startByte, chunkSize, bar)
		if err != nil {
			return err
		}

		if result.Status == "complete" {
			return nil
		}

		// Get next chunk
		startByte = result.Start
	}
}

func uploadChunk(apiClient *api.Client, game *models.Game, gamePackage *models.GamePackage, releaseSemver *semver.Version, browserBuild bool, filepath string, filesize int64, checksum string, startByte, chunkSize int64, bar *pb.ProgressBar) (*files.AddResult, error) {
	result, err := apiClient.FileAdd(game.ID, gamePackage.ID, releaseSemver, !browserBuild, filesize, checksum, false, filepath, startByte, chunkSize, bar)
	if err != nil {
		return nil, err
	}

	if result.Status == "complete" {
		return result, nil
	}

	if result.Status == "error" || result.Start <= startByte {
		return nil, errors.New(`Uh oh, something went wrong!
This could happen for a couple of reasons:
    • The file changed while uploading
    • File upload has expired (no progress in the last day or so)
    • We fucked up`)
	}

	return result, nil
}
