# SlackArchive [![Gitter](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/dutchcoders/slackarchive?utm_source=badge&utm_medium=badge&utm_campaign=&utm_campaign=pr-badge&utm_content=badge) [![Go Report Card](https://goreportcard.com/badge/dutchcoders/slackarchive)](https://goreportcard.com/report/dutchcoders/slackarchive) [![Build Status](https://travis-ci.org/dutchcoders/slackarchive.svg?branch=master)](https://travis-ci.org/dutchcoders/slackarchive) [![codecov](https://codecov.io/gh/dutchcoders/slackarchive/branch/master/graph/badge.svg)](https://codecov.io/gh/dutchcoders/slackarchive) [![Docker pulls](https://img.shields.io/docker/pulls/dutchcoders/slackarchive.svg)](https://hub.docker.com/r/dutchcoders/slackarchive/)

SlackArchive can be started with just a few commands. First make sure you create copies of the config.yaml.sample to their according config.yaml and change the files with the correct parameters. SlackArchive supports Let's Encrypt for https.

## Docker 

Using SlackArchive with Docker is easy, just run the following commands. All components and dependencies will be start correctly.

```
# clone repository
git clone https://github.com/dutchcoders/slackarchive-docker

# copy sample configs
cp slackarchive/config.yaml.sample slackarchive/config.yaml
cp slackarchive-bot/config.yaml.sample slackarchive-bot/config.yaml

# update docker-compose and create mongo passwords and ports
# update slackarchive/config.yaml with correct tokens
# update slackarchive-bot/config.yaml with correct tokens

# create network
docker network create slackarchive

# create elasticsearch and mongodb and wait to be started
docker-compose run --rm wait_for_dependencies

# initialize elasticsearch and mongodb 
docker-compose run --rm slackarchive-init

# start slackarchive
docker-compose up slackarchive

# start slackarchive-bot
docker-compose up slackarchive-bot
```

Now SlackArchive has been started and you can access it at http://127.0.0.1:8080/.

## Components

SlackArchive consists of fhe following components:

* SlackArchive (https://github.com/dutchcoders/slackarchive)
* SlackArchive App (https://github.com/dutchcoders/slackarchive-app)
* SlackArchive ArchiveBot (https://github.com/dutchcoders/slackarchive-bot)
* SlackArchive Importer (https://github.com/dutchcoders/slackarchive-import)

* Docker environment (https://github.com/dutchcoders/slackarchive-docker)
* Docker Init (https://github.com/dutchcoders/slackarchive-init)

## Creators

Remco Verhoef (@remco_verhoef) and Kaspars Sprogis.

## Copyright and license

Code and documentation copyright 2018 DutchCoders.

Code released under [Affero General Public License](LICENSE).
