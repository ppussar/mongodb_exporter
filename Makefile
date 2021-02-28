BIN      = $(CURDIR)/bin
IMAGE    = ppussar/mongodb_exporter
VERSION ?= $(shell git describe --tags --always --dirty 2> /dev/null || echo v0)
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
	@$(GO) test -v -tags "$(BUILDTAGS) cgo" $(shell $(GO) list ./... | grep -v wrapper | grep -v mocks)

.PHONY: test-functional
test-functional: ## Run functional tests
	@${MAKE} GOARGS="${GOARGS} -failfast -race -run ^TestFunctional.*\$$\$$" TEST_REPORT=functional test


.PHONY: test-integration-cmd
test-integration-cmd: ## Run integration tests cmd. Use 'test-integration' target
	@${MAKE} GOARGS="${GOARGS} -run ^TestIntegration.*\$$\$$" TEST_REPORT=integration test

.PHONY: test-integration
test-integration: clear ## Run the integration-tests in docker
	docker-compose -f build/docker-compose.test.yml build $$(echo ${DOCKER_BUILD_ARGS} | sed -e 's/--network host//')
	docker-compose -f build/docker-compose.test.yml up --exit-code-from tester

.PHONY: test-integration-clear
test-integration-clear: clear ## Cleans up the test env for docker
	docker-compose -f build/docker-compose.test.yml down --volumes


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

.PHONY: generate-mocks
generate-mocks: ## regenerates the mocks for the tests
	mockery -dir internal/wrapper -output internal/mocks -name IConnection
	mockery -dir internal/wrapper -output internal/mocks -name ICursor
