package api

import "net/http"

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string, httpClient *http.Client) Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return Client{baseURL: baseURL, httpClient: httpClient}
}

func (c Client) BaseURL() string {
	return c.baseURL
}
