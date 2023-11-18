# Go parameters
GOVARS=CGO_ENABLED=0
GOCMD=go
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOVARS) $(GOCMD) test
GOTOOL=$(GOCMD) tool
GOVET=$(GOCMD) vet
GOGET=$(GOVARS) $(GOCMD) get
GOBUILDOPT=-a -ldflags "-X main.buildnumber=$(BUILDNUMBER) -X main.giturl=$(GITURL) -X main.githash=$(subst $(SPACE),$(UNDERSCORE),$(GITHASH)) -X main.buildstamp=$(TIMESTAMP)"
GOBUILD=$(GOVARS) $(GOCMD) build $(GOBUILDOPT)
GOBUILDOUT=-o bin/${@:build/./cmd/%=%}$(BINARY_EXT)

GOFMT=$(GOCMD) fmt
GOSTATICCHECK=staticcheck
GOIMPORTS=goimports

GOOUTDIR=bin

TMPLMODULE=github.com/mmelnyk/golang-project-layout
TMPLMARKER=.go-layout

UNDERSCORE:= _
EMPTY:=
SPACE:= $(EMPTY) $(EMPTY)

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

ifneq (,$(wildcard $(TMPLMARKER)))
    NEEDED_INIIALIZATION :=yes
endif
