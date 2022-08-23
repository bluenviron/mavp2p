BASE_IMAGE = golang:1.17-alpine3.14
LINT_IMAGE = golangci/golangci-lint:v1.45.2

.PHONY: $(shell ls)

help:
	@echo "usage: make [action] [args...]"
	@echo ""
	@echo "available actions:"
	@echo ""
	@echo "  mod-tidy      run go mod tidy"
	@echo "  format        format source files"
	@echo "  run           run app"
	@echo "  test          run tests"
	@echo "  lint          run linter"
	@echo "  release       build release assets for all platforms"
	@echo ""

include scripts/*.mk
