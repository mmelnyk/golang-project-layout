# Import project parameters
include Project

# Go parameters
GOVARS=CGO_ENABLED=0
GOCMD=go
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOVARS) $(GOCMD) test
GOTOOL=$(GOCMD) tool
GOVET=$(GOCMD) vet
GOGET=$(GOVARS) $(GOCMD) get
GOBUILDOPT=-a -ldflags "-X main.buildnumber=$(BUILDNUMBER)  -X main.githash=$(subst $(SPACE),$(UNDERSCORE),$(GITHASH)) -X main.buildstamp=$(TIMESTAMP)"
GOBUILD=$(GOVARS) $(GOCMD) build $(GOBUILDOPT)
GOBUILDOUT=-o bin/${@:build/./cmd/%=%}$(BINARY_EXT)
GOFMT=$(GOCMD) fmt
GOLINT=golint -set_exit_status
GOIMPORTS=goimports
GOFUZZBUILD=go-fuzz-build
GOFUZZ=go-fuzz

GOOUTDIR=bin

TMPLMODULE=github.com/mmelnyk/golang-project-layout
TMPLMARKER=.go-layout

UNDERSCORE:= _
EMPTY:=
SPACE:= $(EMPTY) $(EMPTY)

HELP_SEL= \033[36m
HELP_NORM= \033[0m

ifeq ($(OS),Windows_NT)
    uname := Windows
else
    uname := $(shell uname)
endif

BINARY_EXT :=
ifeq ($(GOOS),windows) ## Use .exe if our target platform is Windows
	BINARY_EXT := .exe
endif
ifeq ($(uname),Windows) ## On Windows...
ifeq ($(GOOS),) ## ... use .exe if there are no specified target platform
	BINARY_EXT := .exe
endif
endif

SEDI := sed -i
ifeq ($(uname),Darwin) # Mac OS X
    SEDI := sed -i ""
endif

GETDOCKERFILE=$(firstword $(subst ~, ,$1))
GETDOCKERIMAGE=$(or $(word 2,$(subst ~, ,$1)),$(value 2))

.PHONY=all
all: check test build ## Do check - test - build for the project

.PHONY=prebuild
prebuild:
	$(eval TIMESTAMP != date -u '+%Y-%m-%d_%I:%M:%S%p')
	$(eval GITHASH := $(if $(GITHASH), $(GITHASH), $(shell git rev-parse HEAD || echo N/A )))
	$(eval BUILDNUMBER ?= DEVBUILD)

_bindir:
	@mkdir -p $(GOOUTDIR)

.PHONY=build build/%
build: _bindir $(BINARIES:%=build/%) ## do project build
build/%: prebuild
	$(GOBUILD) $(GOBUILDOUT) ${@:build/%=%}

.PHONY=test
test: _bindir prebuild  ## run unit tests with code coverage info
	$(GOTEST) -cover -coverprofile=$(GOOUTDIR)/cover.out -covermode=atomic ./...
	$(GOTOOL) cover -html=$(GOOUTDIR)/cover.out -o $(GOOUTDIR)/cover.html

.PHONY=check
check: check.vet check.lint ## do static code checks

.PHONY=check.vet
check.vet: ## do go vet checks
	$(GOVET) ./...

.PHONY=check.lint
check.lint: tools.golint ## do golint checks
	$(GOLINT) ./...

.PHONY=check.fmt
check.fmt: ## do check for right formating
	@test -z "$(shell gofmt -s -l . | tee /dev/stderr)" || (echo "Formating is needed (please do 'make format')"; false)

.PHONY=clean clean/%
clean: $(BINARIES:./cmd/%=clean/%) ## clean up files
	$(GOCLEAN)
	rm -f $(GOOUTDIR)/cover.out
	rm -f $(GOOUTDIR)/cover.html

clean/%:
	rm -f ${@:clean/%=$(GOOUTDIR)/%}$(BINARY_EXT)

.PHONY=docker docker/% docker.push docker.push/%
docker: prebuild $(DOCKER_IMAGES:%=docker/%) ## build docker image
docker/%:
	@echo Build image $(DOCKER_REGISTRY)/$(call GETDOCKERIMAGE,$(@:docker/%=%)):$(BUILDNUMBER) from docker/$(call GETDOCKERFILE,$(@:docker/%=%))
	docker build -t $(DOCKER_REGISTRY)/$(call GETDOCKERIMAGE,$(@:docker/%=%)):$(BUILDNUMBER) -f docker/$(call GETDOCKERFILE,$(@:docker/%=%)) .

docker.push: $(DOCKER_IMAGES:%=docker.push/%) ## push docker images to the registry
docker.push/%:
	@echo Push docker image $(DOCKER_REGISTRY)/$(call GETDOCKERIMAGE,$@):$(BUILDNUMBER)
	docker push $(DOCKER_REGISTRY)/$(call GETDOCKERIMAGE,$@):$(BUILDNUMBER)

_empty:
list: ## show available targets
	@echo $(shell $(MAKE) -p _empty | grep "^[a-z]*:" | cut -d ":" -f1 | sort)

help: ## this help
	@awk 'BEGIN {FS = ":.*?## "} /^[.a-zA-Z_-]+:.*?## / {printf "$(HELP_SEL)%-15s$(HELP_NORM)%s\n", $$1, $$2}' $(MAKEFILE_LIST) | sort | cat

.PHONY=format
format: ## format go code (via gofmt)
	$(GOFMT) ./...

# setup and configure section
.PHONY=init

init: init.replace init.githooks init.gomod ## setup current project

init.from-tmpl:
ifeq ($(TMPLMARKER),$(wildcard $(TMPLMARKER)))
	-rm $(TMPLMARKER)
	-rm -Rf .git
	-mv README.md README-LAYOUT.md
	-mv README-TEMPLATE.md README.md
	-mv github .github
	git init
	@git add .github .githooks .vscode .editorconfig .gitignore
endif

init.git: init.from-tmpl
ifeq (,$(wildcard .git))
	git init
endif

init.githooks: init.git
	git config core.hooksPath .githooks

init.gomod:
ifeq (,$(wildcard go.mod))
	$(GOCMD) mod init $(GOMODULE)
endif

init.replace:
	-@$(SEDI) "s~$(TMPLMODULE)~$(GOMODULE)~g" *.md
	-@$(SEDI) "s~$(TMPLMODULE)~$(GOMODULE)~g" go.mod
	-@$(SEDI) "s~$(TMPLMODULE)~$(GOMODULE)~g" */*/*.go
	-@$(SEDI) "s~$(TMPLMODULE)~$(GOMODULE)~g" github/*/*.md

# tools section
.PHONY=tools tools.goimports tools.golint tools.gofuzz

tools: tools.goimports tools.golint tools.gofuzz ## install all required tools

tools.goimports:
	@command -v $(GOIMPORTS) >/dev/null ; if [ $$? -ne 0 ]; then \
		echo "[ installing goimports ]"; \
		go get golang.org/x/tools/cmd/goimports; \
	fi

tools.golint:
	@command -v $(GOLINT) >/dev/null ; if [ $$? -ne 0 ]; then \
		echo "[ installing golint ]"; \
		go get -u golang.org/x/lint/golint; \
	fi

tools.gofuzz:
	@command -v $(GOFUZZBUILD) >/dev/null ; if [ $$? -ne 0 ]; then \
		echo "[ installing go-fuzz-build ]"; \
		go get -u github.com/dvyukov/go-fuzz/go-fuzz-build; \
	fi
	@command -v $(GOFUZZ) >/dev/null ; if [ $$? -ne 0 ]; then \
		echo "[ installing go-fuzz ]"; \
		go get -u github.com/dvyukov/go-fuzz/go-fuzz; \
	fi
