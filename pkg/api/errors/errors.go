package errors

import "github.com/gamejolt/cli/pkg/api/models"

const (
	// MissingAuthorization is the error code for when the authorization header is missing in the api request.
	MissingAuthorization = 1

	// InvalidAuthorization is the error code for when the authorization is invalid.
	InvalidAuthorization = 2

	// InvalidFields is the error code for when one or more request fields are either missing or invalid.
	InvalidFields = 3

	// CantCleanTempFile is the error code for when the temporary file could not be cleaned on Game Jolt's servers.
	// This is an issue on GJ's end.
	CantCleanTempFile = 4

	// CantAttachBuild is the error code for when a build cannot be attached to the given package.
	// This can happen for many reason, check the error message for more info.
	CantAttachBuild = 5

	// UnknownError is the error code for any other unspecified error.
	UnknownError = 1000
)

// Error is an error type returned by the API calls
type Error struct {
	err *models.Error
}

// New creates a new API error
func New(err *models.Error) *Error {
	return &Error{err}
}

func (e *Error) Error() string {
	return e.err.Message
}

// Code returns the error code
func (e *Error) Code() int {
	return e.err.Code
}

// Fields returns the invalid or missing fields for the request
func (e *Error) Fields() []string {
	return e.err.Fields
}
