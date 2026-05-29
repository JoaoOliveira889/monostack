BINARY := monostack
CMD := ./cmd/monostack
LOCAL_BIN ?= $(HOME)/bin
GO ?= go

.PHONY: build install-local clean

build:
	$(GO) build -o $(BINARY) $(CMD)

install-local:
	mkdir -p $(LOCAL_BIN)
	$(GO) build -o $(LOCAL_BIN)/$(BINARY) $(CMD)

clean:
	rm -f $(BINARY)
