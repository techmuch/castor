.PHONY: all build run tui test clean

BINARY_NAME=castor

all: build

build:
	go build -o $(BINARY_NAME) ./cmd/castor

run: build
	./$(BINARY_NAME) -i

tui: build
	./$(BINARY_NAME) -tui

test:
	go test -v ./...

clean:
	rm -f $(BINARY_NAME)
	rm -f session.json
