FROM golang:1.23.2-alpine

RUN apk add --no-cache git bash make curl gcc libc-dev openssl
RUN go install golang.org/x/tools/cmd/goimports@latest

WORKDIR /go/src/clean-serverless-book-sample
COPY ../../app /go/src/clean-serverless-book-sample

RUN ./scripts/build-handlers.sh
