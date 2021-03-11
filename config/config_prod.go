// +build prod

package config

// BaseURL is the base url for the service api
const BaseURL = "https://gamejolt.com"

// UploadHost is the hostname for the upload server
const UploadHost = "upload.gamejolt.com"

// ChunkSize is the chunk size a file is uploaded in.
// In production we split to 10 MB chunks.
const ChunkSize = 10 * 1024 * 1024
