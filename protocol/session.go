// Copyright 2024 Cosmos Nicolaou. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package protocol

import (
	"context"
	"sync"
)

type Transport interface {
	Send(ctx context.Context, buf []byte) (int, error)
	ReadUntil(ctx context.Context, expected []string) ([]byte, error)
	Close(ctx context.Context) error
}

type Session interface {
	Send(ctx context.Context, buf []byte)
	ReadUntil(ctx context.Context, expected ...string) []byte
	Close(ctx context.Context) error
	Err() error
}

type session struct {
	mu   sync.Mutex
	err  error
	conn Transport
	idle *IdleTimer
}

func NewSession(t Transport, idle *IdleTimer) Session {
	return &session{conn: t, idle: idle}
}

func (s *session) Err() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.err
}

func (s *session) Send(ctx context.Context, buf []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.err != nil {
		return
	}
	s.idle.Reset()
	_, s.err = s.conn.Send(ctx, buf)
}

func (s *session) ReadUntil(ctx context.Context, expected ...string) []byte {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.err != nil {
		return nil
	}
	s.idle.Reset()
	out, err := s.conn.ReadUntil(ctx, expected)
	if err != nil {
		s.err = err
		return nil
	}
	return out
}

func (s *session) Close(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.err != nil {
		return s.err
	}
	return s.conn.Close(ctx)
}

type errorSession struct {
	err error
}

// NewErrorSession returns a session that always returns the given error.
func NewErrorSession(err error) Session {
	return &errorSession{err: err}
}

func (s *errorSession) Err() error {
	return s.err
}

func (s *errorSession) Send(ctx context.Context, buf []byte) {
}

func (s *errorSession) ReadUntil(ctx context.Context, expected ...string) []byte {
	return nil
}

func (s *errorSession) Close(ctx context.Context) error {
	return s.err
}
