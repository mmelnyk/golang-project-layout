GOBUILDOUTMP=-o bin/${@:build.go/%=%}$(BINARY_EXT)

TARGETS:= $(foreach P,$(BINARIES:./cmd/%=%),$(addprefix $P-,$(BUILDS)))

.PHONY=buildmp build.go/%
buildmp: _bindir $(TARGETS:%=build.go/%) ## do project build
build.go/%: BIN = $(word 1,$(subst -, ,$*))
build.go/%: export GOOS = $(word 2,$(subst -, ,$*))
build.go/%: export GOARCH = $(word 3,$(subst -, ,$(subst ., ,$*)))
build.go/%: BINARY_EXT = $(if $(filter windows, $(word 2,$(subst -, ,$*))),.exe,$(EMPTY))
build.go/%: prebuild
	$(GOBUILD) $(GOBUILDOUTMP) ./cmd/$(BIN)

.PHONY=cleanmp cleanmp/%
cleanmp: $(TARGETS:%=cleanmp/%) ## clean up files
	@$(GOCLEAN)
	@rm -f $(GOOUTDIR)/cover.out
	@rm -f $(GOOUTDIR)/cover.html

cleanmp/%: BINARY_EXT = $(if $(filter windows, $(word 2,$(subst -, ,$*))),.exe,$(EMPTY))
cleanmp/%:
	@rm -f ${@:cleanmp/%=$(GOOUTDIR)/%}$(BINARY_EXT)
