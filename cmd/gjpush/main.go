package main

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/gamejolt/cli/pkg/api"
	"github.com/gamejolt/cli/pkg/api/files"
	"github.com/gamejolt/cli/pkg/api/models"
	"github.com/gamejolt/cli/pkg/api/packages"
	"github.com/gamejolt/cli/pkg/fs"
	"github.com/gamejolt/cli/pkg/project"
	"github.com/gamejolt/cli/pkg/ui"

	"gopkg.in/blang/semver.v3"
	"gopkg.in/cheggaaa/pb.v1"
	"gopkg.in/jessevdk/go-flags.v1"
)

var inReader = bufio.NewReader(os.Stdin)

// Options is the command line options struct
type Options struct {
	Token          string `short:"t" long:"token" value-name:"TOKEN" description:"Your service API authentication token"`
	GameID         string `short:"g" long:"game" value-name:"GAME-ID" description:"The game ID"`
	PackageID      string `short:"p" long:"package" value-name:"PACKAGE" description:"The package ID"`
	ReleaseVersion string `short:"r" long:"release" value-name:"VERSION" description:"The release version to attach the build file to[1]" long-description:"[1] Semver compatible release version. If the specified game doesn't have this release yet, it will be created."`
	IsBrowser      bool   `short:"b" long:"browser" description:"Upload as a browser build"`
	Help           bool   `short:"h" long:"help" description:"Show this help message"`
	Version        bool   `short:"v" long:"version" description:"Display the version"`
	Args           struct {
		File string `positional-arg-name:"FILE" description:"The file to upload"`
	} `positional-args:"1" required:"1"`
}

func main() {
	opts := &Options{}
	parser := flags.NewParser(opts, flags.PassDoubleDash)
	parser.Usage += "[OPTIONS]"
	optStrings, err := parser.Parse()
	if err != nil {
		for _, opt := range optStrings {
			if opt == "-h" || opt == "--help" {
				PrintHelp(parser)
			}
			if opt == "-v" || opt == "--version" {
				PrintVersion()
			}
		}

		ui.Error("Oh no, %s!\n\n", err.Error())
		PrintHelp(parser)
	}

	if opts.Help {
		PrintHelp(parser)
	}

	if opts.Version {
		PrintVersion()
	}

	if len(optStrings) > 0 {
		ErrorAndExit("Too many arguments! Maybe you need to escape the file name if it contains spaces?\n")
	}

	gameID := 0
	if opts.GameID != "" {
		gameID, err = strconv.Atoi(opts.GameID)
		if err != nil || gameID < 1 {
			ErrorAndExit("Oh no, invalid game ID - expected a positive integer\n")
		}
	}

	packageID := 0
	if opts.PackageID != "" {
		packageID, err = strconv.Atoi(opts.PackageID)
		if err != nil || packageID < 1 {
			ErrorAndExit("Oh no, invalid package ID - expected a positive integer\n")
		}
	}

	apiClient, game, gamePackage, releaseSemver, filepath, err := GetParams(opts.Token, gameID, packageID, opts.ReleaseVersion, opts.Args.File)
	if err != nil {
		ErrorAndExit("%s\n", err.Error())
	}

	filesize, err := fs.Filesize(filepath)
	if err != nil {
		ErrorAndExit("Failed to determine filesize for some reason\n")
	}

	checksum, err := MD5File(filepath)
	if err != nil {
		ErrorAndExit("Failed to calculate checksum for the file.\nHas it changed while I was running?\n")
	}

	fileStatus, err := apiClient.FileStatus(game.ID, filesize, checksum)
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

	err = Upload(apiClient, game, gamePackage, releaseSemver, opts.IsBrowser, filepath, filesize, checksum, fileStatus.Start)
	if err != nil {
		ErrorAndExit("%s\n", err.Error())
	}
	ui.Success("Upload complete :D\n")
}

// PrintVersion prints the version and exits the program
func PrintVersion() {
	PrintAndExit("%s %s\n", project.Name, project.Version)
}

// PrintHelp prints the help and exits the program
func PrintHelp(parser *flags.Parser) {
	parser.WriteHelp(os.Stdout)
	os.Exit(0)
}

// PrintAndExit prints something and exits with code 0
func PrintAndExit(str string, a ...interface{}) {
	fmt.Printf(str, a...)
	os.Exit(0)
}

// ErrorAndExit prints an string with error formatting and exits with code 1
func ErrorAndExit(str string, a ...interface{}) {
	ui.Error(str, a...)
	os.Exit(1)
}

// GetParams gets the parsed parameters, prompts for missing ones, validates, and returns them if they are valid
func GetParams(token string, gameID, packageID int, releaseVersion, path string) (*api.Client, *models.Game, *models.GamePackage, *semver.Version, string, error) {
	apiClient, user, err := Authenticate(token)
	if err != nil {
		return nil, nil, nil, nil, "", err
	}

	token = apiClient.Token()

	ui.Success("Hello, %s\n\n", user.Username)
	game, err := GetGame(apiClient, gameID)
	if err != nil {
		return nil, nil, nil, nil, "", err
	}

	gamePackage, err := GetGamePackage(apiClient, game.ID, packageID)
	if err != nil {
		return nil, nil, nil, nil, "", err
	}

	releaseSemver, err := GetGameRelease(apiClient, releaseVersion)
	if err != nil {
		return nil, nil, nil, nil, "", err
	}

	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			err = errors.New("File doesn't exist")
		} else if os.IsPermission(err) {
			err = errors.New("No permission to read the file")
		}
		return nil, nil, nil, nil, "", err
	}
	defer file.Close()

	return apiClient, game, gamePackage, releaseSemver, path, nil
}

// Authenticate uses the given token to authenticate the user.
// On a successful authentication, an API client will be returned for use in the rest of the lifetime of the program.
// If auth token is not given, it will be prompted.
func Authenticate(token string) (*api.Client, *models.User, error) {
	// Prompt for the auth token if not given
	if token == "" {
		ui.Prompt("Enter your authentication token: ")
		var err error
		token, err = inReader.ReadString('\n')
		if err != nil {
			return nil, nil, err
		}
		token = strings.TrimSpace(token)
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
func Upload(apiClient *api.Client, game *models.Game, gamePackage *models.GamePackage, releaseSemver *semver.Version, browserBuild bool, filepath string, filesize int64, checksum string, startByte int64) error {
	// Create a new progress bar that starts from the given start byte
	bar := pb.New64(filesize).SetUnits(pb.U_BYTES)
	bar.Add64(startByte)

	// The bar will be set to visible by the apiClient as soon as it knows it wouldn't print any errors right off the bat
	bar.ShowBar = false
	bar.Start()
	defer bar.Finish()

	for {
		result, err := uploadChunk(apiClient, game, gamePackage, releaseSemver, browserBuild, filepath, filesize, checksum, startByte, bar)
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

func uploadChunk(apiClient *api.Client, game *models.Game, gamePackage *models.GamePackage, releaseSemver *semver.Version, browserBuild bool, filepath string, filesize int64, checksum string, startByte int64, bar *pb.ProgressBar) (*files.AddResult, error) {
	result, err := apiClient.FileAdd(game.ID, gamePackage.ID, releaseSemver, !browserBuild, filesize, checksum, false, filepath, startByte, bar)
	if err != nil {
		return nil, err
	}

	if result.Status == "complete" {
		return result, nil
	}

	if result.Status == "error" || result.Start <= startByte {
		return nil, errors.New("Uh oh, something went wrong! Did the file change while I was uploading? :(")
	}

	return result, nil
}

// MD5File calculates a file's MD5. This is a blocking operation
func MD5File(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	var result []byte
	result = hash.Sum(result)
	return hex.EncodeToString(result), nil
}
