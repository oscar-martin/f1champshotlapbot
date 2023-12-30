# F1Champs Telegram Bot

Telegram Bot for accessing the hotlaps and live sessions from f1champs.es league.

## Features

- Multiple servers
- See servers status
- See current session data/standings
- Pushes notifications when a new session starts with at least one driver
- LiveMap
- Generate the track map for the current session
- Fetch the car image for drivers in current session

## Usage

Follow instructions in [Telegram Bot Father](https://core.telegram.org/bots#6-botfather) to create a new bot.

When creating the bot, it is recommended to add the next commands
[(via /setcommands)](https://core.telegram.org/bots/features#edit-bots) to the bot:

```
start - Give a welcome message
menu - Show the bot menu
```

Go to the [releases](https://github.com/oscar-martin/f1champshotlapbot/releases) and download the binary for your platform.

Certain environment variable must be set:

- `TELEGRAM_TOKEN`: the token provided by Telegram Bot Father for your bot.
- `API_DOMAIN`: it is the domain where the F1Champs API is listening on. For example: `https://f1champs-domain.es`
- `LIVEMAP_DOMAIN`: it is the domain where the livemap will be exposed publicly. For example: `https://<my-public-domain>`
- `WEBSERVER_ADDRESS` it is the address where the bot will be listening to server livemap data. For example:
  `http://<my-lan-ip>:8080`. Default value is `0.0.0.0:8080`.
- `RF2_SERVERS`: it is following the next format `<server_id>,<server_url>;<server_id>,<server_url>;...`.
    For example: `PrimaryServer,http://my-server-1:5397;TrainingServer1,http://my-server-2:5397`

### Example

#### Linux

```bash
export TELEGRAM_TOKEN=<your token>
export RF2_SERVERS=<your rf2 servers>
export API_DOMAIN=<your api domain>
export LIVEMAP_DOMAIN=<your livemap domain>
export WEBSERVER_ADDRESS=<your webserver address>
./rfactor2telegrambot
```

#### Windows

```
set TELEGRAM_TOKEN=<your token>
set RF2_SERVERS=<your rf2 servers>
set API_DOMAIN=<your api domain>
set LIVEMAP_DOMAIN=<your livemap domain>
set WEBSERVER_ADDRESS=<your webserver address>
rfactor2telegrambot.exe
```

With powershell:

```
$env:TELEGRAM_TOKEN = '<your token>'
$env:RF2_SERVERS = '<your rf2 servers>'
$env:API_DOMAIN = '<your api domain>'
$env:LIVEMAP_DOMAIN = '<your livemap domain>'
rfactor2telegrambot.exe
```

### Network configuration

- The bot must have access to the internet to be able to connect to Telegram servers.
- The bot is recommended to run in the same LAN where the rFactor2 servers are running, although it is not mandatory. If
  the bot is running in a different LAN, the rFactor2 servers (at least, port 5397) must be exposed publicly.
- The bot must be able to connect to the rFactor2 servers (port 5397) to be able to get the data.
- The bot exposes a webserver (port 8080 by default) to serve livemap data. This port must be exposed publicly for the
  livemap feature to work.
- The bot will send the livemap data as a link to the `LIVEMAP_DOMAIN`. You are responsible to configure the domain to
  point to the bot webserver at the port configured with `WEBSERVER_ADDRESS`.

For testing locally, you can use LAN IP address for `LIVEMAP_DOMAIN`, example:

```bash
export LIVEMAP_DOMAIN=http://192.168.1.12:8080
export WEBSERVER_ADDRESS=:8080
```

With the previous configuration, the bot will send the livemap data as a link to `http://192.168.1.12:8080` and your
Telegram client will be able to access it if you are in the same LAN.

## Miscellaneous

- The bot will create a file called `livetiming-bot.db` that will contain the ID of users that have subscribed to
  notifications. This file is created in the same directory where the bot is running. This file should not be deleted
  unless you want to lose the subscriptions.
- The bot will create a folder called `resources` to cache the files for the cars and trackmaps that are
  downloaded/generated from the rFactor2 servers. The content of this folder can be deleted at any time.

## Development notes for MacOS

To build it for linux:

1. Install musl-cross-make and then run:

  ```bash
  brew install FiloSottile/musl-cross/musl-cross
  ```

2. And then:

  ```bash
  CC=x86_64-linux-musl-gcc CXX=x86_64-linux-musl-g++ GOARCH=amd64 GOOS=linux CGO_ENABLED=1 go build -ldflags "-linkmode external -extldflags -static" -o f1champshotlapbot-linux .
  ```
