package slack_test

import (
	"github.com/tappleby/slack_auth_proxy/slack"
	"github.com/stretchr/testify/assert"
	"testing"
)


func Test_NewOAuthClient(t *testing.T) {
	client := slack.NewOAuthClient("foo", "bar", "baz")

	assert.Equal(t, client.ClientId, "foo")
	assert.Equal(t, client.ClientSecret, "bar")
	assert.Equal(t, client.RedirectUri, "baz")
}

func Test_OAuthClient_LoginUrl(t *testing.T) {
	client := slack.NewOAuthClient("foo", "bar", "baz")

	u := client.LoginUrl("1234")
	uv := u.Query()

	assert.Contains(t, u.String(), "https://slack.com/oauth/authorize")
	assert.Equal(t, uv.Get("redirect_uri"), "baz")
	assert.Equal(t, uv.Get("scope"), "identify")
	assert.Equal(t, uv.Get("client_id"), "foo")
	assert.Equal(t, uv.Get("state"), "1234")

	assert.Empty(t, uv.Get("team"));


	client.TeamId = "foobar"
	u = client.LoginUrl("1234")
	uv = u.Query()

	assert.Equal(t, uv.Get("team"), "foobar")
}
