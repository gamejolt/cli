package http

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/gamejolt/cli/config"
	customIO "github.com/gamejolt/cli/pkg/io"
)

// SimpleClient is a simplified http client
type SimpleClient struct {

	// Base is the base to use for all urls. If empty, will use the given urls as is.
	Base string

	// NewRequest allows to customize which http.Request the simple client will use.
	NewRequest func(method, urlStr string, body io.Reader) (*http.Request, error)
}

// NewSimpleClient creates a new default client
func NewSimpleClient() *SimpleClient {
	client := &SimpleClient{
		Base: "",
		NewRequest: func(method, urlStr string, body io.Reader) (*http.Request, error) {
			return http.NewRequest(method, urlStr, body)
		},
	}
	return client
}

func (c *SimpleClient) send(req *http.Request) (*http.Request, *http.Response, error) {
	res, err := http.DefaultClient.Do(req)
	return req, res, err
}

func (c *SimpleClient) getURL(urlStr string, forUpload bool) (*url.URL, error) {
	base, err := url.Parse(c.Base)
	if err != nil {
		return nil, err
	}

	if forUpload {
		base.Host = config.UploadHost
	}

	u, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	return base.ResolveReference(u), nil
}

func (c *SimpleClient) buildQuery(urlStr string, get url.Values, forUpload bool) (string, error) {
	urlData, err := c.getURL(urlStr, forUpload)
	if err != nil {
		return "", err
	}

	query := urlData.Query()
	for key, values := range get {
		if len(values) == 1 {
			query.Set(key, values[0])
		} else {
			key += "[]"
			for _, value := range values {
				query.Add(key, value)
			}
		}
	}
	urlData.RawQuery = query.Encode()
	return urlData.String(), nil
}

// Get does an http get
func (c *SimpleClient) Get(urlStr string, params url.Values) (*http.Request, *http.Response, error) {
	urlStr, err := c.buildQuery(urlStr, params, false)
	if err != nil {
		return nil, nil, err
	}

	req, err := c.NewRequest("GET", urlStr, nil)
	if err != nil {
		return nil, nil, err
	}

	return c.send(req)
}

// Post does an http post of type application/json
func (c *SimpleClient) Post(urlStr string, get url.Values, post interface{}) (*http.Request, *http.Response, error) {
	urlStr, err := c.buildQuery(urlStr, get, false)
	if err != nil {
		return nil, nil, err
	}

	jsonBytes, err := json.Marshal(post)
	if err != nil {
		return nil, nil, err
	}

	req, err := c.NewRequest("POST", urlStr, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return c.send(req)
}

// MultipartFileEntry is a multipart file entry.
// It is used to map a file to it's field value and label for Multipart requests.
type MultipartFileEntry struct {
	Param string
	Path  string
	File  *os.File
}

func (c *SimpleClient) makeMultipartEntries(files map[string]string) ([]MultipartFileEntry, error) {
	fileEntries := []MultipartFileEntry{}
	for fileParam, path := range files {
		fileHandle, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		fileEntries = append(fileEntries, MultipartFileEntry{
			Param: fileParam,
			Path:  path,
			File:  fileHandle,
		})
	}
	return fileEntries, nil
}

// WriteFileFunc allows the caller to customize how the file is written.
// If this is null, a simple io.Copy will be done.
type WriteFileFunc func(dst io.Writer, src MultipartFileEntry) (int64, error)

func (c *SimpleClient) uploadMultipartEntries(fileEntries []MultipartFileEntry, reqWriter io.WriteCloser, multipartWriter *multipart.Writer, writeFileCallback WriteFileFunc) {
	for _, fileEntry := range fileEntries {
		defer fileEntry.File.Close()
	}
	defer reqWriter.Close()
	defer multipartWriter.Close()

	for _, fileEntry := range fileEntries {
		fileField, err := multipartWriter.CreateFormFile(fileEntry.Param, filepath.Base(fileEntry.Path))
		if err != nil {
			return
		}

		if writeFileCallback == nil {
			_, err = io.Copy(fileField, customIO.NewReader(fileEntry.File))
		} else {
			_, err = writeFileCallback(fileField, fileEntry)
		}
		fileEntry.File.Close()
		if err != nil {
			return
		}
	}
}

// Multipart does a multipart file upload request of type multipart/form-data
func (c *SimpleClient) Multipart(urlStr string, files map[string]string, get, post url.Values, writeFileCallback WriteFileFunc) (*http.Request, *http.Response, error) {
	urlStr, err := c.buildQuery(urlStr, get, true)
	if err != nil {
		return nil, nil, err
	}

	fileEntries, err := c.makeMultipartEntries(files)
	if err != nil {
		return nil, nil, err
	}

	reqReader, reqWriter := io.Pipe()
	multipartWriter := multipart.NewWriter(reqWriter)

	go func() {
		for key, values := range post {
			if len(values) == 1 {
				if err := multipartWriter.WriteField(key, values[0]); err != nil {
					return
				}
			} else {
				key += "[]"
				for _, value := range values {
					if err := multipartWriter.WriteField(key, value); err != nil {
						return
					}
				}
			}
		}

		c.uploadMultipartEntries(fileEntries, reqWriter, multipartWriter, writeFileCallback)
	}()

	req, err := c.NewRequest("POST", urlStr, reqReader)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Content-Type", multipartWriter.FormDataContentType())

	return c.send(req)
}
