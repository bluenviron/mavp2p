define DOCKERFILE_TEST
FROM $(BASE_IMAGE)
RUN apk add --no-cache make gcc musl-dev
WORKDIR /s
COPY go.mod go.sum ./
RUN go mod download
endef
export DOCKERFILE_TEST

test-pkg:
	go test -v -race -coverprofile=coverage-pkg.txt ./pkg/...

test-root:
	go test -v -race -coverprofile=coverage-root.txt .

test-nodocker: test-pkg test-root

test:
	echo "$$DOCKERFILE_TEST" | docker build -q . -f - -t temp
	docker run --rm -v "$(shell pwd):/s" temp make test-nodocker
