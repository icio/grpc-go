/*
 * Copyright 2018 gRPC authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

// Package jsonpb implements a json-proto codec for gRPC.
package jsonpb

import (
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc/encoding"
)

// Name is the name registered for the proto compressor.
const Name = "jsonpb"

func init() {
	encoding.RegisterCodec(&codec{&jsonpb.Marshaler{}})
}

// codec is a Codec implementation with protobuf. It is the default codec for gRPC.
type codec struct {
	m *jsonpb.Marshaler
}

func (c *codec) Marshal(v interface{}) ([]byte, error) {
	protoMsg := v.(proto.Message)
	s, err := c.m.MarshalToString(protoMsg)
	return []byte(s), err
}

func (c *codec) Unmarshal(data []byte, v interface{}) error {
	protoMsg := v.(proto.Message)
	return jsonpb.UnmarshalString(string(data), protoMsg)
}

func (*codec) Name() string {
	return Name
}
