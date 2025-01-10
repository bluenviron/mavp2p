define DOCKERFILE_TEST
FROM $(BASE_IMAGE)
RUN apk add --no-cache make gcc musl-dev
WORKDIR /s
COPY go.mod go.sum ./
RUN go mod download
endef
export DOCKERFILE_TEST

test:
	echo "$$DOCKERFILE_TEST" | docker build -q . -f - -t temp
	docker run --rm \
	-v $(shell pwd):/s -w /s \
	temp \
	go test -v -race -coverprofile=coverage.txt .
