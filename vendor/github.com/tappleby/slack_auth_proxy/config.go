package main

import (
	yaml "gopkg.in/yaml.v1"
	"io/ioutil"
	"fmt"
)

const (
	defaultServerAddr = "127.0.0.1:4180"
)



type Configuration struct {
	// Server settings
	ServerAddr		string 						`yaml:"server_addr,omitempty"`
	Upstreams		[]*UpstreamConfiguration 	`yaml:"upstreams,omitempty"`
	RedirectUri		string						`yaml:"redirect_uri,omitempty"`
	PassBasicAuth	bool 					 	`yaml:"pass_basic_auth,omitempty"`

	// Cookie settings
	CookieDomain	string						`yaml:"cookie_domain,omitempty"`
	CookieHashKey	string						`yaml:"cookie_hash_key,omitempty"`
	CookieBlockKey	string						`yaml:"cookie_block_key,omitempty"`

	// Slack Settings
	ClientId 		string 						`yaml:"client_id"`
	ClientSecret 	string 						`yaml:"client_secret"`
	SlackTeam 		string 					 	`yaml:"slack_team,omitempty"`
	AuthToken		string 					 	`yaml:"auth_token,omitempty"`

	//Other settings
	HtPasswdFile	string						`yaml:"htpasswd_file,omitempty"`
	Debug			bool 					 	`yaml:"debug,omitempty"`

}

func LoadConfiguration(configFile string) (config *Configuration, err error) {

	configBuf, err := ioutil.ReadFile(configFile)

	if err != nil {
		err = fmt.Errorf("Failed to read configuration %s: %v", configFile, err)
		return
	}

	config = &Configuration{
		ServerAddr: defaultServerAddr,
		PassBasicAuth: true,
	}


	if err = yaml.Unmarshal(configBuf, &config); err != nil {
		return
	}


	if config.ClientId == "" {
		err = fmt.Errorf("Client id must be set in configuration")
		return
	}

	if config.ClientSecret == "" {
		err = fmt.Errorf("Client secret must be set in configuration")
		return
	}

	if config.SlackTeam == "" {
		err = fmt.Errorf("slack_team must be set in the configuration")
		return
	}

	if config.CookieHashKey == "" || config.CookieBlockKey == "" {
		err = fmt.Errorf("cookie_hash_key and cookie_block_key must be set. please use ./slack_auth_proxy --keys to generate.")
		return
	}

	if config.RedirectUri == "" {
		config.RedirectUri = fmt.Sprintf("http://%s%s", config.ServerAddr, oauthCallbackPath)
	}

	for _, upstream := range config.Upstreams {
		if err = upstream.Parse(); err != nil {
			err = fmt.Errorf("Error parsing upstream %s: %v", upstream.Host, err)
			return
		}
	}


	return
}
