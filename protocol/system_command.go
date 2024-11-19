// Copyright 2024 Cosmos Nicolaou. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package protocol

import (
	"context"
	"strconv"
)

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

// NormalizeTimeZone normalizes a timezone string to the form: (+|-)HH:MM
// given an input of (+|-)H:MM as returned by QS systems.
func NormalizeTimeZone(tz string) string {
	switch len(tz) {
	case 0:
		return "0000"
	case 6:
		return tz
	case 5:
		mm := tz[3:]
		h := tz[1:2]
		return tz[0:1] + "0" + h + ":" + mm
	}
	return tz
}

// System sends a '[#?]System' command to the Lutron system.
func System(ctx context.Context, s Session, set bool, action SystemActions, parameters ...string) (string, error) {
	cmd := NewCommand(SystemCommands, set, []byte(strconv.Itoa(int(action))))
	r, err := cmd.Call(ctx, s)
	if err != nil {
		return "", err
	}
	if len(r) == 0 {
		return "", ErrorNullParsedResponse
	}
	return r[0], nil
}
