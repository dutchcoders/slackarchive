package slack

import (
	"net/url"
	"errors"
	"net/http"
	"bytes"
	"encoding/json"
)

const (
	slackOAuthAuthorizeUrl = "https://slack.com/oauth/authorize"
	slackOAuthRedeemUrl = "https://slack.com/api/oauth.access"
)

type OAuthClient struct {
	ClientId string
	ClientSecret string
	TeamId string
	RedirectUri string

	redeemUrl *url.URL
	authorizeUrl *url.URL
	httpClient *http.Client
}

type AccessToken struct {
	Token string `json:"access_token"`
	Scope string `json:"scope"`
}

func NewOAuthClient(clientID, clientSecret, redirectUri string) *OAuthClient {

	authorize, _ := url.Parse(slackOAuthAuthorizeUrl)
	redeem, _ := url.Parse(slackOAuthRedeemUrl)

	return &OAuthClient{
		ClientId: clientID,
		ClientSecret: clientSecret,
		RedirectUri: redirectUri,

		authorizeUrl: authorize,
		redeemUrl: redeem,

		httpClient: http.DefaultClient,
	}
}



func (cl *OAuthClient) LoginUrl(state string) *url.URL {
	u := *cl.authorizeUrl;
	uq := u.Query()

	uq.Add("redirect_uri", cl.RedirectUri)
	uq.Add("scope", "identify")
	uq.Add("client_id", cl.ClientId)
	uq.Add("state", state)

	if cl.TeamId != "" {
		uq.Add("team", cl.TeamId)
	}

	u.RawQuery = uq.Encode()

	return &u
}

func (cl *OAuthClient) RedeemCode(code string) (*AccessToken, error) {

	if code == "" {
		return nil, errors.New("Missing code")
	}


	uq := url.Values{}

	uq.Add("redirect_uri", cl.RedirectUri)
	uq.Add("scope", "identify")
	uq.Add("client_id", cl.ClientId)
	uq.Add("client_secret", cl.ClientSecret)
	uq.Add("code", code)

	req, err := http.NewRequest(_POST, cl.redeemUrl.String(), bytes.NewBufferString(uq.Encode()));

	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := cl.httpClient.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	authToken := new(AccessToken)
	err = json.NewDecoder(resp.Body).Decode(authToken)

	if err != nil {
		return nil, err
	}

	if authToken.Token == "" {
		return nil, errors.New("Invalid auth token")
	}

	return authToken, nil
}
