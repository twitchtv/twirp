#!/usr/bin/env python

from distutils.core import setup

setup(
    name='pycompat',
    description='Twirp-Python compatibility test client',
    py_modules=['clientcompat_pb2', 'clientcompat_pb2_twirp'],
    install_requires=['protobuf'],
    scripts=['pycompat.py'],
)
