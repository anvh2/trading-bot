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

docker:
	docker build --no-cache --progress=plain -t analyzer:latest -f ./internal/servers/analyzer/Dockerfile .
	docker build --no-cache --progress=plain -t crawler:latest -f ./internal/servers/crawler/Dockerfile .
	docker build --no-cache --progress=plain -t notifier:latest -f ./internal/servers/notifier/Dockerfile .
	docker build --no-cache --progress=plain -t commander:latest -f ./internal/servers/commander/Dockerfile .

deploy:
	docker-compose up --detach --build

mock:
	moq -pkg ntfmock -out ./pkg/api/v1/notifier/mock/service.mock.go ./pkg/api/v1/notifier NotifierServiceClient