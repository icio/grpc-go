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

// Package grpcsync implements additional synchronization primitives built upon
// the sync package.
package grpcsync

import (
	"sync"
)

// Event represents a one-time event that may occur in the future.
type Event struct {
	mu     sync.Mutex
	c      chan struct{}
	onFire []func()
}

// Fire causes e to complete and all callbacks registered via OnFire to be
// called synchronously.  It is safe to call multiple times, and concurrently.
// It returns true iff this call to Fire caused the signaling channel returned
// by Done to close.
func (e *Event) Fire() bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.HasFired() {
		return false
	}
	for _, of := range e.onFire {
		of()
	}
	close(e.c)
	return true
}

// OnFire registers f to be called when e fires.  If e has already fired, calls
// f synchronously.
func (e *Event) OnFire(f func()) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.HasFired() {
		f()
		return
	}
	e.onFire = append(e.onFire, f)
}

// DoThenFire calls f if e has not yet fired, and then causes the event to be
// fired by this call, and returns true.  Other synchronous callers of Do or
// Fire will block and return false when the event is fired.  If e has already
// fired, does nothing and returns false.
func (e *Event) DoThenFire(f func()) bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.HasFired() {
		return false
	}
	f()
	for _, of := range e.onFire {
		of()
	}
	close(e.c)
	return true
}

// Done returns a channel that will be closed when Fire is called.
func (e *Event) Done() <-chan struct{} {
	return e.c
}

// HasFired returns true if Fire has been called.
func (e *Event) HasFired() bool {
	select {
	case <-e.c:
		return true
	default:
		return false
	}
}

// NewEvent returns a new, ready-to-use Event.
func NewEvent() *Event {
	return &Event{c: make(chan struct{})}
}
