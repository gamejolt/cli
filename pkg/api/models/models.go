package models

// Error is an error model. It is returned for any error that may occur with api calls.
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`

	// Optional fields
	Fields    []string `json:"fields,omitempty"`     // Returned for invalid fields error
	HTTPError *int     `json:"http_error,omitempty"` // Returned for any other unknown errors. This is just repeating the http status code
}

// User is a user model
type User struct {
	ID          int    `json:"id"`
	URL         string `json:"url"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	CreatedOn   int64  `json:"created_on"`
}

// Game is a game model
type Game struct {
	ID          int      `json:"id"`
	URL         string   `json:"url"`
	Owner       *User    `json:"owner"`
	Title       string   `json:"title"`
	Tags        []string `json:"tags"`
	CreatedOn   int64    `json:"created_on"`
	PublishedOn int64    `json:"published_on"`
}

// GamePackage is a package model
type GamePackage struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Sort        int    `json:"sort"`
	CreatedOn   int64  `json:"created_on"`
	PublishedOn int64  `json:"published_on"`
	UpdatedOn   int64  `json:"updated_on"`
	Visibility  string `json:"visibility"`
	Status      string `json:"status"`
}

// GameRelease is a release model
type GameRelease struct {
	ID          int    `json:"id"`
	Version     string `json:"version"`
	CreatedOn   int64  `json:"created_on"`
	PublishedOn int64  `json:"published_on"`
	UpdatedOn   int64  `json:"updated_on"`
	Status      string `json:"status"`
	Sort        int    `json:"sort"`
}

// GameBuild is a build model
type GameBuild struct {
	File                     *GameBuildFile          `json:"file"`
	LaunchOptions            []GameBuildLaunchOption `json:"launch_options"`
	ArchiveType              string                  `json:"archive_type"`
	Type                     string                  `json:"type"`
	Windows                  bool                    `json:"os_windows"`
	Windows64                bool                    `json:"os_windows_64"`
	Mac                      bool                    `json:"os_mac"`
	Mac64                    bool                    `json:"os_mac_64"`
	Linux                    bool                    `json:"os_linux"`
	Linux64                  bool                    `json:"os_linux_64"`
	Other                    bool                    `json:"os_other"`
	EmulatorType             string                  `json:"emulator_type"`
	EmbedWidth               int                     `json:"embed_width"`
	EmbedHeight              int                     `json:"embed_height"`
	BrowserDisableRightClick bool                    `json:"browser_disable_right_click"`
	Errors                   string                  `json:"errors"`
	CreatedOn                int64                   `json:"created_on"`
	UpdatedOn                int64                   `json:"updated_on"`
	Status                   string                  `json:"status"`
}

// GameBuildFile is a build file model
type GameBuildFile struct {
	ID       int    `json:"id"`
	Filename string `json:"filename"`
	Filesize int64  `json:"filesize"`
}

// GameBuildLaunchOption is a launch option model
type GameBuildLaunchOption struct {
	ID             int    `json:"id"`
	OS             string `json:"os"`
	ExecutablePath string `json:"executable_path"`
}
