# Trading Bot

#### _Trading bot is written in Go._

## Installation

### Setup ENV
> Note: go to api key management from Binance and create key only read for this bussiness.
```sh
touch .env
cat > LIVE_API_KEY=${api_key} \n LIVE_SECRET_KEY=${secret_key}
```

### Install softwares
- Docker
  > Recommend docker desktop for visuzelizes all containers.

### Run signaler
```sh
git clone github.com/anvh2/trading-bot
cd trading-bot && make deploy-signaler
```

## Development
> Note: that this bot is under development.
