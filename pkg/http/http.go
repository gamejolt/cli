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

func (c *SimpleClient) getURL(urlStr string) (*url.URL, error) {
	base, err := url.Parse(c.Base)
	if err != nil {
		return nil, err
	}

	u, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	return base.ResolveReference(u), nil
}

// Get does an http get
func (c *SimpleClient) Get(urlStr string, params url.Values) (*http.Request, *http.Response, error) {
	urlData, err := c.getURL(urlStr)
	if err != nil {
		return nil, nil, err
	}

	query := urlData.Query()
	for key, values := range params {
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
	urlStr = urlData.String()

	req, err := c.NewRequest("GET", urlStr, nil)
	if err != nil {
		return nil, nil, err
	}

	return c.send(req)
}

// Post does an http post of type application/json
func (c *SimpleClient) Post(urlStr string, get url.Values, post interface{}) (*http.Request, *http.Response, error) {
	urlData, err := c.getURL(urlStr)
	if err != nil {
		return nil, nil, err
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
	urlStr = urlData.String()

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

// WriteFileFunc allows the caller to customize how the file is written.
// If this is null, a simple io.Copy will be done.
type WriteFileFunc func(dst io.Writer, src MultipartFileEntry) (int64, error)

// Multipart does a multipart file upload request of type multipart/form-data
func (c *SimpleClient) Multipart(urlStr string, files map[string]string, get, post url.Values, writeFileCallback WriteFileFunc) (*http.Request, *http.Response, error) {
	urlData, err := url.Parse(c.Base + urlStr)
	if err != nil {
		return nil, nil, err
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
	urlStr = urlData.String()

	fileEntries := []MultipartFileEntry{}
	for fileParam, path := range files {
		fileHandle, err := os.Open(path)
		if err != nil {
			return nil, nil, err
		}
		fileEntries = append(fileEntries, MultipartFileEntry{
			Param: fileParam,
			Path:  path,
			File:  fileHandle,
		})
	}

	reqReader, reqWriter := io.Pipe()
	multipartWriter := multipart.NewWriter(reqWriter)

	go func() {
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
	}()

	req, err := c.NewRequest("POST", urlStr, reqReader)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Content-Type", multipartWriter.FormDataContentType())

	return c.send(req)
}