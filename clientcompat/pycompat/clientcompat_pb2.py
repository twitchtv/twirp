# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: clientcompat.proto

import sys
_b=sys.version_info[0]<3 and (lambda x:x) or (lambda x:x.encode('latin1'))
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from google.protobuf import reflection as _reflection
from google.protobuf import symbol_database as _symbol_database
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()




DESCRIPTOR = _descriptor.FileDescriptor(
  name='clientcompat.proto',
  package='twirp.clientcompat',
  syntax='proto3',
  serialized_options=_b('Z\014clientcompat'),
  serialized_pb=_b('\n\x12\x63lientcompat.proto\x12\x12twirp.clientcompat\"\x07\n\x05\x45mpty\"\x10\n\x03Req\x12\t\n\x01v\x18\x01 \x01(\t\"\x11\n\x04Resp\x12\t\n\x01v\x18\x01 \x01(\x05\"\xb9\x01\n\x13\x43lientCompatMessage\x12\x17\n\x0fservice_address\x18\x01 \x01(\t\x12K\n\x06method\x18\x02 \x01(\x0e\x32;.twirp.clientcompat.ClientCompatMessage.CompatServiceMethod\x12\x0f\n\x07request\x18\x03 \x01(\x0c\"+\n\x13\x43ompatServiceMethod\x12\x08\n\x04NOOP\x10\x00\x12\n\n\x06METHOD\x10\x01\x32\x90\x01\n\rCompatService\x12;\n\x06Method\x12\x17.twirp.clientcompat.Req\x1a\x18.twirp.clientcompat.Resp\x12\x42\n\nNoopMethod\x12\x19.twirp.clientcompat.Empty\x1a\x19.twirp.clientcompat.EmptyB\x0eZ\x0c\x63lientcompatb\x06proto3')
)



_CLIENTCOMPATMESSAGE_COMPATSERVICEMETHOD = _descriptor.EnumDescriptor(
  name='CompatServiceMethod',
  full_name='twirp.clientcompat.ClientCompatMessage.CompatServiceMethod',
  filename=None,
  file=DESCRIPTOR,
  values=[
    _descriptor.EnumValueDescriptor(
      name='NOOP', index=0, number=0,
      serialized_options=None,
      type=None),
    _descriptor.EnumValueDescriptor(
      name='METHOD', index=1, number=1,
      serialized_options=None,
      type=None),
  ],
  containing_type=None,
  serialized_options=None,
  serialized_start=231,
  serialized_end=274,
)
_sym_db.RegisterEnumDescriptor(_CLIENTCOMPATMESSAGE_COMPATSERVICEMETHOD)


_EMPTY = _descriptor.Descriptor(
  name='Empty',
  full_name='twirp.clientcompat.Empty',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  serialized_options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=42,
  serialized_end=49,
)


_REQ = _descriptor.Descriptor(
  name='Req',
  full_name='twirp.clientcompat.Req',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='v', full_name='twirp.clientcompat.Req.v', index=0,
      number=1, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  serialized_options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=51,
  serialized_end=67,
)


_RESP = _descriptor.Descriptor(
  name='Resp',
  full_name='twirp.clientcompat.Resp',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='v', full_name='twirp.clientcompat.Resp.v', index=0,
      number=1, type=5, cpp_type=1, label=1,
      has_default_value=False, default_value=0,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  serialized_options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=69,
  serialized_end=86,
)


_CLIENTCOMPATMESSAGE = _descriptor.Descriptor(
  name='ClientCompatMessage',
  full_name='twirp.clientcompat.ClientCompatMessage',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='service_address', full_name='twirp.clientcompat.ClientCompatMessage.service_address', index=0,
      number=1, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='method', full_name='twirp.clientcompat.ClientCompatMessage.method', index=1,
      number=2, type=14, cpp_type=8, label=1,
      has_default_value=False, default_value=0,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='request', full_name='twirp.clientcompat.ClientCompatMessage.request', index=2,
      number=3, type=12, cpp_type=9, label=1,
      has_default_value=False, default_value=_b(""),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
    _CLIENTCOMPATMESSAGE_COMPATSERVICEMETHOD,
  ],
  serialized_options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=89,
  serialized_end=274,
)

_CLIENTCOMPATMESSAGE.fields_by_name['method'].enum_type = _CLIENTCOMPATMESSAGE_COMPATSERVICEMETHOD
_CLIENTCOMPATMESSAGE_COMPATSERVICEMETHOD.containing_type = _CLIENTCOMPATMESSAGE
DESCRIPTOR.message_types_by_name['Empty'] = _EMPTY
DESCRIPTOR.message_types_by_name['Req'] = _REQ
DESCRIPTOR.message_types_by_name['Resp'] = _RESP
DESCRIPTOR.message_types_by_name['ClientCompatMessage'] = _CLIENTCOMPATMESSAGE
_sym_db.RegisterFileDescriptor(DESCRIPTOR)

Empty = _reflection.GeneratedProtocolMessageType('Empty', (_message.Message,), dict(
  DESCRIPTOR = _EMPTY,
  __module__ = 'clientcompat_pb2'
  # @@protoc_insertion_point(class_scope:twirp.clientcompat.Empty)
  ))
_sym_db.RegisterMessage(Empty)

Req = _reflection.GeneratedProtocolMessageType('Req', (_message.Message,), dict(
  DESCRIPTOR = _REQ,
  __module__ = 'clientcompat_pb2'
  # @@protoc_insertion_point(class_scope:twirp.clientcompat.Req)
  ))
_sym_db.RegisterMessage(Req)

Resp = _reflection.GeneratedProtocolMessageType('Resp', (_message.Message,), dict(
  DESCRIPTOR = _RESP,
  __module__ = 'clientcompat_pb2'
  # @@protoc_insertion_point(class_scope:twirp.clientcompat.Resp)
  ))
_sym_db.RegisterMessage(Resp)

ClientCompatMessage = _reflection.GeneratedProtocolMessageType('ClientCompatMessage', (_message.Message,), dict(
  DESCRIPTOR = _CLIENTCOMPATMESSAGE,
  __module__ = 'clientcompat_pb2'
  # @@protoc_insertion_point(class_scope:twirp.clientcompat.ClientCompatMessage)
  ))
_sym_db.RegisterMessage(ClientCompatMessage)


DESCRIPTOR._options = None

_COMPATSERVICE = _descriptor.ServiceDescriptor(
  name='CompatService',
  full_name='twirp.clientcompat.CompatService',
  file=DESCRIPTOR,
  index=0,
  serialized_options=None,
  serialized_start=277,
  serialized_end=421,
  methods=[
  _descriptor.MethodDescriptor(
    name='Method',
    full_name='twirp.clientcompat.CompatService.Method',
    index=0,
    containing_service=None,
    input_type=_REQ,
    output_type=_RESP,
    serialized_options=None,
  ),
  _descriptor.MethodDescriptor(
    name='NoopMethod',
    full_name='twirp.clientcompat.CompatService.NoopMethod',
    index=1,
    containing_service=None,
    input_type=_EMPTY,
    output_type=_EMPTY,
    serialized_options=None,
  ),
])
_sym_db.RegisterServiceDescriptor(_COMPATSERVICE)

DESCRIPTOR.services_by_name['CompatService'] = _COMPATSERVICE

# @@protoc_insertion_point(module_scope)
