FROM golang:1-bullseye

MAINTAINER thomas@leroux.io

ENV DEBIAN_FRONTEND noninteractive
ENV LANG C.UTF-8
ENV LC_ALL C.UTF-8

RUN apt-get -y update \
    && apt-get upgrade -y \
    && apt-get -y install git \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/* \
    && useradd -ms /bin/bash gopher

COPY go.mod go.sum /media/ulule/limiter/
RUN chown -R gopher:gopher /media/ulule/limiter
ENV GOPATH /home/gopher/go
ENV PATH $GOPATH/bin:$PATH
USER gopher

RUN go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

WORKDIR /media/ulule/limiter
RUN go mod download
COPY --chown=gopher:gopher . /media/ulule/limiter

CMD [ "/bin/bash" ]
