HOME = $(shell pwd)
BIN = $(shell basename ${HOME})

clean:
	rm -f $(BIN)

gen: 
	buf generate --path ./api -o ${HOME}/pkg

go-vendor:
	go mod tidy && go mod vendor

build: clean go-vendor
	GOOS=linux GOARCH=amd64 go build -o $(BIN)

docker-build:
	docker build --no-cache --progress=plain -t analyzer:latest -f ./internal/servers/analyzer/Dockerfile .
	docker build --no-cache --progress=plain -t crawler:latest -f ./internal/servers/crawler/Dockerfile .
	docker build --no-cache --progress=plain -t notifier:latest -f ./internal/servers/notifier/Dockerfile .
	docker build --no-cache --progress=plain -t commander:latest -f ./internal/servers/commander/Dockerfile .
	docker build --no-cache --progress=plain -t appraiser:latest -f ./internal/servers/appraiser/Dockerfile .

deploy:
	docker-compose up --detach --build

deploy-redis:
	docker-compose up -d --no-deps --build redis-server

deploy-analyzer:
	docker-compose up -d --no-deps --build analyzer

deploy-crawler:
	docker-compose up -d --no-deps --build crawler

deploy-notifier:
	docker-compose up -d --no-deps --build notifier

deploy-commander:
	docker-compose up -d --no-deps --build commander

deploy-appraiser:
	docker-compose up -d --no-deps --build appraiser

deploy-signaler:
	docker-compose up -d --no-deps --build redis-server
	docker-compose up -d --no-deps --build analyzer
	docker-compose up -d --no-deps --build crawler
	docker-compose up -d --no-deps --build notifier

mock:
	moq -pkg ntfmock -out ./pkg/api/v1/notifier/mock/service.mock.go ./pkg/api/v1/notifier NotifierServiceClient