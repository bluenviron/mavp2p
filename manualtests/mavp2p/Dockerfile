FROM amd64/golang:1.11-stretch

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./
RUN go install .

ENTRYPOINT [ "mavp2p" ]
