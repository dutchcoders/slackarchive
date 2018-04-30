package main

import (
	"math/rand"
	"os"
	_ "os/exec"
	"time"

	cli "gopkg.in/urfave/cli.v1"

	_ "github.com/go-sql-driver/mysql"

	slackarchiveapi "github.com/dutchcoders/slackarchive/api"
	config "github.com/dutchcoders/slackarchive/config"
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

var version string = "0.1"

func main() {
	app := cli.NewApp()
	app.Name = "SlackArchive"
	app.Version = version
	app.Flags = append(app.Flags, []cli.Flag{
		cli.StringFlag{
			Name:   "config, c",
			Value:  "config.yaml",
			Usage:  "Custom configuration file path",
			EnvVar: "",
		},
	}...)

	app.Action = run

	app.Run(os.Args)
}

func run(c *cli.Context) {
	conf := config.MustLoad(c.GlobalString("config"))

	api := slackarchiveapi.New(conf)
	api.Serve()
}
