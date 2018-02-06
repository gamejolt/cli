// +build !prod

package config

// BaseURL is the base url for the service api
const BaseURL = "http://development.gamejolt.com"

// ChunkSize is the chunk size a file is uploaded in.
// In development we split to 5 MB chunks.
const ChunkSize = 5 * 1024 * 1024
