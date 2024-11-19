// Copyright 2024 Cosmos Nicolaou. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package protocol_test

import (
	"testing"
	"time"

	"github.com/cosnicolaou/lutron/protocol"
)

func TestIdleTime(t *testing.T) {
	timer := protocol.NewIdleTimer(10 * time.Minute)
	if got, want := timer.Expired(), false; got != want {
		t.Errorf("got %v, want %v", got, want)
	}
	time.Sleep(time.Second)
	n1 := timer.Remaining()
	if got, want := timer.Remaining(), 10*time.Minute; got >= want {
		t.Errorf("got %v, want > %v", timer.Remaining(), 10*time.Minute)
	}
	timer.Reset()
	n2 := timer.Remaining()
	if n1 >= n2 {
		t.Errorf("remaining time n2 should be less than n1 %v < %v", n1, n2)
	}

	timer = protocol.NewIdleTimer(10 * time.Millisecond)
	time.Sleep(time.Second)
	if got, want := timer.Expired(), true; got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}
