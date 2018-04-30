package slack

import (
	"net/http"
	"net/url"
	"bytes"
	"encoding/json"
	"log"
)

const (
	slackBaseUrl = "https://slack.com/api/"
	_GET = "GET"
	_POST = "POST"
)

type Response struct {
	*http.Response
}

type SlackClient struct {
	httpClient *http.Client

	BaseUrl *url.URL

	Token string
	Groups *GroupService
	Auth *AuthService
}

func NewClient(token string) *SlackClient {

	baseURL, _ := url.Parse(slackBaseUrl)

	api := &SlackClient{
		httpClient: http.DefaultClient,
		BaseUrl: baseURL,
		Token: token,
	}
	api.Groups = &GroupService{ api: api }
	api.Auth = &AuthService{ api: api }

	return api;
}

func (s *SlackClient) NewRequest(method, path string, body interface {}) (*http.Request, error) {
	rel, err := url.Parse(path)
	if err != nil {
		return nil, err
	}

	params := rel.Query()

	if params.Get("token") == "" {
		params.Set("token", s.Token)
	}


	u := s.BaseUrl.ResolveReference(rel)
	u.RawQuery = params.Encode()

	buf := new(bytes.Buffer)
	if body != nil {
		err := json.NewEncoder(buf).Encode(body)
		if err != nil {
			return nil, err
		}
	}

	log.Printf("Making request to %s", u.String())

	req, err := http.NewRequest(method, u.String(), buf)
	if err != nil {
		return nil, err
	}


	return req, nil
}

func (s *SlackClient) Do(req *http.Request, v interface {}) (*Response, error) {
	resp, err := s.httpClient.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	response := &Response{ Response: resp }

	err = nil
	if v != nil {
		err = json.NewDecoder(resp.Body).Decode(v)
	}

	return response, err
}
