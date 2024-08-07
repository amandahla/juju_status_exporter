MAIN_PACKAGE_PATH := .
BINARY_NAME := juju_status_exporter

## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

## fmt: format code
.PHONY: fmt
fmt:
	go fmt ./...

## tidy: format code and tidy modfile
.PHONY: tidy
tidy: fmt
	go mod tidy -v

## audit: run quality control checks
.PHONY: audit
audit:
	go mod verify
	go vet ./...
	go run honnef.co/go/tools/cmd/staticcheck@latest -checks=all,-ST1000,-U1000 ./...
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...
	go test -race -buildvcs -vet=off ./...

## test: run all tests
.PHONY: test
test:
	go test -v -race -buildvcs ./...

## build: build the application
.PHONY: build
build:
	# Include additional build steps, like TypeScript, SCSS or Tailwind compilation here...
	GOOS=linux GOARCH=amd64 go build -ldflags='-s' -o=${BINARY_NAME} ${MAIN_PACKAGE_PATH}
