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

package transport

// OutgoingMessage blah
type OutgoingMessage interface {
	// Marshal marshals to a byte buffer and returns information about the
	// encoding, or an error.  Repeated calls to Marshal must always return the
	// same values.
	//
	// TODO:
	// - MarshalTo([]byte) or MarshalTo([][]byte) for memory reuse?
	// - MarshalTo(io.Writer)? (is that an extra copy?)
	Marshal() ([]byte, *MessageInfo, error) // TODO: [][]byte??? any value?
	// RawMessage provides access to the backing message.  If nil, the message
	// is not available, but Marshal must still provide the correct encoding.
	RawMessage() interface{}
}

// IncomingMessage blah
type IncomingMessage interface {
	// Unmarshal unmarshals from the scatter-gather byte buffer given.
	Unmarshal([][]byte, *MessageInfo) error
	// RawMessage provides access to the backing message.
	RawMessage() interface{}
}

// MessageInfo blah
type MessageInfo struct {
	// Encoding is the message's content-type encoding.
	Encoding string
	// Compressor is the compressor's name or the empty string if compression
	// is disabled.
	Compressor string
}
