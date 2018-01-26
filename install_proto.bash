#!/usr/bin/env bash

# Copyright 2018 Twitch Interactive, Inc.  All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License"). You may not
# use this file except in compliance with the License. A copy of the License is
# located at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# or in the "license" file accompanying this file. This file is distributed on
# an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
# express or implied. See the License for the specific language governing
# permissions and limitations under the License.

which protoc
PROTOC_EXISTS=$?
if [ $PROTOC_EXISTS -eq 0 ]; then
    echo "Protoc already installed"
	PROTOC_VERSION=`protoc --version`
	if [ "$PROTOC_VERSION" == "libprotoc 3.5.1" ]; then
		exit 0
	fi
	echo "libprotoc 3.5.1 required, but found: $PROTOC_VERSION"
	exit 1
fi

if [ "$(uname)" == "Darwin" ]; then
    brew install protobuf
elif [ `whoami` == "root" ]; then
    mkdir -p /usr/local/src/protoc
    pushd /usr/local/src/protoc
    wget https://github.com/google/protobuf/releases/download/v3.5.1/protoc-3.5.1-linux-x86_64.zip -O /usr/local/src/protoc-3.5.1-linux-x86_64.zip
    unzip -x ../protoc-3.5.1-linux-x86_64.zip
    if [ ! -e /usr/local/bin/protoc ]; then
        ln -s `pwd`/bin/protoc /usr/local/bin/protoc
    fi
    popd
elif [ "$(expr substr $(uname -s) 1 5)" == "Linux" ]; then
    echo "Project setup needs sudo to put protoc in /usr/local/src, so it will ask a few times"
    sudo chmod a+w /usr/local/src
    mkdir -p /usr/local/src/protoc
    pushd /usr/local/src/protoc
    wget https://github.com/google/protobuf/releases/download/v3.5.1/protoc-3.5.1-linux-x86_64.zip -O /usr/local/src/protoc-3.5.1-linux-x86_64.zip
    unzip -x ../protoc-3.5.1-linux-x86_64.zip
    if [ ! -e /usr/local/bin/protoc ]; then
        sudo ln -s `pwd`/bin/protoc /usr/local/bin/protoc
    fi
    popd
fi
exit 0
