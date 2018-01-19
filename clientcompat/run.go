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

package main

import (
	"bytes"
	"log"
	"os/exec"

	"github.com/gogo/protobuf/proto"
	"github.com/pkg/errors"
	"github.com/twitchtv/twirp/clientcompat/internal/clientcompat"
)

func runClient(clientBin string, msg *clientcompat.ClientCompatMessage) (resp []byte, errCode string, err error) {
	cmd := exec.Command(clientBin)

	msgBytes, err := proto.Marshal(msg)
	if err != nil {
		return nil, "", errors.Wrap(err, "unable to marshal ClientCompatMessage message")
	}
	cmd.Stdin = bytes.NewReader(msgBytes)

	stdout, stderr := bytes.NewBuffer(nil), bytes.NewBuffer(nil)
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	err = cmd.Run()
	if err != nil {
		err = errors.Wrap(err, "error running client binary")
		log.Printf("client stdout: %s", stdout.String())
		log.Printf("client stderr: %s", stderr.String())
		return nil, "", err
	}

	if stdout.Len() > 0 && stderr.Len() > 0 {
		return nil, "", errors.Errorf("client bin should write to either stdout or stderr, but never both in one invocation")
	}
	if stderr.Len() > 0 {
		return nil, stderr.String(), err
	}

	return stdout.Bytes(), "", nil
}

func runClientNoop(serverURL string, clientBin string) (resp *clientcompat.Empty, twirpErrCode string, err error) {
	req := &clientcompat.Empty{}
	reqBytes, err := proto.Marshal(req)
	if err != nil {
		return nil, "", errors.Wrap(err, "unable to marshal Empty message")
	}
	msg := &clientcompat.ClientCompatMessage{
		ServiceAddress: serverURL,
		Method:         clientcompat.ClientCompatMessage_NOOP,
		Request:        reqBytes,
	}

	respBytes, code, err := runClient(clientBin, msg)
	if err != nil {
		return nil, "", err
	}

	if respBytes != nil {
		resp = new(clientcompat.Empty)
		err = proto.Unmarshal(respBytes, resp)
		if err != nil {
			return nil, "", errors.Wrap(err, "unable to unmarshal stdout from client bin as an Empty response")
		}
	}
	return resp, code, nil
}

func runClientMethod(serverURL string, clientBin string, req *clientcompat.Req) (resp *clientcompat.Resp, twirpErrCode string, err error) {
	reqBytes, err := proto.Marshal(req)
	if err != nil {
		return nil, "", errors.Wrap(err, "unable to marshal Req")
	}
	msg := &clientcompat.ClientCompatMessage{
		ServiceAddress: serverURL,
		Method:         clientcompat.ClientCompatMessage_METHOD,
		Request:        reqBytes,
	}

	respBytes, code, err := runClient(clientBin, msg)
	if err != nil {
		return nil, "", err
	}

	if respBytes != nil {
		resp = new(clientcompat.Resp)
		err = proto.Unmarshal(respBytes, resp)
		if err != nil {
			return nil, "", errors.Wrap(err, "unable to unmarshal stdout from client bin as a Resp")
		}
	}

	return resp, code, nil
}
