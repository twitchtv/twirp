# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: service.proto

import sys
_b=sys.version_info[0]<3 and (lambda x:x) or (lambda x:x.encode('latin1'))
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from google.protobuf import reflection as _reflection
from google.protobuf import symbol_database as _symbol_database
from google.protobuf import descriptor_pb2
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()


from google.protobuf import empty_pb2 as google_dot_protobuf_dot_empty__pb2
from google.protobuf import wrappers_pb2 as google_dot_protobuf_dot_wrappers__pb2


DESCRIPTOR = _descriptor.FileDescriptor(
  name='service.proto',
  package='twirp.twirptest.use_empty',
  syntax='proto3',
  serialized_pb=_b('\n\rservice.proto\x12\x19twirp.twirptest.use_empty\x1a\x1bgoogle/protobuf/empty.proto\x1a\x1egoogle/protobuf/wrappers.proto2C\n\x03Svc\x12<\n\x04Send\x12\x1c.google.protobuf.StringValue\x1a\x16.google.protobuf.EmptyB\x19Z\x17google_protobuf_importsb\x06proto3')
  ,
  dependencies=[google_dot_protobuf_dot_empty__pb2.DESCRIPTOR,google_dot_protobuf_dot_wrappers__pb2.DESCRIPTOR,])



_sym_db.RegisterFileDescriptor(DESCRIPTOR)


DESCRIPTOR.has_options = True
DESCRIPTOR._options = _descriptor._ParseOptions(descriptor_pb2.FileOptions(), _b('Z\027google_protobuf_imports'))

_SVC = _descriptor.ServiceDescriptor(
  name='Svc',
  full_name='twirp.twirptest.use_empty.Svc',
  file=DESCRIPTOR,
  index=0,
  options=None,
  serialized_start=105,
  serialized_end=172,
  methods=[
  _descriptor.MethodDescriptor(
    name='Send',
    full_name='twirp.twirptest.use_empty.Svc.Send',
    index=0,
    containing_service=None,
    input_type=google_dot_protobuf_dot_wrappers__pb2._STRINGVALUE,
    output_type=google_dot_protobuf_dot_empty__pb2._EMPTY,
    options=None,
  ),
])
_sym_db.RegisterServiceDescriptor(_SVC)

DESCRIPTOR.services_by_name['Svc'] = _SVC

# @@protoc_insertion_point(module_scope)