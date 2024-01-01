## tidy: format code and tidy modfile
.PHONY: tidy
tidy:
	go fmt ./...
	go mod tidy -v

## build: build the application
.PHONY: build
build:
	go mod verify
	go build -ldflags='-s' -o=./bin/create-db create-db.go

## run: run the application
.PHONY: run
run: tidy build
	./bin/create-db
