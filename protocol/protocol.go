// Copyright 2024 Cosmos Nicolaou. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package protocol

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

var (

	// Failed to login.
	ErrQSLogin = errors.New("login failed")

	// See https://assets.lutron.com/a/documents/040249.pdf, page 13
	ErrAccessPointParameterCount      = errors.New("access point paremeter count mismatch")
	ErrAccessPointObjectDoesNotExist  = errors.New("access point object does not exist")
	ErrAccessPointInvalidActionNumber = errors.New("access point invalid action number")
	ErrAccessPointParemeterOutOfRange = errors.New("access point parameter out of range")
	ErrAccessPointParamaterMalformed  = errors.New("access point parameter malformed")
	ErrAccessPointUnsupportedCommand  = errors.New("access point unsupported command")
)

func QSLogin(s Session, user, pass string) error {
	s.Expect("login: ")
	s.Send(user + "\r\n")
	s.Expect("password: ")
	s.Send(pass + "\r\n")
	prompt := s.ReadUntil("QNET> ", "login:")
	if err := s.Err(); err != nil {
		return err
	}
	fmt.Printf("prompt: %q\n", prompt)
	if !strings.Contains(prompt, "QNET>") {
		return ErrQSLogin
	}
	return nil
}

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

func ParseResponseForError(s string) error {
	if strings.Contains(s, "bad login") {
		return ErrQSLogin
	}
	if strings.Contains(s, "~ERROR") {
		return ParseError(s)
	}
	return nil
}

// #DEVICE,2,4,3<CR><LF>
// Integration ID, Component Number, Action Number
// See https://assets.lutron.com/a/documents/040249.pdf
func Device(s Session, id, component, action int) error {
	cmd := fmt.Sprintf("#DEVICE,%d,%d,%d\r\n", id, component, action)
	s.Send(cmd)
	response := s.ReadUntil("QNET> ")
	if err := s.Err(); err != nil {
		return err
	}
	return ParseResponseForError(response)
}

// #OUTPUT,1,1,75,01:30<CR><LF>
// Integration ID, Action Number, Level, Fade Time (seconds)
// See https://assets.lutron.com/a/documents/040249.pdf
func Output(s Session, id, component, level int, time time.Duration) {
}

// ?OUTPUT,3,1<CR><LF>
// Integration ID, Action
// See https://assets.lutron.com/a/documents/040249.pdf
func Query(s Session, id, action int) {
}

// ~OUTPUT,3,1,90.00<CR><LF>
// Integration ID, Action, Level
// See https://assets.lutron.com/a/documents/040249.pdf
func Monitor(s Session, id, action, level int) {
}

type SystemActions int

const (
	SystemTime SystemActions = iota + 1
	SystemDate
	_
	SystemLatLong
	SystemTimeZone
	SystemSunset
	SystemSunrise
	SystemOSRev
)

func System(s Session, set bool, action SystemActions, parameters ...string) (string, error) {
	prefix := "?"
	if set {
		prefix = "#"
	}
	cmd := fmt.Sprintf("%vSYSTEM,%d", prefix, action)
	s.Send(cmd + "\r\n")
	response := s.ReadUntil("QNET> ")
	if err := s.Err(); err != nil {
		return "", err
	}
	return strings.TrimPrefix(response, cmd), ParseResponseForError(response)
}
