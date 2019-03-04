export GOPATH?=$(HOME)/go
export GOBIN?=$(GOPATH)/bin
export PATH:=$(GOBIN):$(PATH)

APP_NAME:=hrapp

GOGENERATE=go generate
GOTEST:=go test -race

BINDIR:=bin
PKGS     = $(or $(PKG),$(shell env GO111MODULE=on $(GO) list ./...))

GO=go
export GO111MODULE=on

.PHONY: bclient
bclient: proto
	 GOOS=linux GOARCH=amd64 $(GO) build \
		-o $(BINDIR)/$(APP_NAME)-client client/hrapp.go

.PHONY: build
build: bclient
	 GOOS=linux GOARCH=amd64 $(GO) build \
		-o $(BINDIR)/$(APP_NAME) cmd/main.go

.DEFAULT_GOAL:=hrapp-docker

.PHONY: proto
proto:
	@if ! which protoc > /dev/null; then \
		echo "error: protoc not installed" >&2; \
		exit 1; \
	fi
	@$(GOGENERATE) $(PACKAGES)

.PHONY: test
test: proto
	@$(GOTEST) $(PKGS)

.PHONY: cover
cover: proto
	@$(GOTEST) -cover $(PKGS)

.PHONY: bench
bench: proto
	@$(GOTEST) -bench=. -benchmem $(PKGS) -run ^Benchmark

.PHONY: hrapp-docker
hrapp-docker: build
	@docker-compose -f docker-compose.yml build --no-cache
	@echo "Finished building '$@'"

.PHONY: hrapp-up
hrapp-up:
	@docker-compose -f docker-compose.yml up -d hrapp

.PHONY: cassandra-up
cassandra-up:
	@docker-compose -f docker-compose.yml up -d cassandra

.PHONY: down
down:
	@docker-compose -f docker-compose.yml down

.PHONY: clean
clean:
	@rm -rf {bin,tmp}
	@for x in $(GENERATED_SRC); do \
		rm -f $$x; \
	done