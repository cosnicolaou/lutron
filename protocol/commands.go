// Copyright 2024 Cosmos Nicolaou. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package protocol

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/cosnicolaou/automation/net/streamconn"
)

var (
	ErrorNullParsedResponse = errors.New("empty response after parsing for command, prompt and errors")

	ErrUnknownCommand = errors.New("unknown command")

	// See https://assets.lutron.com/a/documents/040249.pdf, page 13
	ErrAccessPointParameterCount      = errors.New("access point paremeter count mismatch")
	ErrAccessPointObjectDoesNotExist  = errors.New("access point object does not exist")
	ErrAccessPointInvalidActionNumber = errors.New("access point invalid action number")
	ErrAccessPointParemeterOutOfRange = errors.New("access point parameter out of range")
	ErrAccessPointParamaterMalformed  = errors.New("access point parameter malformed")
	ErrAccessPointUnsupportedCommand  = errors.New("access point unsupported command")
)

// ParseError parses an error message from the Lutron system of
// the form: ~ERROR,Error Number
func ParseError(s string) error {
	var n int
	_, err := fmt.Sscanf(s, "~ERROR,%d", &n)
	if err != nil {
		return fmt.Errorf("error parsing error message: %w", err)
	}
	switch n {
	case 1:
		return ErrAccessPointParameterCount
	case 2:
		return ErrAccessPointObjectDoesNotExist
	case 3:
		return ErrAccessPointInvalidActionNumber
	case 4:
		return ErrAccessPointParemeterOutOfRange
	case 5:
		return ErrAccessPointParamaterMalformed
	case 6:
		return ErrAccessPointUnsupportedCommand
	}
	return fmt.Errorf("unknown error: %q: %d", s, n)
}

func parseResponseLine(cmd, response []byte) (string, error) {
	line := bytes.TrimPrefix(response, cmd)
	if bytes.Contains(line, []byte("bad login")) {
		return "", ErrQSLogin
	}
	if bytes.Contains(line, []byte("unknown command")) {
		return "", ErrUnknownCommand
	}
	if bytes.Contains(line, []byte("~ERROR")) {
		return "", ParseError(string(line))
	}
	return string(line), nil
}

// ParseResponse parses a possibly multi-line response from the Lutron system
// looking for a response to the issued command. There may be multiple other
// responses due to monitoring outpot from the system.
func ParseResponse(cmd, _, response []byte) (string, error) {
	var line []byte
	for _, b := range response {
		if b == 0x00 { // the QS responses sometimes include leading null byte
			continue
		}
		if b == '\r' || b == '\n' {
			if !bytes.HasPrefix(line, cmd) {
				// Unrelated messages, most likely monitoring notifications.
				line = line[:0]
				continue
			}
			return parseResponseLine(cmd, line)
		}
		line = append(line, b)
	}
	if len(line) > 0 && bytes.HasPrefix(line, cmd) {
		return parseResponseLine(cmd, line)
	}
	return "", ErrorNullParsedResponse
}

type CommandGroup int

const (
	SystemCommands CommandGroup = iota
	DeviceCommands
	OutputCommands
	MonitorCommands
	ShadeGroupCommands
)

type Command struct {
	storage [128]byte
	req     []byte
	idx     int
}

func (cg CommandGroup) appendTo(b []byte) []byte {
	switch cg {
	case SystemCommands:
		return append(b, "SYSTEM"...)
	case DeviceCommands:
		return append(b, "DEVICE"...)
	case OutputCommands:
		return append(b, "OUTPUT"...)
	case MonitorCommands:
		return append(b, "MONITOR"...)
	case ShadeGroupCommands:
		return append(b, "SHADEGRP"...)
	}
	return b
}

func NewCommand(grp CommandGroup, set bool, parameters []byte) Command {
	c := Command{}
	if set {
		c.storage[0] = '#'
	} else {
		c.storage[0] = '?'
	}
	c.req = c.storage[:1]
	c.req = grp.appendTo(c.req)
	c.idx = len(c.req)
	if len(parameters) > 0 {
		c.req = append(c.req, ',')
		c.req = append(c.req, parameters...)
	}
	c.req = append(c.req, '\r', '\n')
	return c
}

func (c Command) request() []byte {
	return c.req
}

// The protocol response always includes the original command as a prefix.
func (c Command) responsePrefix() []byte {
	c.req[0] = '~'
	c.req[len(c.req)-2] = ','
	return c.req[:len(c.req)-1]
}

// Call sends the command to the Lutron system, waits for a prompt
// and returns the response.
func (c Command) Call(ctx context.Context, s streamconn.Session) (string, error) {
	s.Send(ctx, c.request())
	response := s.ReadUntil(ctx, "QNET> ")
	if err := s.Err(); err != nil {
		return "", err
	}
	return ParseResponse(c.responsePrefix(), qsPrompt, response)
}
