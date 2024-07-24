.PHONY: all
all: build

.PHONY: build
build:
	CGO_ENABLED=0 go build

.PHONY: clean
clean:
	rm -f wylmo
	rm -rf hard_timeout
	rm -rf inactivity_timeout
