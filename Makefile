.PHONY: run build install clean test test-full lint tidy

BINARY  := music-for-coding-tui
CMD     := ./cmd/mfp
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-X main.version=$(VERSION)"

run:
	go run $(LDFLAGS) $(CMD)

build:
	go build $(LDFLAGS) -o $(BINARY) $(CMD)

install:
	go install $(LDFLAGS) $(CMD)
	@echo "✓ mfp $(VERSION) installed to $$(go env GOPATH)/bin/mfp"
	@echo "  Run: mfp"

clean:
	rm -f $(BINARY)

test:
	go test ./... -short -timeout 90s

test-full:
	go test ./... -timeout 90s

lint:
	go vet ./...

tidy:
	go mod tidy
