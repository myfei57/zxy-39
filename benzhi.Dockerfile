FROM golang:1.23.12

ENV GOPROXY=off \
    GOSUMDB=off

WORKDIR /app

COPY go.mod go.sum ./
COPY vendor ./vendor
COPY cmd ./cmd
COPY internal ./internal

RUN go build -mod=vendor ./...

CMD ["bash"]
