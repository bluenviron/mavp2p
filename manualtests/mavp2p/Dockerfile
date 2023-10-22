FROM golang:1.21

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./
RUN go install .

ENTRYPOINT [ "mavp2p" ]
