BIN = subscribe

clean:
	rm -f $(BIN)

go-vendor:
	go mod tidy && go mod vendor

build: clean go-vendor
	GOOS=linux GOARCH=amd64 go build -o $(BIN)

rsync: build
	rsync -avz $(BIN) config.* root@128.199.173.40:/server/trading-bot/analyzer
