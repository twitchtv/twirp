#!/usr/bin/env python
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
