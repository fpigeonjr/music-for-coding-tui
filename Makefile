.PHONY: run build clean

BINARY := music-for-coding-tui
CMD     := ./cmd/mfp

run:
	go run $(CMD)

build:
	go build -o $(BINARY) $(CMD)

clean:
	rm -f $(BINARY)

lint:
	go vet ./...

tidy:
	go mod tidy
