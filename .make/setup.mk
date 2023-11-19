# Go parameters
GOVARS=CGO_ENABLED=0
GOCMD=go
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOVARS) $(GOCMD) test
GOTOOL=$(GOCMD) tool
GOVET=$(GOCMD) vet
GOGET=$(GOVARS) $(GOCMD) get
GOLDGLAGS=-X main.buildnumber=$(BUILDNUMBER) -X main.giturl=$(GITURL) -X main.binary=$(BINARY_OUT) -X main.githash=$(subst $(SPACE),$(UNDERSCORE),$(GITHASH)) -X main.buildstamp=$(TIMESTAMP)
GOBUILDOPT=-a -ldflags "$(GOLDGLAGS) $(GOLDFLAGSEXTRA)"
GOBUILD=$(GOVARS) $(GOCMD) build $(GOTAGS) $(GOBUILDOPT)
GOBUILDOUT=-o bin/${@:build/./cmd/%=%}$(BINARY_EXT)

BINARY_A = $(@:build.go/%=%)
BINARY_OUT = $(BINARY_A:build/./cmd/%=%)$(BINARY_EXT)

GOFMT=$(GOCMD) fmt
GOSTATICCHECK=staticcheck
GOIMPORTS=goimports
MSIGN=msign

GOOUTDIR=bin

TMPLMODULE=github.com/mmelnyk/golang-project-layout
TMPLMARKER=.go-layout

UNDERSCORE:= _
EMPTY:=
SPACE:= $(EMPTY) $(EMPTY)

# Setup features build tags
ifeq ($(FEATURE_SHOW_VERSION),yes)
	GOTAGS := $(GOTAGS)showversion,
endif

ifeq ($(FEATURE_SELF_UPDATE),yes)
	GOTAGS := $(GOTAGS)selfupdate,
	ifeq ($(MSIGN_SIGNATURE),yes)
		GOLDFLAGSEXTRA := $(GOLDFLAGSEXTRA) -X $(GOMODULE)/internal/selfupdate.msignPublic=$(MSIGN_PUBLIC)
	endif
endif

ifneq ($(GOTAGS),)
	GOTAGS := -tags "$(GOTAGS)"
endif

# Platform specific settings
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

# Initialization detection
ifneq (,$(wildcard $(TMPLMARKER)))
    NEEDED_INIIALIZATION :=yes
endif
