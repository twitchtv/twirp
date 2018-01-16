#!/usr/bin/env bash

which protoc
PROTOC_EXISTS=$?
if [ $PROTOC_EXISTS -eq 0 ]; then
    echo "Protoc already installed"
    exit 0
fi

if [ "$(uname)" == "Darwin" ]; then
    brew install protobuf
elif [ `whoami` == "root" ]; then
    mkdir -p /usr/local/src/protoc
    pushd /usr/local/src/protoc
    wget https://github.com/google/protobuf/releases/download/v3.1.0/protoc-3.1.0-linux-x86_64.zip -O /usr/local/src/protoc-3.1.0-linux-x86_64.zip
    unzip -x ../protoc-3.1.0-linux-x86_64.zip
    if [ ! -e /usr/local/bin/protoc ]; then
        ln -s `pwd`/bin/protoc /usr/local/bin/protoc
    fi
    popd
elif [ "$(expr substr $(uname -s) 1 5)" == "Linux" ]; then
    echo "Project setup needs sudo to put protoc in /usr/local/src, so it will ask a few times"
    sudo chmod a+w /usr/local/src
    mkdir -p /usr/local/src/protoc
    pushd /usr/local/src/protoc
    wget https://github.com/google/protobuf/releases/download/v3.1.0/protoc-3.1.0-linux-x86_64.zip -O /usr/local/src/protoc-3.1.0-linux-x86_64.zip
    unzip -x ../protoc-3.1.0-linux-x86_64.zip
    if [ ! -e /usr/local/bin/protoc ]; then
        sudo ln -s `pwd`/bin/protoc /usr/local/bin/protoc
    fi
    popd
fi
exit 0