package github

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
)

// Client represents a GitHub client.
type Client interface {
	GetUser() (*User, error)
	GetRepo(string) (*Repo, error)
	ListIssues(string, *ListIssuesParams) Issues
	ListComments(string, int) Comments
	ListPullReqs(string, *ListPullReqsParams) PullReqs
	Import(string, *Import) error
}

// New creates a new GitHub client.
func New(token, endpoint string) Client {
	cli := &http.Client{Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: endpoint != "https://api.github.com",
		},
	}}
	return &client{token: token, endpoint: endpoint, client: cli}
}

type client struct {
	token, endpoint string
	client          *http.Client
}

func (c *client) url(path string) string {
	return c.endpoint + path
}

func (c *client) get(path string) (*http.Response, error) {
	req, err := c.request("GET", path, nil)
	if err != nil {
		return nil, err
	}
	fmt.Printf("fetching: %s\n", req.URL)
	return c.client.Do(req)
}

func (c *client) post(path string, body io.Reader) (*http.Response, error) {
	req, err := c.request("POST", path, body)
	if err != nil {
		return nil, err
	}
	fmt.Printf("posting: %s\n", req.URL)
	return c.client.Do(req)
}

func (c *client) request(method, path string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, path, body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "token "+c.token)
	req.Header.Add("Accept", "application/vnd.github.golden-comet-preview+json")
	req.Header.Add("User-Agent", "github-migrator")
	return req, nil
}
