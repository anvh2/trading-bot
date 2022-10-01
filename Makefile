HOME = $(shell pwd)
BIN = $(shell basename ${HOME})

clean:
	rm -f $(BIN)

gen: 
	buf generate -o ${HOME}/pkg

go-vendor:
	go mod tidy && go mod vendor

build: clean go-vendor
	GOOS=linux GOARCH=amd64 go build -o $(BIN)

rsync-crawler: build
	rsync -avz $(BIN) config.* root@159.223.44.181:/server/$(BIN)/crawler

rsync-analyzer: build
	rsync -avz $(BIN) config.* root@128.199.173.40:/server/$(BIN)/analyzer

rsync-notifier: build
	rsync -avz $(BIN) config.* root@159.223.67.54:/server/$(BIN)/notifier

rsync-scripts:
	rsync -avz runserver root@159.223.44.181:/server/$(BIN)/crawler
	rsync -avz runserver root@128.199.173.40:/server/$(BIN)/analyzer
	rsync -avz runserver root@159.223.67.54:/server/$(BIN)/notifier

deploy: build
	ssh root@165.22.103.161 sh /server/$(BIN)/runserver restart

local:
	rm -f tmp/server.log
	go run main.go start --config config.dev.toml