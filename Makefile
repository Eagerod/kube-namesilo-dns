GO := go

MAIN_FILE := main.go

BUILD_DIR := build
EXECUTABLE := nsdns
BIN_NAME := $(BUILD_DIR)/$(EXECUTABLE)
INSTALLED_NAME := /usr/local/bin/$(EXECUTABLE)

CMD_PACKAGE_DIR := ./cmd/$(EXECUTABLE)
PKG_PACKAGE_DIR := ./pkg/*
PACKAGE_PATHS := $(CMD_PACKAGE_DIR) $(PKG_PACKAGE_DIR)

AUTOGEN_VERSION_FILENAME=./cmd/$(EXECUTABLE)/version-temp.go
COVERAGE_FILE=./coverage.out

ALL_GO_DIRS = $(shell find . -iname "*.go" -exec dirname {} \; | sort | uniq)
SRC := $(shell find . -iname "*.go" -and -not -name "*_test.go") $(AUTOGEN_VERSION_FILENAME)
SRC_WITH_TESTS := $(shell find . -iname "*.go") $(AUTOGEN_VERSION_FILENAME)

# Publish targets are treated as phony to force rebuilds.
PUBLISH_DIR=publish
PUBLISH := \
	$(PUBLISH_DIR)/linux-amd64 \
	$(PUBLISH_DIR)/darwin-amd64 \
	$(PUBLISH_DIR)/darwin-arm64

.PHONY: $(PUBLISH)


.PHONY: all
all: $(BIN_NAME)

$(BIN_NAME): $(SRC)
	@mkdir -p $(BUILD_DIR)
	$(GO) build -o $(BIN_NAME) $(MAIN_FILE)


.PHONY: publish
publish: $(PUBLISH)

$(PUBLISH):
	rm -f $(BIN_NAME)
	GOOS_GOARCH="$$(basename $@)" \
	GOOS="$$(cut -d '-' -f 1 <<< "$$GOOS_GOARCH")" \
	GOARCH="$$(cut -d '-' -f 2 <<< "$$GOOS_GOARCH")" \
		$(MAKE) $(BIN_NAME)
	mkdir -p $$(dirname "$@")
	mv $(BIN_NAME) $@


.PHONY: install isntall
install isntall: $(INSTALLED_NAME)

$(INSTALLED_NAME): $(BIN_NAME)
	cp $(BIN_NAME) $(INSTALLED_NAME)

.PHONY: test
test: $(SRC) $(BIN_NAME)
	@if [ -z $$T ]; then \
		$(GO) test -v ./...; \
	else \
		$(GO) test -v ./... -run $$T; \
	fi


$(COVERAGE_FILE): $(SRC_WITH_TESTS)
	$(GO) test -v --coverprofile=$(COVERAGE_FILE) ./...

.PHONY: coverage
coverage: $(COVERAGE_FILE)
	$(GO) tool cover -func=$(COVERAGE_FILE)

.PHONY: pretty-coverage
pretty-coverage: $(COVERAGE_FILE)
	$(GO) tool cover -html=$(COVERAGE_FILE)

.INTERMEDIATE: $(AUTOGEN_VERSION_FILENAME)
$(AUTOGEN_VERSION_FILENAME):
	@version="$${VERSION:-$$(git describe --dirty)}"; \
	printf "package cmd\n\nconst VersionBuild = \"%s\"" "$$version" > $@

.PHONY: fmt
fmt:
	@$(GO) fmt ./...

.PHONY: clean
clean:
	rm -rf $(COVERAGE_FILE) $(BUILD_DIR) $(PUBLISH_DIR)
