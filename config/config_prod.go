// +build prod

package config

// BaseURL is the base url for the service api
const BaseURL = "https://gamejolt.com"

// ChunkSize is the chunk size a file is uploaded in.
// In production we split to 100 MB chunks.
const ChunkSize = 100 * 1024 * 1024
