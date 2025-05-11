// Copyright 2024 Cosmos Nicolaou. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package protocol_test

import (
	"context"
	"errors"
	"strings"
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
	mgr := &streamconn.SessionManager{}
	s := mgr.New(mock, idle)

	if _, err := mock.Send(ctx, []byte("login: ")); err != nil {
		t.Fatal(err)
	}

	err := protocol.QSLogin(ctx, s, "admin", "password")
	if err != nil {
		t.Fatal(err)
	}

	s.Release()
	s = mgr.New(mock, idle)

	if _, err := mock.Send(ctx, []byte("login: ")); err != nil {
		t.Fatal(err)
	}

	err = protocol.QSLogin(ctx, s, "admin", "bad-password")

	if !errors.Is(err, protocol.ErrQSLogin) {
		t.Fatal(err)
	}
}

func TestParseResponse(t *testing.T) {

	pr := func(i int, c, r string) string {
		t.Helper()
		line, err := protocol.ParseResponse([]byte(c), []byte(r))
		if err != nil {
			t.Fatalf("unexpected error for case %v: %q: %v", i, c, err)
		}
		return line
	}

	prompt := "QNET> "
	withPrompt := func(s string) string {
		if strings.HasSuffix(s, "\r\n") {
			return s + prompt
		}
		return s + "\r\n" + prompt
	}
	extraResponses := "~OUTPUT,450,29,6\r\n~OUTPUT,450,30,1,100.00\r\n"
	for i, tc := range []struct {
		cmd  string
		resp string
		want string
	}{
		{"~SYSTEM,5,", "~SYSTEM,5,-8:00\r\n", "-8:00"},
		{"~SYSTEM,5,", "~SYSTEM,5,-8:00\r\n", "-8:00"},
		{"~SYSTEM,1,", "~SYSTEM,1,18:33:16\r\n", "18:33:16"},
		{"~SYSTEM,2,", "~SYSTEM,2,11/17/2024\r\n", "11/17/2024"},
		{"~SYSTEM,5,", "~SYSTEM,5,-8:00", "-8:00"},
		{"~SYSTEM,1,", "~SYSTEM,1,18:33:16", "18:33:16"},
		{"~SYSTEM,2,", "~SYSTEM,2,11/17/2024", "11/17/2024"},
		{"~SYSTEM,5,", "~SYSTEM,5,-8:00\r", "-8:00"},
		{"~SYSTEM,1,", "~SYSTEM,1,18:33:16\r", "18:33:16"},
		{"~SYSTEM,2,", "~SYSTEM,2,11/17/2024\r", "11/17/2024"},
	} {
		if got, want := pr(i, tc.cmd, withPrompt(tc.resp)), tc.want; got != want {
			t.Errorf("%v: %v: got %v, want %v", i, tc.cmd, got, want)
		}
		if got, want := pr(i, tc.cmd, withPrompt(extraResponses+tc.resp)), tc.want; got != want {
			t.Errorf("%v: %v: got %v, want %v", i, tc.cmd, got, want)
		}
		if got, want := pr(i, tc.cmd, withPrompt(extraResponses+tc.resp+"\r\n"+extraResponses)), tc.want; got != want {
			t.Errorf("%v: %v: got %v, want %v", i, tc.cmd, got, want)
		}
		if got, want := pr(i, tc.cmd, withPrompt(extraResponses+tc.resp+"\r\n"+extraResponses)), tc.want; got != want {
			t.Errorf("%v: %v: got %v, want %v", i, tc.cmd, got, want)
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
