HOME = $(shell pwd)
BIN = $(shell basename ${HOME})

clean:
	rm -f $(BIN)

go-vendor:
	go mod tidy && go mod vendor

build: clean go-vendor
	GOOS=linux GOARCH=amd64 go build -o $(BIN)

rsync:
	rsync -avz $(BIN) config.* root@165.22.103.161:/server/$(BIN)/

deploy: build rsync
	ssh root@165.22.103.161 sh /server/$(BIN)/runserver restart

local:
	rm -f tmp/server.log
	go run main.go start --config config.dev.toml