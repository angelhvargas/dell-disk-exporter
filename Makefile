# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOFMT=$(GOCMD) fmt
GOVET=$(GOCMD) vet
BINARY_NAME=dell-disk-exporter
BUILD_DIR=build

# Directories
IDRAC_DIR=pkg/idrac
SMART_DIR=pkg/smart

# Coverage
COVERAGE_DIR=coverage
COVERAGE_FILE=$(COVERAGE_DIR)/coverage.out
COVERAGE_HTML=$(COVERAGE_DIR)/coverage.html

.PHONY: all build test coverage fmt vet clean

all: build

build: 
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) -v

test: 
	$(GOTEST) -v ./...

coverage:
	mkdir -p $(COVERAGE_DIR)
	$(GOTEST) -coverprofile=$(COVERAGE_FILE) ./...
	$(GOCMD) tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	$(GOCMD) tool cover -func=$(COVERAGE_FILE)

fmt:
	$(GOFMT) ./...

vet:
	$(GOVET) ./...

clean: 
	$(GOCMD) clean
	rm -rf $(BUILD_DIR) $(COVERAGE_DIR)
