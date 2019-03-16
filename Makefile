
.PHONY: help
help:
	@echo "usage: make [action] [args...]"
	@echo ""
	@echo "available actions:"
	@echo ""
	@echo "  format       format source files."
	@echo ""
	@echo "  buildall     build for all supported platforms."
	@echo ""


.PHONY: format
format:
	@docker run --rm -it \
		-v $(PWD):/src \
		amd64/golang:1.11-stretch \
		sh -c "cd /src \
		&& find . -type f -name '*.go' | xargs gofmt -l -w -s"


.PHONY: buildall
buildall:
	@docker run --rm -it \
		-v $(PWD):/src \
		amd64/golang:1.11-stretch \
		sh -c "cd /src \
		&& make buildall-nodocker"


.PHONY: buildall-nodocker
buildall-nodocker:
	@rm -rf build && mkdir build
	GOOS=linux GOARCH=amd64        go build -o build/mavp2p_linux_amd64
	GOOS=linux GOARCH=arm GOARM=6  go build -o build/mavp2p_linux_arm6
	GOOS=linux GOARCH=arm GOARM=7  go build -o build/mavp2p_linux_arm7
	GOOS=linux GOARCH=arm64        go build -o build/mavp2p_linux_arm64
	GOOS=windows GOARCH=amd64      go build -o build/mavp2p_windows_amd64.exe


.PHONY: travis-setup-releases
define TRAVIS_DOCKERFILE
FROM ruby:alpine
RUN apk add --no-cache build-base git \
	&& gem install travis
endef
export TRAVIS_DOCKERFILE
travis-setup-releases:
	@echo "$$TRAVIS_DOCKERFILE" | docker build - -t travis-setup-releases
	docker run --rm -it \
		-v $(PWD):/src \
		travis-setup-releases \
		sh -c "cd /src \
		&& travis setup releases"
