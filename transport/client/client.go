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

package client

import (
	"net"
	"time"

	"golang.org/x/net/context"

	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/resolver"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/transport"
)

// A TransportMonitor is a monitor for client-side transports.
type TransportMonitor interface {
	// OnConnected reports that the Transport is fully connected - i.e. the
	// remote server has confirmed it is a gRPC server.
	//
	// May only be called once.
	OnConnected()
	// OnError reports that the Transport has closed due to the error provided.
	// Existing streams may or may not continue.
	//
	// TODO: idleness should be decided based on this error (?).. maybe if the
	// error implements "GoIdle() bool" then we idle?
	//
	// Once called, no further calls in the TransportMonitor are valid.
	OnError(error)
}

// TransportBuildOptions blah
type TransportBuildOptions struct {
	ByteTransportOptions ByteTransportOptions
	HTTPTransportOptions HTTPTransportOptions
	// Options contains opaque Transport configuration.
	Options []interface{}
	// May contain:
	// TransportCredentials
	// keepalive.ClientParameters (wrapped in a thing to allow it to be modified?)
	// stats.Handler???
	// GRPCHTTP2Options
}

// ByteTransportOptions blah.
type ByteTransportOptions struct {
	// Dialer should be used to dial the remote target
	Dialer func(string, time.Duration) (net.Conn, error) // XXX TODO: use ctx instead of duration?
	// TransportCredentials should be used to secure the connection created by
	// the transport.
	TransportCredentials credentials.TransportCredentials
}

// HTTPTransportOptions blah.
type HTTPTransportOptions struct {
	UserAgent string
	// MaxHeaderListSize sets the max (uncompressed) size of header list that
	// is prepared to be received.
	MaxHeaderListSize uint32
	// InitialWindowSize sets the initial window size for a stream.
	InitialWindowSize int32
	// InitialConnWindowSize sets the initial window size for a connection.
	InitialConnWindowSize int32
}

// TransportBuilder blah
type TransportBuilder interface {
	// Build begins connecting to the address.  It must return a Transport that
	// is ready to accept new streams or an error.
	Build(context.Context, resolver.Address, TransportMonitor, TransportBuildOptions) (Transport, error)
}

// Header blah
type Header struct {
	Method           string      // required: remote server's RPC method
	Authority        string      // for transports supporting virtual hosting
	Metadata         metadata.MD // optional
	PreviousAttempts int
	MaxRecvMsgSize   *int // if non-nil, maximum size for a received message
	Options          []interface{}
}

// HTTPStreamOptions contains http-specific Stream options.
type HTTPStreamOptions struct {
	ContentSubtype string // append this string to the HTTP content-type header
}

// A Transport is a client-side gRPC transport.  It begins in a
// "half-connected" state where the client may opportunistically start new
// streams by calling NewStream.  Some clients will wait until the
// TransportMonitor's Connected method is called.
type Transport interface {
	// NewStream begins a new Stream on the Transport.  Blocks until sufficient
	// stream quota is available, if applicable.  If the Transport is closed,
	// returns an error.
	NewStream2(context.Context, Header) (Stream, error)

	// GracefulClose closes the Transport.  Outstanding and pending Streams
	// created by NewStream continue uninterrupted and this function blocks
	// until the Streams are finished.
	GracefulClose2()

	// Close closes the Transport.  Outstanding and pending Streams created by
	// NewStream are canceled.
	Close2()

	// Info returns information about the transport's current state.
	Info2() TransportInfo
}

// TransportInfo contains information about the transport's current state.  All
// information is optional.
type TransportInfo struct {
	// RemoteAddr is the address of the server (typically an IP/port).
	RemoteAddr net.Addr
	IsSecure   bool // if set, WithInsecure is not required and Per-RPC Credentials are allowed.
	AuthInfo   credentials.AuthInfo
}

// StreamSendMsgOptions blah
type StreamSendMsgOptions struct {
	CloseSend bool          // set if this is the final outgoing message on the stream.
	Options   []interface{} // E.g. whether to compress this message.
}

// Trailer blah
type Trailer struct {
	Status   *status.Status
	Metadata metadata.MD
}

// Stream blah
type Stream interface {
	// SendMsg queues the message m to be sent by the Stream and returns true
	// unless the Stream terminates before sending.
	SendMsg2(m transport.OutgoingMessage, opts StreamSendMsgOptions) bool

	// CloseSend queues a notification to the remote server that the client is
	// done sending and returns true unless the Stream terminates before
	// sending.
	// CloseSend2() bool

	// XXX TODO
	// SendMsg option (optionally delete CloseSend; nil m would indicate no msg)
	// Other options:
	// - SendMsg parameter (...)
	// - Just don't have it and call both.  (How to combine to save that extra data frame for unary RPCs?)
	// - SendLastMsg2(m transport.OutgoingMessage, opts StreamSendMsgOptions) bool
	// - ???

	// RecvHeader blocks until the Stream receives the server's header and then
	// returns it.  Returns nil if the Stream terminated without a valid
	// header.  Repeated calls will return the same header.
	RecvHeader2() *ServerHeader

	// RecvMsg receives the next message on the Stream into m and returns true
	// unless the Stream terminates before a full message is received.
	RecvMsg2(m transport.IncomingMessage) bool

	// RecvTrailer blocks until the Stream receives the server's trailer and
	// then returns it.  Returns a synthesized trailer containing an
	// appropriate status if the RPC terminates before receiving a trailer from
	// the server.  Repeated calls will return the same trailer.
	//
	// If all messages have not be retrieved from the stream before calling
	// RecvTrailer, subsequent calls to RecvMsg should immediately fail.  This
	// is to prevent the status in the trailer from changing as a result of
	// parsing the messages.
	RecvTrailer2() Trailer

	// Cancel unconditionally cancels the RPC.  Queued messages may not be
	// sent.  If the stream does not already have a status, the one provided
	// (which must be non-nil) is used.
	Cancel2(*status.Status)

	// Info returns information about the stream's current state.
	Info2() StreamInfo
}

// ServerHeader contains header data sent by the server.
type ServerHeader struct { /// XXX TODO: name?
	Metadata metadata.MD
}

// StreamInfo contains information about the stream's current state.
//
// All information is optional.  Fields not supported by a Stream returning
// this struct should be nil.  If a transport does not support all features,
// some grpc features (e.g. transparent retry or grpclb load reporting) may not
// work properly.
type StreamInfo struct {
	// BytesReceived is true iff the client received any data for this stream
	// from the server (e.g. partial header bytes), false if the stream ended
	// without receiving any data, or nil if data may still be received.
	BytesReceived *bool

	// Unprocessed is true if the server has confirmed the application did not
	// process this stream*, false if the server sent a response indicating the
	// application may have processed the stream, or nil if it is uncertain.
	//
	// *: In HTTP/2, this is true if a RST_STREAM with REFUSED_STREAM is
	//    received or if a GOAWAY including this stream's ID is received.
	Unprocessed *bool
}
