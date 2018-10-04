FROM golang:1.11.1-stretch

MAINTAINER thomas.leroux@ulule.com

ENV DEBIAN_FRONTEND noninteractive
ENV LANG C.UTF-8
ENV LC_ALL C.UTF-8

RUN apt-get -y update \
    && apt-get upgrade -y \
    && apt-get -y install git \
    && mkdir -p /go/src/github.com/golang \
    && cd /go/src/github.com/golang \
    && git clone https://github.com/golang/dep.git \
    && cd dep \
    && git checkout tags/v0.4.1 \
    && go get github.com/golang/dep/cmd/dep \
    && go get -u -v github.com/stretchr/testify/require \
    && go get -u -v github.com/davecgh/go-spew/spew \
    && go get -u -v github.com/alecthomas/gometalinter \
    && gometalinter --install \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

COPY . /go/src/github.com/ulule/limiter
WORKDIR /go/src/github.com/ulule/limiter
RUN dep ensure

CMD /bin/bash
