// Copyright 2024 Cosmos Nicolaou. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package protocol_test

import (
	"errors"
	"testing"

	"github.com/cosnicolaou/lutron/protocol"
)

func TestLogin(t *testing.T) {

	mock := protocol.NewMockTransport(testing.Verbose())
	mock.SetResponse("login:", "login: ")
	mock.SetResponse("admin\r\n", "password:")
	mock.SetResponse("password\r\n", "\r\nQNET> ")
	mock.SetResponse("bad-password\r\n", "\r\nbad login\r\nlogin:\r\n")

	c := protocol.New("login", protocol.WithTransport(mock))

	mock.Send("login:")

	err := c.Run(func(s protocol.Session) error {
		return protocol.QSLogin(s, "admin", "password")
	}).Err()

	if err != nil {
		t.Fatal(err)
	}

	c = protocol.New("login", protocol.WithTransport(mock))

	mock.Send("login:")

	err = c.Run(func(s protocol.Session) error {
		return protocol.QSLogin(s, "admin", "bad-password")
	}).Err()

	if !errors.Is(err, protocol.QSLoginError) {
		t.Fatal(err)
	}
}
