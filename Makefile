.PHONY: all
all: build

.PHONY: build
build:
	CGO_ENABLED=0 go build
