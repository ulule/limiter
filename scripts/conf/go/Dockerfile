FROM golang:1.12.7-stretch

MAINTAINER thomas@leroux.io

ENV DEBIAN_FRONTEND noninteractive
ENV LANG C.UTF-8
ENV LC_ALL C.UTF-8

RUN apt-get -y update \
    && apt-get upgrade -y \
    && apt-get -y install git \
    && mkdir -p /media/ulule/limiter \
    && go get -u -v -d github.com/stretchr/testify/require \
    && go get -u -v -d github.com/davecgh/go-spew/spew \
    && go get -u -d -v github.com/golangci/golangci-lint/cmd/golangci-lint \
    && go install github.com/golangci/golangci-lint/cmd/golangci-lint \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

COPY . /media/ulule/limiter
WORKDIR /media/ulule/limiter
RUN go mod download

CMD [ "/bin/bash" ]
