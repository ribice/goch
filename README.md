# goch
[![Build Status](https://travis-ci.org/ribice/goch.svg?branch=master)](https://travis-ci.org/ribice/goch)
[![codecov](https://codecov.io/gh/ribice/goch/branch/master/graph/badge.svg)](https://codecov.io/gh/ribice/goch)
[![Go Report Card](https://goreportcard.com/badge/github.com/ribice/goch)](https://goreportcard.com/report/github.com/ribice/goch)
[![Maintainability](https://api.codeclimate.com/v1/badges/c3cb09dbc0bc43186464/maintainability)](https://codeclimate.com/github/ribice/goch/maintainability)

goch is a self-hosted live-chat server written in Go.

It allows you to run a live-chat software on your own infrastructure.

You can create multiple private and public chatrooms where two or more users can be at the same time.

For communication, it uses RESTful endpoints, Websockets, NATS Streaming, and Redis.

Goch is a fork of [Gossip](https://github.com/aneshas/gossip), with many added features and fixes.

## Getting started

To run goch locally, you need `docker`, `docker-compose` and `go` installed and set on your path. After downloading/cloning the project, run `./up` which compiles the binary and runs docker-compose with goch, NATS Streaming, and Redis. If there were no errors, goch should be running on localhost (port 8080).

## How it works

In order for the server to run, `ADMIN_USERNAME` and `ADMIN_PASSWORD` env variables have to be set. In the repository, they are set to `admin` and `pass` respectively, but you should obviously change those for security reasons.

Once the server is running, the following routes are available:

`POST /admin/channels`: Creates a new channel. You have to provide a unique name for a channel (usually an ID), and the response includes channel's secret which will be used for connecting to channel later on. This endpoint should be invoked server-side with provided admin credentials. The response should be saved in order to connect to the channel later on.

`POST /register`: Register a user in a channel. In order to register for the channel, a UID, DisplayName, ChannelSecret, and ChannelName needs to be provided. Optionally user secret needs to be provided, but if not the server will generate and return one.

`GET /connect`: Connects to a chat and returns a WebSocket connection, along with chat history. Channel, UID, and Secret need to be provided. Optionally LastSeq is provided which will return chat history only after LastSeq (UNIX timestamp).

The remaining routes are only used as 'helpers':

`GET /channels/{name}?secret=$SECRET`: Returns list of members in a channel. Channel name has to be provided as URL param and channel secret as a query param.

`GET /admin/channels`: Returns list of all available channels.

`GET /admin/channels/{name}/user/{uid}`: Returns list of unread messages on a chat for a user.

## License

goch is licensed under the MIT license. Check the [LICENSE](LICENSE) file for details.

## Author

[Emir Ribic](https://ribice.ba)