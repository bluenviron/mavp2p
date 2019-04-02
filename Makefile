
help:
	@echo "usage: make [action] [args...]"
	@echo ""
	@echo "available actions:"
	@echo ""
	@echo "  mod-tidy      run go mod tidy"
	@echo "  format        format source files"
	@echo "  release       build release assets for all platforms"
	@echo "  travis-setup  set up travis for automatic releases"
	@echo ""

mod-tidy:
	docker run --rm -it -v $(PWD):/src \
		amd64/golang:1.11 \
		sh -c "cd /src && go get -m ./... && go mod tidy"

format:
	@docker run --rm -it \
		-v $(PWD):/src \
		amd64/golang:1.11-stretch \
		sh -c "cd /src \
		&& find . -type f -name '*.go' | xargs gofmt -l -w -s"

.PHONY: release
release:
	echo "FROM amd64/golang:1.11-stretch \n\
		RUN apt-get update && apt-get install -y zip \n\
		WORKDIR /src \n\
		COPY go.mod go.sum ./ \n\
		RUN go mod download \n\
		COPY .git ./.git \n\
		COPY *.go Makefile ./ \n\
		RUN make release-nodocker" | docker build . -f - -t mavp2p-release \
		&& docker run --rm -it -v $(PWD):/out \
		mavp2p-release sh -c "rm -rf /out/release && cp -r /src/release /out/"

release-nodocker:
	$(eval VERSION := $(shell git describe --tags))
	$(eval GOBUILD := go build -ldflags '-X "main.Version=$(VERSION)"')
	rm -rf release && mkdir release

	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) -o /tmp/mavp2p.exe
	cd /tmp && zip -q $(PWD)/release/mavp2p_$(VERSION)_windows_amd64.zip mavp2p.exe

	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o /tmp/mavp2p
	tar -C /tmp -czf $(PWD)/release/mavp2p_$(VERSION)_linux_amd64.tar.gz --owner=0 --group=0 mavp2p

	CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=6 $(GOBUILD) -o /tmp/mavp2p
	tar -C /tmp -czf $(PWD)/release/mavp2p_$(VERSION)_linux_arm6.tar.gz --owner=0 --group=0 mavp2p

	CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 $(GOBUILD) -o /tmp/mavp2p
	tar -C /tmp -czf $(PWD)/release/mavp2p_$(VERSION)_linux_arm7.tar.gz --owner=0 --group=0 mavp2p

	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 $(GOBUILD) -o /tmp/mavp2p
	tar -C /tmp -czf $(PWD)/release/mavp2p_$(VERSION)_linux_arm64.tar.gz --owner=0 --group=0 mavp2p

travis-setup:
	@echo "FROM ruby:alpine \n\
		RUN apk add --no-cache build-base git \n\
		RUN gem install travis" | docker build - -t mavp2p-travis-sr \
		&& docker run --rm -it \
		-v $(PWD):/src \
		mavp2p-travis-sr \
		sh -c "cd /src \
		&& travis setup releases"
