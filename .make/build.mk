.PHONY=all
all: check test build ## Do check - test - build for the project

.PHONY=prebuild
prebuild:
	$(eval TIMESTAMP ?= $(shell date -u '+%Y-%m-%d_%I:%M:%S%p'))
	$(eval GITHASH := $(if $(GITHASH), $(GITHASH), $(shell git rev-parse HEAD 2>/dev/null|| echo N/A )))
	$(eval BUILDNUMBER ?= DEVBUILD)
	$(eval GITURL ?= $(shell git config --get remote.origin.url 2>/dev/null|| echo N/A))

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
check: check.vet check.static ## do static code checks

.PHONY=check.vet
check.vet: ## do go vet checks
	$(GOVET) ./...

.PHONY=check.static
check.static: tools.staticcheck ## do staticcheck checks
	$(GOSTATICCHECK) ./...

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

.PHONY=format
format: ## format go code (via gofmt)
	$(GOFMT) ./...
