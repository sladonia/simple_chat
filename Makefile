
run:
	@ go run ./src/.

build:
	go build -o ./bin/app ./src/.

fmt:
	go fmt ./src/...

.PHONY: run build fmt
