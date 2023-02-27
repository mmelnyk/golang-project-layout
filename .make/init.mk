# setup and configure section
.PHONY=init

init: init.check init.replace init.githooks init.gomod ## setup the project

ifeq ($(TMPLMODULE), $(GOMODULE))
	FAILED_MESSAGE=Please change the module name in Project file
else ifeq ($(BUILD_DOCKER),yes)
	ifndef DOCKER_IMAGES
		FAILED_MESSAGE=Please define DOCKER_IMAGES variable in Project file
	endif
endif

init.check:
ifdef FAILED_MESSAGE
	@echo $(FAILED_MESSAGE)
	@exit 1
endif

init.from-tmpl:
ifeq ($(TMPLMARKER),$(wildcard $(TMPLMARKER)))
	-rm $(TMPLMARKER)
	-rm -Rf .git
	-mv README.md README-LAYOUT.md
	-mv README-TEMPLATE.md README.md
	-mv github .github
	git init
	@git add .github .githooks .make .vscode .editorconfig .gitignore
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
	$(GOCMD) get -u
	$(GOCMD) mod tidy

init.replace:
	-@$(SEDI) "s~$(TMPLMODULE)~$(GOMODULE)~g" *.md
	-@$(SEDI) "s~$(TMPLMODULE)~$(GOMODULE)~g" go.mod
	-@$(SEDI) "s~$(TMPLMODULE)~$(GOMODULE)~g" */*/*.go
	-@$(SEDI) "s~$(TMPLMODULE)~$(GOMODULE)~g" github/*/*.md
