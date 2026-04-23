.PHONY: run build install clean test test-full lint tidy

BINARY := music-for-coding-tui
CMD     := ./cmd/mfp

run:
	go run $(CMD)

build:
	go build -o $(BINARY) $(CMD)

install:
	go install $(CMD)
	@echo "✓ mfp installed to $$(go env GOPATH)/bin/mfp"
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
