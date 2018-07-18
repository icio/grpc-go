/*
 *
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

package server

import (
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// TransportListener blah
type TransportListener interface {
	// Accept blocks until a new Stream is created.  grpc calls this until
	// error != nil.

	Accept() (Stream, error)

	// GracefulClose causes the TransportListener to stop accepting new
	// incoming streams, and returns when all outstanding streams have
	// completed.
	GracefulClose()

	// Close immediately closes all outstanding connections and streams.
	Close()
}

// Header blah
type Header struct {
	MethodName string
}

// Trailer blah
type Trailer struct {
	Status   status.Status
	Metadata metadata.MD
}

// Stream blah
type Stream interface {
	RecvHeader() Header
	RecvMsg() [][]byte
	SetHeader(Header)
	SendHeader(Header)
	SendMsg([][]byte)
	Close(Trailer)

	// Info returns information about the stream's current state.
	Info() StreamInfo
}

// StreamInfo contains information about the stream's current state.  All
// information is optional.
type StreamInfo struct {
	// RemoteAddr is the address of the client (typically an IP/port).
	RemoteAddr string
}
