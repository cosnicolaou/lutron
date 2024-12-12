// Copyright 2024 Cosmos Nicolaou. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package protocol_test

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/cosnicolaou/automation/net/netutil"
	"github.com/cosnicolaou/automation/net/streamconn"
	"github.com/cosnicolaou/lutron/internal/testutil"
	"github.com/cosnicolaou/lutron/protocol"
)

func TestLogin(t *testing.T) {
	ctx := context.Background()

	mock := testutil.NewMockTransport(testing.Verbose())
	mock.SetResponse("login: ", "login: ")
	mock.SetResponse("admin\r\n", "password: ")
	mock.SetResponse("password\r\n", "\r\nQNET> ")
	mock.SetResponse("bad-password\r\n", "\r\nbad login\r\nlogin:\r\n")

	idle := netutil.NewIdleTimer(10)
	s := streamconn.NewSession(mock, idle)

	mock.Send(ctx, []byte("login: "))

	err := protocol.QSLogin(ctx, s, "admin", "password")
	if err != nil {
		t.Fatal(err)
	}

	s = streamconn.NewSession(mock, idle)

	mock.Send(ctx, []byte("login: "))

	err = protocol.QSLogin(ctx, s, "admin", "bad-password")

	if !errors.Is(err, protocol.ErrQSLogin) {
		t.Fatal(err)
	}
}

func TestParseResponse(t *testing.T) {

	pr := func(c, p, r string) []string {
		lines, err := protocol.ParseResponse([]byte(c), []byte(p), []byte(r))
		if err != nil {
			t.Fatal(err)
		}
		return lines
	}

	sl := func(s ...string) []string {
		return s
	}

	for _, tc := range []struct {
		cmd  string
		resp string
		want []string
	}{
		{"~SYSTEM,5,", "~SYSTEM,5,-8:00\r\nQNET> ", sl("-8:00")},
		{"~SYSTEM,1,", "~SYSTEM,1,18:33:16\r\nQNET> ", sl("18:33:16")},
		{"~SYSTEM,2,", "~SYSTEM,2,11/17/2024\r\nQNET> ", sl("11/17/2024")},
		{"~SYSTEM,5,", "~SYSTEM,5,-8:00", sl("-8:00")},
		{"~SYSTEM,1,", "~SYSTEM,1,18:33:16", sl("18:33:16")},
		{"~SYSTEM,2,", "~SYSTEM,2,11/17/2024", sl("11/17/2024")},
		{"~SYSTEM,5,", "~SYSTEM,5,-8:00\r", sl("-8:00")},
		{"~SYSTEM,1,", "~SYSTEM,1,18:33:16\r", sl("18:33:16")},
		{"~SYSTEM,2,", "~SYSTEM,2,11/17/2024\r", sl("11/17/2024")},
	} {
		if got, want := pr(tc.cmd, "QNET> ", tc.resp), tc.want; !slices.Equal(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	}
}

func TestParseDateTime(t *testing.T) {
	sysTime := "20:21:47"
	sysDate := "11/17/2024"
	sysTz := "-8:00" // Note this is missing the leading 0 for a 07:00 timezone format
	sysDateTime := sysDate + " " + sysTime + " " + protocol.NormalizeTimeZone(sysTz)
	st, err := time.Parse("01/02/2006 15:04:05 -07:00", sysDateTime)
	if err != nil {
		t.Fatal(err)
	}
	if !st.Equal(time.Date(2024, 11, 17, 20, 21, 47, 0, time.FixedZone("PST", -8*60*60))) {
		t.Errorf("got %v, want %v", st, time.Date(2024, 11, 17, 20, 21, 47, 0, time.FixedZone("PST", -8*60*60)))
	}
}
