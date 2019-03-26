FROM golang:1.12

RUN apt-get update &&\
    apt-get install -y --no-install-recommends unzip python-dev python-pip && \
    pip install virtualenv

WORKDIR /go/src/github.com/twitchtv/twirp
COPY . .

ENTRYPOINT ["make"]
