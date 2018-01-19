#!/usr/bin/env python
#
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

from __future__ import print_function

import sys

import clientcompat_pb2
import clientcompat_pb2_twirp

def main():
    req = read_request()
    client = clientcompat_pb2_twirp.CompatServiceClient(req.service_address)
    try:
        resp = do_request(client, req)
        sys.stdout.write(resp.SerializeToString())
    except clientcompat_pb2_twirp.TwirpException as e:
        sys.stderr.write(e.code)


def read_request():
    input_str = sys.stdin.read()
    return clientcompat_pb2.ClientCompatMessage.FromString(input_str)


def do_request(client, req):
    if req.method == clientcompat_pb2.ClientCompatMessage.NOOP:
        input_type = clientcompat_pb2.Empty
        method = client.noop_method
    elif req.method == clientcompat_pb2.ClientCompatMessage.METHOD:
        input_type = clientcompat_pb2.Req
        method = client.method

    req = input_type.FromString(req.request)
    return method(req)


if __name__ == "__main__":
    main()
