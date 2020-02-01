BUILD_ENV_VARS=GOOS=linux GOARCH=amd64 CGO_ENABLED=0
SERVICE_NAME=simple_chat

run:
	@ go run ./src/.

build:
	$(BUILD_ENV_VARS) go build -o ./bin/app ./src/.

docker-build: build
	docker build -t $(SERVICE_NAME) .

fmt:
	go fmt ./src/...

test:
	@ go test -cover ./src/...

dep:
	@ cd ./src
	go mod tidy

.PHONY: run build docker-build test fmt dep
