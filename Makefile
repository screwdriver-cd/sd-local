GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOLIST=$(GOCMD) list
GOTOOL=$(GOCMD) tool
GOLIST_PKG=$(GOLIST) ./... | grep -v /vendor/
GOLINT=$(GOPATH)/bin/golint
GORELEASER=goreleaser
BINARY_NAME=sd-local
COVERPROFILE?=cover.out

all: test build
test: format vet lint clean_mod_file
	$(GOTEST) -race -cover -coverprofile=$(COVERPROFILE) -covermode=atomic ./...
vet:
	$(GOCMD) vet -v ./...
lint:
	$(GOLINT) $$($(GOLIST_PKG))
format:
	find . -name '*.go' | xargs gofmt -s -w
clean_mod_file:
	$(GOCMD) mod tidy
mod_download:
	$(GOCMD) mod download
build: mod_download
	$(GOBUILD) -o $(BINARY_NAME) -v
publish_dry_run:
	$(GORELEASER) --snapshot --skip-publish --rm-dist
publish:
	$(GORELEASER) --rm-dist
run: build
	./$(BINARY_NAME)
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -rf ./dist
cover_html:
	$(GOTEST) -race -cover -coverprofile=cover.out -covermode=atomic ./...
	$(GOTOOL) cover -html=cover.out -o cover.html
	open cover.html
