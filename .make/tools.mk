
# tools section
.PHONY=tools tools.goimports tools.staticcheck

tools: tools.goimports tools.staticcheck ## install all required tools

tools.goimports:
	@command -v $(GOIMPORTS) >/dev/null ; if [ $$? -ne 0 ]; then \
		echo "[ installing goimports ]"; \
		go install golang.org/x/tools/cmd/goimports@latest; \
	fi

tools.staticcheck:
	@command -v $(GOSTATICCHECK) >/dev/null ; if [ $$? -ne 0 ]; then \
		echo "[ installing staticcheck ]"; \
		o install honnef.co/go/tools/cmd/staticcheck@latest; \
	fi
