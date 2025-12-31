.PHONY: run build test test-race lint fmt vet ci clean

# Binary name
BINARY=jenkins-tui

# Build flags
LDFLAGS=-ldflags "-s -w"

run:
	go run ./cmd/jenkins-tui

build:
	go build $(LDFLAGS) -o $(BINARY) ./cmd/jenkins-tui

test:
	go test ./...

test-race:
	go test -race ./...

lint:
	golangci-lint run

fmt:
	gofmt -w .
	goimports -w .

vet:
	go vet ./...

ci: fmt vet test

clean:
	rm -f $(BINARY)
	go clean

# Development helpers
deps:
	go mod tidy
	go mod download

# Build for multiple platforms
build-all:
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY)-linux-amd64 ./cmd/jenkins-tui
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY)-darwin-amd64 ./cmd/jenkins-tui
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BINARY)-darwin-arm64 ./cmd/jenkins-tui
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY)-windows-amd64.exe ./cmd/jenkins-tui
