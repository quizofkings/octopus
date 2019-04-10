# versioning
CURRENT_BRANCH = $(shell git rev-parse --abbrev-ref HEAD)
COMMIT_SHORT_HASH = $(shell git rev-parse --short HEAD)
COMMIT_HASH = $(shell git rev-parse HEAD)
DATE = $(shell date -u +%Y.%m.%d-%H%M%S)
VERSION = $(COMMIT_SHORT_HASH)

# flags
LDFLAGS = -ldflags '-extldflags "-static" -X app.Version=$(VERSION) -X app.CommitHash=$(COMMIT_HASH)'
BUILDFLAGS = -a -installsuffix cgo

build:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ./octopus .
	docker build --no-cache -f Dockerfile -t amirsoleimani/octopus:$(VERSION) .