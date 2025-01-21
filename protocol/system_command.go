// Copyright 2024 Cosmos Nicolaou. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package protocol

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/cosnicolaou/automation/net/streamconn"
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
func System(ctx context.Context, s streamconn.Session, set bool, action SystemActions) (string, error) {
	cmd := NewCommand(SystemCommands, set, []byte(strconv.Itoa(int(action))))
	r, err := cmd.Call(ctx, s)
	if err != nil {
		return "", err
	}
	if len(r) == 0 {
		return "", ErrorNullParsedResponse
	}
	return r, nil
}

func SystemQuery(ctx context.Context, s streamconn.Session, action SystemActions) (string, error) {
	response, err := System(ctx, s, false, action)
	if err != nil {
		return "", fmt.Errorf("%v: %w", action, err)
	}
	return response, nil
}

func GetTime(ctx context.Context, s streamconn.Session) (time.Time, error) {
	date, err := SystemQuery(ctx, s, SystemDate)
	if err != nil {
		return time.Time{}, err
	}
	tod, err := SystemQuery(ctx, s, SystemTime)
	if err != nil {
		return time.Time{}, err
	}
	tz, err := SystemQuery(ctx, s, SystemTimeZone)
	if err != nil {
		return time.Time{}, err
	}
	tzn := NormalizeTimeZone(tz)
	sysTime, err := time.Parse("01/02/2006 15:04:05 -07:00", date+" "+tod+" "+tzn)
	if err != nil {
		return time.Time{}, err
	}
	return sysTime, nil
}

func GetLatLong(ctx context.Context, s streamconn.Session) (float64, float64, error) {
	latlong, err := SystemQuery(ctx, s, SystemLatLong)
	if err != nil {
		return 0, 0, err
	}
	parts := strings.Split(latlong, ",")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("unexpected response: %v", latlong)
	}
	lat, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to parse latitude: %v", err)
	}
	long, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to parse longitude: %v", err)
	}
	return lat, long, nil
}

func GetSunriseSunset(ctx context.Context, s streamconn.Session) (time.Time, time.Time, error) {
	sunrise, err := SystemQuery(ctx, s, SystemSunrise)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	sunset, err := SystemQuery(ctx, s, SystemSunset)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	sunriseT, err := time.Parse("15:04:05", sunrise)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	sunsetT, err := time.Parse("15:04:05", sunset)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	return sunriseT, sunsetT, nil
}

func GetVersion(ctx context.Context, s streamconn.Session) (string, error) {
	data, err := SystemQuery(ctx, s, SystemOSRev)
	if err != nil {
		return "", err
	}
	return data, nil
}
