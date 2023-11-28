# F1Champs Telegram Bot

Telegram Bot for accessing the hotlaps from f1champs.es league.

Once it is build, you can run it with:

```bash
export TELEGRAM_TOKEN=<your token>
export API_DOMAIN=<your domain>
export RF2_SERVERS=<your rf2 servers>
./f1champshotlapbot
```

`API_DOMAIN` is the domain where the API is running, e.g. `https://my-server.com`
`RF2_SERVERS` is following the next format `<server_id>,<server_url>;<server_id>,<server_url>;...`.
For example: `PrimaryServer,http://my-server.com:8080;TrainingServer1,http://my-server.com:8081`

To build it for linux:

1. Install musl-cross-make and then run:

  ```bash
  brew install FiloSottile/musl-cross/musl-cross
  ```

2. And then:

  ```bash
  CC=x86_64-linux-musl-gcc CXX=x86_64-linux-musl-g++ GOARCH=amd64 GOOS=linux CGO_ENABLED=1 go build -ldflags "-linkmode external -extldflags -static" -o f1champshotlapbot-linux .
  ```

## TODO

- Export session data once the session is over and create a naming convention for the folders/files.
- Add bot app to browser over the historical data (previous point).
