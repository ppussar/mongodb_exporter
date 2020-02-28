BIN      = $(CURDIR)/bin
IMAGE    = ppussar/mongodb_exporter
VERSION ?= $(shell git describe --tags --always --dirty --match=v* 2> /dev/null || echo v0)
GO      = go
TIMEOUT = 15

.PHONY: all
all:
	@echo $(VERSION)

.PHONY: run
run:
	$(GO) run *.go docker/configuration.yaml

.PHONY: build
build:
	$(GO) build -v -o $(BIN)/mongodb_exporter .

.PHONY: clean
clean:
	@rm -rf $(BIN)

.PHONY: image
image:
	CGO_ENABLED=0 GOOS=linux $(GO) build -a -installsuffix cgo -o $(BIN)/mongodb_exporter .
	@cp docker/Dockerfile $(BIN)
	@docker build -t $(IMAGE):$(VERSION) $(BIN)

.PHONY: push-image
push-image:
	@docker push $(IMAGE):$(VERSION)

.PHONY: startdb
startdb: dockerize
	@docker-compose -f docker/docker-compose.yaml up --build

.PHONY: stopdb
stopdb:
	@docker-compose -f docker/docker-compose.yaml down --remove-orphans
