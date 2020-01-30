
run:
	@ go run ./src/.

build:
	go build -o ./bin/app ./src/.

fmt:
	go fmt ./src/...

test:
	@ go test ./src/...

dep:
	@ cd ./src
	go mod tidy

.PHONY: run build fmt dep
