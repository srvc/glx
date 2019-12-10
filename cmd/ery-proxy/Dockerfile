FROM golang:1.12-alpine3.10 as builder

ENV GO111MODULE on
WORKDIR /go/github.com/srvc/glx

RUN apk add --update --no-cache git

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
RUN go build -o /usr/local/bin/glx-proxy ./cmd/glx-proxy

FROM alpine:3.10

COPY --from=builder /usr/local/bin/glx-proxy /usr/local/bin/
