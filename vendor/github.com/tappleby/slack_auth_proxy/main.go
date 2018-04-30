package main

import (
	"log"
	"net"
	"net/http"
	"strings"
	"flag"
	"path/filepath"
	"fmt"
	"github.com/tappleby/slack_auth_proxy/slack"
	"github.com/gorilla/securecookie"
	"encoding/base64"
)

const VERSION = "0.0.1"

var (
	defaultConfigFile, _ 	= filepath.Abs("./config.yml")
	configFile           	= flag.String("config", defaultConfigFile, "path to config file.")

	showVersion 			= flag.Bool("version", false, "print version string.")
	showKeys 				= flag.Bool("keys", false, "prints encryption keys for secure cookie.")
)

func main() {

	flag.Parse()

	if *showVersion {
		fmt.Printf("slack_auth_proxy v%s\n", VERSION)
		return
	}

	if *showKeys {
		enc := base64.StdEncoding
		hashKey := securecookie.GenerateRandomKey(64)
		blockKey := securecookie.GenerateRandomKey(32)

		fmt.Printf("cookie_hash_key: %s\n", enc.EncodeToString(hashKey))
		fmt.Printf("cookie_block_key: %s\n", enc.EncodeToString(blockKey))
		return
	}

	// Load config
	config, err := LoadConfiguration(*configFile)

	if err != nil {
		log.Fatal("Error loading config: ", err)
	}

	log.Println(config.Upstreams[0].HostURL)

	listener, err := net.Listen("tcp", config.ServerAddr)
	if err != nil {
		log.Fatalf("FATAL: listen (%s) failed - %s", config.ServerAddr, err.Error())
	}
	log.Printf("listening on %s", config.ServerAddr)

	oauthClient := slack.NewOAuthClient(config.ClientId, config.ClientSecret, config.RedirectUri)
	oauthClient.TeamId = config.SlackTeam

	oauthServer := NewOauthServer(oauthClient, config)

	if config.HtPasswdFile != "" {
		oauthServer.HtpasswdFile, err = NewHtpasswdFromFile(config.HtPasswdFile)
		if err != nil {
			log.Fatalf("FATAL: unable to open %s %s", config.HtPasswdFile, err.Error())
		}
	}

	server := &http.Server{Handler: oauthServer}
	err = server.Serve(listener)
	if err != nil && !strings.Contains(err.Error(), "use of closed network connection") {
	log.Printf("ERROR: http.Serve() - %s", err.Error())
	}

	log.Printf("HTTP: closing %s", listener.Addr().String())
}
