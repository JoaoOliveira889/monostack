BINARY := monostack
CMD := ./cmd/monostack
LOCAL_BIN ?= $(HOME)/bin
GO ?= go

.PHONY: build install-local clean test test-cover lint

build:
	$(GO) build -o $(BINARY) $(CMD)

install-local:
	mkdir -p $(LOCAL_BIN)
	$(GO) build -o $(LOCAL_BIN)/$(BINARY) $(CMD)

clean:
	rm -f $(BINARY)

test:
	$(GO) test ./...

test-cover:
	$(GO) test -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html

lint:
	golangci-lint run ./...
