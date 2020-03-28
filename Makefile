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
	$(GO) run *.go configuration.yaml

.PHONY: test
test: ## Runs the go tests.
	@echo "+ $@"
	@$(GO) test -v -tags "$(BUILDTAGS) cgo" $(shell $(GO) list ./... | grep -v vendor)

.PHONY: cover
cover: ## Runs go test with coverage.
	@echo "" > coverage.txt
	@for d in $(shell $(GO) list ./... | grep -v vendor); do \
		$(GO) test -race -coverprofile=profile.out -covermode=atomic "$$d"; \
		if [ -f profile.out ]; then \
			cat profile.out >> coverage.txt; \
			rm profile.out; \
		fi; \
	done;

.PHONY: build
build: test
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

.PHONY: start-demo
start-demo: image
	@docker-compose -f docker/docker-compose.yaml up --build -d

.PHONY: stop-demo
stop-demo:
	@docker-compose -f docker/docker-compose.yaml down --remove-orphans
