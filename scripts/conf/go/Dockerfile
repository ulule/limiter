FROM golang:1.13-buster

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

COPY . /media/ulule/limiter
RUN chown -R gopher:gopher /media/ulule/limiter
ENV GOPATH /home/gopher/go
ENV PATH $GOPATH/bin:$PATH
USER gopher

RUN go get -u github.com/golangci/golangci-lint/cmd/golangci-lint

WORKDIR /media/ulule/limiter
RUN go mod download

CMD [ "/bin/bash" ]
