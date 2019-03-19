# goch

goch is a self-hosted live-chat server written in Go.

It allows you to run a live-chat software on your own infrastructure.

You can create multiple private and public chatrooms where two or more users can be at the same time.

For communication it uses RESTful endpoints, Websockets, NATS Streaming and Redis.

Goch is a fork of [Gossip](https://github.com/aneshas/gossip), with many added features and fixes.

## Getting started

To run goch locally, you need `docker`, `docker-compose` and `go` installed and set on your path. After downloading/cloning the project, run `./up` which compiles the binary and runs docker-compose with goch, NATS Streaming and Redis. If there were no errors, goch should be running on localhost (port 80).

## License

goch is licensed under the MIT license. Check the [LICENSE](LICENSE) file for details.

## Author

[Emir Ribic](https://ribice.ba)