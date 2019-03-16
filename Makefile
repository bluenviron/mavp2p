
.PHONY: help
help:
	@echo "usage: make [action] [args...]"
	@echo ""
	@echo "available actions:"
	@echo ""
	@echo "  format            format source files."
	@echo ""


.PHONY: format
format:
	@docker run --rm -it \
		-v $(PWD):/src \
		amd64/golang:1.11-stretch \
		sh -c "cd /src \
		&& find . -type f -name '*.go' | xargs gofmt -l -w -s"


# -ldflags "-X main.Rev=`git rev-parse --short HEAD`"
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
