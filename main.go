package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
)

var inReader = bufio.NewReader(os.Stdin)
var token = ""

func main() {
	token = "test"
	req, err := NewMultipart("http://postman-echo.com/post?param1=eyy",
		map[string]string{
			"file":  "test.txt",
			"file2": "test2.txt",
		},
		url.Values(map[string][]string{
			"param2": []string{"eyy2"},
			"array1": []string{"element1", "element2"},
		}),
		url.Values(map[string][]string{
			"post1":  []string{"eyy1"},
			"post2":  []string{"eyy2"},
			"array1": []string{"element1", "element2"},
		}))

	if err != nil {
		log.Fatal(err)
	}
	reqBytes, err := httputil.DumpRequest(req, true)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(string(reqBytes))

	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	resBytes, err := httputil.DumpResponse(res, true)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(string(resBytes))
}

func GetAuthToken() (string, error) {
	if token == "" {
		return "", errors.New("Not authenticated")
	}
	return token, nil
}

func promptAuthToken() (string, error) {
	fmt.Print("Enter your token: ")
	line, err := inReader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(line), nil
}
