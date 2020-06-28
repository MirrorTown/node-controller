package common

import (
	"bytes"
	"errors"
	"log"
	"net/http"
	"time"
)

var ErrHttpRequest = errors.New("create HTTP request failed")

const (
	ADD           = "ADD"
	UPDATE        = "UPDATE"
	DELETE        = "DELETE"
	WatchFromKind = "watch"
)

type Ready2Send struct {
	Title  string
	Start  string
	User   string
	Alerts string
}

func HttpPost(url string, params map[string]string, headers map[string]string, body []byte) (*http.Response, error) {
	//new request
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, ErrHttpRequest
	}
	//add params
	q := req.URL.Query()
	if params != nil {
		for key, val := range params {
			q.Add(key, val)
		}
		req.URL.RawQuery = q.Encode()
	}
	//add headers
	if headers != nil {
		for key, val := range headers {
			req.Header.Add(key, val)
		}
	}
	//http client
	client := &http.Client{Timeout: 5 * time.Second} //Add the timeout,the reason is that the default client has no timeout set; if the remote server is unresponsive, you're going to have a bad day.
	resp, err := client.Do(req)

	defer resp.Body.Close()
	return resp, err
}

func HttpGet(url string, params map[string]string, headers map[string]string) (*http.Response, error) {
	//new request
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Println(err)
		return nil, ErrHttpRequest
	}
	//add params
	q := req.URL.Query()
	if params != nil {
		for key, val := range params {
			q.Add(key, val)
		}
		req.URL.RawQuery = q.Encode()
	}
	//add headers
	if headers != nil {
		for key, val := range headers {
			req.Header.Add(key, val)
		}
	}
	//http client
	client := &http.Client{Timeout: 5 * time.Second} //Add the timeout,the reason is that the default client has no timeout set; if the remote server is unresponsive, you're going to have a bad day.
	resp, err := client.Do(req)

	defer resp.Body.Close()
	return resp, err
}
