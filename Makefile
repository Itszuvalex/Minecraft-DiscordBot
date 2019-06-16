GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
BINARY_NAME=mcdiscord
BINARY_UNIX=$(BINARY_NAME)_unix
PACKAGE_NAME="github.com/itszuvalex/mcdiscord"
GOMAIN=cmd\mcdiscord\main.go

.PHONY: clean test docker-build $(BINARY_NAME) $(BINARY_UNIX)

all: test build
build: $(BINARY_NAME)
$(BINARY_NAME): 
	$(GOBUILD) -o $(BINARY_NAME) -v $(GOMAIN)

test:
	$(GOTEST) -v ./...
clean:
	$(GOCLEAN) $(GOMAIN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_UNIX)
run: build
	.\$(BINARY_NAME)
deps:
	$(GOGET) ...

build-linux: $(BINARY_UNIX)
$(BINARY_UNIX):
	scripts\BuildLinux.bat $(BINARY_UNIX)

docker-build: build-linux
	scripts\BuildLinuxDockerfile.bat $(BINARY_UNIX) $(BINARY_NAME)
