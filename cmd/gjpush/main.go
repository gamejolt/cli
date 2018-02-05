package main

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/droundy/goopt"
	"github.com/gamejolt/cli/pkg/api"
	"github.com/gamejolt/cli/pkg/api/models"
	"github.com/gamejolt/cli/pkg/fs"
	"github.com/gamejolt/cli/pkg/project"

	"gopkg.in/blang/semver.v3"
)

var inReader = bufio.NewReader(os.Stdin)

func main() {
	tokenArg := goopt.StringWithLabel([]string{"-t", "--token"}, "", "TOKEN", "Your service API authentication token")
	gameIDArg := goopt.IntWithLabel([]string{"-g", "--game"}, 0, "GAME", "The game ID")
	packageIDArg := goopt.IntWithLabel([]string{"-p", "--port"}, 0, "PACKAGE", "The package ID")
	releaseVersionArg := goopt.StringWithLabel([]string{"-r", "--release"}, "", "RELEASE", "The release version to attach the build file to[1]")
	browserArg := goopt.Flag([]string{"-b", "--browser"}, []string{}, "Upload as a browser build", "")

	versionArg := goopt.Flag([]string{"-v", "--version"}, []string{}, "Displays the version", "")

	goopt.Version = project.Version
	goopt.Summary = project.Name + " [options] FILE"
	oldUsage := goopt.Usage
	goopt.Usage = func() string {
		return oldUsage() + `
Notes:
  [1] Semver compatible release version. If the specified game doesn't have this release yet, it will be created.
`
	}
	goopt.Parse(nil)

	if *versionArg {
		PrintAndExit(project.Name, goopt.Version)
	}

	// If no arguments are given, simply display the help
	if len(goopt.Args) == 0 {
		PrintHelpAndExit()
	}

	filepathArg := goopt.Args[0]
	if filepathArg == "help" {
		PrintHelpAndExit()
	}

	apiClient, user, game, gamePackage, releaseSemver, filepath, err := GetParams(*tokenArg, *gameIDArg, *packageIDArg, *releaseVersionArg, filepathArg)
	if err != nil {
		log.Println(err.Error())
		os.Exit(1)
	}

	err = Upload(apiClient, game, gamePackage, releaseSemver, *browserArg, filepath)
	if err != nil {
		log.Println(err.Error())
		os.Exit(1)
	}

	fmt.Printf("Token: %s\nUser ID: %d\nGame ID: %d\nPackage ID: %d\nRelease Version: %s\nFile: %s\n", apiClient.Token(), user.ID, game.ID, gamePackage.ID, releaseSemver.String(), filepath)
}

// PrintHelpAndExit prints the help usage and exits with code 0
func PrintHelpAndExit() {
	PrintAndExit(goopt.Usage())
}

// PrintAndExit prints something using log.Println and exits with code 0
func PrintAndExit(a ...interface{}) {
	fmt.Println(a...)
	os.Exit(0)
}

// GetParams gets the parsed parameters, prompts for missing ones, validates, and returns them if they are valid
func GetParams(token string, gameID, packageID int, releaseVersion, path string) (*api.Client, *models.User, *models.Game, *models.GamePackage, *semver.Version, string, error) {
	apiClient, user, err := Authenticate(token)
	if err != nil {
		return nil, nil, nil, nil, nil, "", err
	}

	token = apiClient.Token()

	fmt.Printf("Hello, %s\n\n", user.Username)
	game, err := GetGame(apiClient, gameID)
	if err != nil {
		return nil, nil, nil, nil, nil, "", err
	}

	gamePackage, err := GetGamePackage(apiClient, game.ID, packageID)
	if err != nil {
		return nil, nil, nil, nil, nil, "", err
	}

	releaseSemver, err := GetGameRelease(apiClient, releaseVersion)
	if err != nil {
		return nil, nil, nil, nil, nil, "", err
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, nil, nil, nil, nil, "", err
	}
	defer file.Close()

	return apiClient, user, game, gamePackage, releaseSemver, path, nil
}

// Authenticate uses the given token to authenticate the user.
// On a successful authentication, an API client will be returned for use in the rest of the lifetime of the program.
// If auth token is not given, it will be prompted.
func Authenticate(token string) (*api.Client, *models.User, error) {
	// Prompt for the auth token if not given
	if token == "" {
		fmt.Print("Enter your authentication token: ")
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
		fmt.Print("Enter a game ID: ")
		gameIDStr, err := inReader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		gameID, err = strconv.Atoi(strings.TrimSpace(gameIDStr))
		if err != nil {
			return nil, err
		}
	}

	game, err := apiClient.Game(gameID)
	if err != nil {
		return nil, err
	}

	if game == nil {
		return nil, errors.New("No such game for your account")
	}

	return game, nil
}

// GetGamePackage gets and validates a game package by a given id. If the id is not given, it will be prompted.
func GetGamePackage(apiClient *api.Client, gameID, packageID int) (*models.GamePackage, error) {
	if gameID == 0 {
		return nil, errors.New("Game ID must be provided")
	}

	if packageID == 0 {
		fmt.Print("Enter a game package ID: ")
		packageIDStr, err := inReader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		packageID, err = strconv.Atoi(strings.TrimSpace(packageIDStr))
		if err != nil {
			return nil, err
		}
	}

	gamePackage, err := apiClient.GamePackage(packageID)
	if err != nil {
		return nil, err
	}

	if gamePackage == nil {
		return nil, errors.New("No such package for this game")
	}

	return gamePackage, nil
}

// GetGameRelease gets and validates a game release by a given release version.
// If the release version is not given, it will be prompted.
func GetGameRelease(apiClient *api.Client, releaseVersion string) (*semver.Version, error) {
	if releaseVersion == "" {
		fmt.Print("Enter a release version (1.2.3): ")
		var err error
		releaseVersion, err = inReader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		releaseVersion = strings.TrimSpace(releaseVersion)
	}

	semver, err := semver.Make(releaseVersion)
	if err != nil {
		return nil, err
	}
	return &semver, nil
}

// Upload uploads a file to a game
func Upload(apiClient *api.Client, game *models.Game, gamePackage *models.GamePackage, releaseSemver *semver.Version, browserBuild bool, filepath string) error {
	filesize, err := fs.Filesize(filepath)
	if err != nil {
		return err
	}

	checksum, err := MD5File(filepath)
	if err != nil {
		return err
	}

	result, err := apiClient.FileAdd(game.ID, gamePackage.ID, releaseSemver, !browserBuild, filesize, checksum, false, filepath)
	if err != nil {
		return err
	}

	if result == nil {
		return errors.New("Failed to upload file")
	}

	bytes, _ := json.Marshal(result)
	fmt.Println("Uploaded file:", string(bytes))
	return nil
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
