// Copyright 2018 Twitch Interactive, Inc.  All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may not
// use this file except in compliance with the License. A copy of the License is
// located at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// or in the "license" file accompanying this file. This file is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

// Package twirp provides core types used in generated Twirp servers and client.
//
// Twirp services handle errors using the `twirp.Error` interface.
//
// For example, a server method may return an InvalidArgumentError:
//
//     if req.Order != "DESC" && req.Order != "ASC" {
//         return nil, twirp.InvalidArgumentError("Order", "must be DESC or ASC")
//     }
//
// And the same twirp.Error is returned by the client, for example:
//
//     resp, err := twirpClient.RPCMethod(ctx, req)
//     if err != nil {
//         if twerr, ok := err.(twirp.Error); ok {
//             switch twerr.Code() {
//             case twirp.InvalidArgument:
//                 log.Error("invalid argument "+twirp.Meta("argument"))
//             default:
//                 log.Error(twerr.Error())
//             }
//         }
//     }
//
// Clients may also return Internal errors if something failed on the system:
// the server, the network, or the client itself (i.e. failure parsing
// response).
//
package twirp

// TwirpPackageIsVersion7 is a constant referenced from generated code to cause
// a readable compile error if the generated code requires this version.
const TwirpPackageIsVersion7 = true
