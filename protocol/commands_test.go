// Copyright 2024 Cosmos Nicolaou. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package protocol

import (
	"bytes"
	"testing"
)

func TestCommand(t *testing.T) {
	c := NewCommand(SystemCommands, false, []byte("1,23"))
	if got, want := c.request(), []byte("?SYSTEM,1,23\r\n"); !bytes.Equal(got, want) {
		t.Errorf("got %s, want %s", got, want)
	}
	if got, want := c.responsePrefix(), []byte("~SYSTEM,1,23,"); !bytes.Equal(got, want) {
		t.Errorf("got %s, want %s", got, want)
	}
}
