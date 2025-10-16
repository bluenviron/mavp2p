BASE_IMAGE = golang:1.25-alpine3.22
LINT_IMAGE = golangci/golangci-lint:v2.5.0

.PHONY: $(shell ls)

help:
	@echo "usage: make [action] [args...]"
	@echo ""
	@echo "available actions:"
	@echo ""
	@echo "  format        format source files"
	@echo "  test          run tests"
	@echo "  lint          run linter"
	@echo "  binaries      build binaries for all platforms"
	@echo ""

include scripts/*.mk
