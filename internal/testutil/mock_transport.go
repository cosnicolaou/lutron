// Copyright 2024 Cosmos Nicolaou. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package testutil

import (
	"context"
	"fmt"
	"io"
	"strings"
)

type MockTransport struct {
	lookup  map[string]string
	connCh  chan byte
	verbose bool
}

func NewMockTransport(verbose bool) *MockTransport {
	return &MockTransport{
		lookup:  make(map[string]string),
		connCh:  make(chan byte, 1000),
		verbose: verbose,
	}
}

func (m *MockTransport) SetResponse(input, response string) {
	m.lookup[input] = response
}

func (m *MockTransport) log(format string, args ...interface{}) {
	if m.verbose {
		fmt.Printf(format, args...)
	}
}

func (m *MockTransport) Send(_ context.Context, buf []byte) (int, error) {
	response := m.lookup[string(buf)]
	m.log("sending: %q in response to %q\n", response, buf)
	for _, r := range []byte(response) {
		m.connCh <- r
	}
	return len(response), nil
}

func (m *MockTransport) SendSensitive(_ context.Context, buf []byte) (int, error) {
	response := m.lookup[string(buf)]
	m.log("sending: %q in response to %q\n", response, buf)
	for _, r := range []byte(response) {
		m.connCh <- r
	}
	return len(response), nil
}

func (m *MockTransport) ReadUntil(ctx context.Context, expected []string) ([]byte, error) {
	seen := []byte{}
	m.log("reading until %v\n", expected)
	for r := range m.connCh {
		seen = append(seen, r)
		m.log("% 30q\n", seen)
		for _, t := range expected {
			if strings.HasSuffix(string(seen), t) {
				m.log("read: %q\n", t)
				return seen, nil
			}
		}
	}
	return nil, io.EOF
}

func (m *MockTransport) Close(context.Context) error {
	return nil
}
