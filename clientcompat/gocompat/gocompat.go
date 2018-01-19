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
	"context"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/golang/protobuf/proto"
	"github.com/twitchtv/twirp"
	"github.com/twitchtv/twirp/clientcompat/internal/clientcompat"
)

func main() {
	var in clientcompat.ClientCompatMessage
	inBytes, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatalf("read stdin err: %v", err)
	}
	err = proto.Unmarshal(inBytes, &in)
	if err != nil {
		log.Fatalf("unmarshal err: %v", err)
	}

	client := clientcompat.NewCompatServiceProtobufClient(in.ServiceAddress, http.DefaultClient)

	switch in.Method {
	case clientcompat.ClientCompatMessage_NOOP:
		if err := doNoop(client, in.Request); err != nil {
			log.Fatalf("doNoop err: %v", err)
		}
	case clientcompat.ClientCompatMessage_METHOD:
		if err := doMethod(client, in.Request); err != nil {
			log.Fatalf("doMethod err: %v", err)
		}
	default:
		log.Fatalf("unexpected method: %v", in.Method)
	}
}

func doNoop(client clientcompat.CompatService, req []byte) error {
	var e clientcompat.Empty
	err := proto.Unmarshal(req, &e)
	if err != nil {
		return err
	}
	resp, err := client.NoopMethod(context.Background(), &e)
	if err != nil {
		errCode := err.(twirp.Error).Code()
		os.Stderr.Write([]byte(errCode))
	} else {
		respBytes, err := proto.Marshal(resp)
		if err != nil {
			return err
		}
		os.Stdout.Write(respBytes)
	}
	return nil
}

func doMethod(client clientcompat.CompatService, req []byte) error {
	var r clientcompat.Req
	err := proto.Unmarshal(req, &r)
	if err != nil {
		return err
	}
	resp, err := client.Method(context.Background(), &r)
	if err != nil {
		errCode := err.(twirp.Error).Code()
		os.Stderr.Write([]byte(errCode))
	} else {
		respBytes, err := proto.Marshal(resp)
		if err != nil {
			return err
		}
		os.Stdout.Write(respBytes)
	}
	return nil
}
