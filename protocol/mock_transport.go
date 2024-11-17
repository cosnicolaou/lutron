// Copyright 2024 Cosmos Nicolaou. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package protocol

import (
	"fmt"
	"io"
	"strings"
)

type MockTransport struct {
	lookup  map[string]string
	connCh  chan rune
	verbose bool
}

func NewMockTransport(verbose bool) *MockTransport {
	return &MockTransport{
		lookup:  make(map[string]string),
		connCh:  make(chan rune, 1000),
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

func (m *MockTransport) Send(text string) error {
	response := m.lookup[text]
	m.log("sending: %q in response to %q\n", response, text)
	for _, r := range response {
		m.connCh <- r
	}
	return nil
}

func (m *MockTransport) ReadUntil(text ...string) (string, error) {
	seen := ""
	m.log("reading until %v\n", text)
	for r := range m.connCh {
		seen += string(r)
		m.log("% 30q\n", seen)
		for _, t := range text {
			if strings.HasSuffix(seen, t) {
				m.log("read: %q\n", t)
				return seen, nil
			}
		}
	}
	return "", io.EOF
}

func (m *MockTransport) Expect(text ...string) error {
	_, err := m.ReadUntil(text...)
	return err
}

func (m *MockTransport) Close() error {
	return nil
}
