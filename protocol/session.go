// Copyright 2024 Cosmos Nicolaou. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package protocol

import (
	"iter"
	"sync"
)

type Session interface {
	SetTransport(c Transport)
	Append(text string)
	Entries() iter.Seq[string]
	Close() error
	Err() error
	Send(text string)
	ReadUntil(text string) string
	Expect(text string)
}

type session struct {
	mu      sync.Mutex
	err     error
	entries []string
	conn    Transport
	busy    bool
}

func NewSession() Session {
	return &session{}
}

func (s *session) SetTransport(c Transport) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.conn = c
}

func (s *session) Append(text string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries = append(s.entries, text)
}

func (s *session) Entries() iter.Seq[string] {
	s.mu.Lock()
	defer s.mu.Unlock()
	return func(yield func(s string) bool) {
		for _, e := range s.entries {
			if !yield(e) {
				break
			}
		}
	}
}

func (s *session) Err() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.err
}

func (s *session) Send(text string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.err != nil {
		return
	}
	s.err = s.conn.Send(text)
}

func (s *session) ReadUntil(text string) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.err != nil {
		return ""
	}
	out, err := s.conn.ReadUntil(text)
	if err != nil {
		s.err = err
		return ""
	}
	return out
}

func (s *session) Expect(text string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.err != nil {
		return
	}
	_, err := s.conn.ReadUntil(text)
	if err != nil {
		s.err = err
		return
	}
	return
}

func (s *session) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.err != nil {
		return s.err
	}
	return s.conn.Close()
}
