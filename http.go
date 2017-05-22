package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

var client = &http.Client{}

func NewRequest(method, urlStr string, body io.Reader) (*http.Request, error) {
	token, err := GetAuthToken()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, urlStr, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authentication", token)
	req.Header.Set("User-Agent", "Jolt/1.0.0")
	return req, nil
}

func NewGet(urlStr string, params url.Values) (*http.Request, error) {
	urlData, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
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

	return NewRequest("GET", urlStr, nil)
}

func NewPost(urlStr string, get url.Values, post interface{}) (*http.Request, error) {
	urlData, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
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
		return nil, err
	}

	req, err := NewRequest("POST", urlStr, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

type multipartFileEntry struct {
	param string
	path  string
	file  *os.File
}

func NewMultipart(urlStr string, files map[string]string, get, post url.Values) (*http.Request, error) {
	urlData, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
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

	fileEntries := []multipartFileEntry{}
	for fileParam, path := range files {
		fileHandle, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		fileEntries = append(fileEntries, multipartFileEntry{
			param: fileParam,
			path:  path,
			file:  fileHandle,
		})
	}

	reqReader, reqWriter := io.Pipe()
	multipartWriter := multipart.NewWriter(reqWriter)

	go func() {
		for _, fileEntry := range fileEntries {
			defer fileEntry.file.Close()
		}
		defer reqWriter.Close()
		defer multipartWriter.Close()

		for _, fileEntry := range fileEntries {
			log.Println("1")
			log.Println(filepath.Base(fileEntry.path))
			fileField, err := multipartWriter.CreateFormFile(fileEntry.param, filepath.Base(fileEntry.path))
			if err != nil {
				log.Fatal(err)
				return
			}
			log.Println("wat")

			log.Println("2")
			_, err = io.Copy(fileField, fileEntry.file)
			if err != nil {
				return
			}
		}

		log.Println("3")
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

	log.Println("4")
	req, err := NewRequest("POST", urlStr, reqReader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", multipartWriter.FormDataContentType())
	return req, nil
}
