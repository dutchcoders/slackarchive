package config

import (
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"

	"github.com/tappleby/slack_auth_proxy/slack"
)

type Config struct {
	Listen    string `yaml:"listen"`
	ListenTLS string `yaml:"listen_tls"`

	Bot struct {
		Token string `yaml:"token"`
	} `yaml:"bot"`

	Team string `yaml:"team"`

	Database struct {
		DSN string `yaml:"dsn"`
	} `yaml:"database"`

	Slack struct {
		ClientId     string `yaml:"client_id"`
		ClientSecret string `yaml:"client_secret"`
	} `yaml:"slack"`

	Cookies struct {
		AuthenticationKey string `yaml:"authentication_key"`
		EncryptionKey     string `yaml:"encryption_key"`
	} `yaml:"cookies"`

	Data string `yaml:"data"`

	ElasticSearch struct {
		URL string `yaml:"url"`
	} `yaml:"elasticsearch"`

	SessionName string `yaml:"session_name"`
}

func MustLoad(path string) *Config {
	var c Config
	if err := c.Load(path); err != nil {
		panic(err)
	}
	return &c
}

func (c *Config) NewSlackOAuthClient(redirectUri string) *slack.OAuthClient {
	client := slack.NewOAuthClient(c.Slack.ClientId, c.Slack.ClientSecret, redirectUri)
	return client
}

func (c *Config) Load(path string) error {
	var err error
	var b []byte
	if b, err = ioutil.ReadFile(path); err != nil {
		return err
	}

	if err = yaml.Unmarshal(b, &c); err != nil {
		return err
	}

	if c.Data == "" {
		c.Data = "."
	}

	if c.Listen == "" {
		c.Listen = "127.0.0.1:8080"
	}

	err = c.init()
	return err
}

// initialize connections and auth
func (c *Config) init() error {
	return nil
}
