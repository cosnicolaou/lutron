// Copyright 2025 Cosmos Nicolaou. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package homeworks

import (
	"context"
	"log/slog"
	"strconv"
	"sync"
	"time"

	"cloudeng.io/logging/ctxlog"
	"github.com/cosnicolaou/automation/devices"
)

type ContactClosureOpenCloseConfig struct {
	OpenID            int           `yaml:"open_id"`
	CloseID           int           `yaml:"close_id"`
	PulseLow          bool          `yaml:"pulse_low"`
	PulseDuration     time.Duration `yaml:"pulse_duration"`
	OperationInterval time.Duration `yaml:"operation_interval"`
}

type ContactClosureOpenClose struct {
	devices.DeviceBase[ContactClosureOpenCloseConfig]
	processor *QSProcessor
	mu        sync.Mutex
}

func (cc *ContactClosureOpenClose) Operations() map[string]devices.Operation {
	return map[string]devices.Operation{
		"open":  cc.Open,
		"close": cc.Close,
	}
}

func (cc *ContactClosureOpenClose) OperationsHelp() map[string]string {
	return map[string]string{
		"open":  "pulse the contact closure to open",
		"close": "pulse the contact closure to close",
	}
}

func (cc *ContactClosureOpenClose) SetController(c devices.Controller) {
	cc.processor = c.Implementation().(*QSProcessor)
}

func (cc *ContactClosureOpenClose) ControlledBy() devices.Controller {
	return cc.processor
}

func (cc *ContactClosureOpenClose) pulse(ctx context.Context, id []byte, pulse, interval time.Duration) (any, error) {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	if cc.DeviceConfigCustom.PulseLow {
		return cc.processor.contactClosurePulse(ctx, id, pulse, interval, '0', '1')
	}
	return cc.processor.contactClosurePulse(ctx, id, pulse, interval, '1', '0')
}

func (cc *ContactClosureOpenClose) defaultIntervals() (time.Duration, time.Duration) {
	// Default pulse duration is 10 milliseconds.
	pulse := cc.DeviceConfigCustom.PulseDuration
	if pulse == 0 {
		cc.DeviceConfigCustom.PulseDuration = 10 * time.Millisecond
	}
	// Default operation interval is 30 seconds.
	// If the operation interval is not set, we use a default of 30 seconds.
	// This is to ensure that the device does not get polled too frequently.
	interval := cc.DeviceConfigCustom.OperationInterval
	if interval == 0 {
		interval = time.Second * 30
	}
	return pulse, interval
}

func (cc *ContactClosureOpenClose) Open(ctx context.Context, _ devices.OperationArgs) (any, error) {
	ids := strconv.Itoa(cc.DeviceConfigCustom.OpenID)
	id := []byte(ids)
	pulse, interval := cc.defaultIntervals()
	grp := slog.Group("lutron",
		"device", "contact-closure",
		"id", ids,
		"op", "open",
		"pulse", pulse.String(),
		"interval", interval.String())
	ctx = ctxlog.WithAttributes(ctx, grp)
	return cc.pulse(ctx, id, pulse, interval)
}

func (cc *ContactClosureOpenClose) Close(ctx context.Context, _ devices.OperationArgs) (any, error) {
	ids := strconv.Itoa(cc.DeviceConfigCustom.CloseID)
	id := []byte(ids)
	pulse, interval := cc.defaultIntervals()
	grp := slog.Group("lutron",
		"device", "contact-closure",
		"id", ids,
		"op", "close",
		"pulse", pulse.String(),
		"interval", interval.String())
	ctx = ctxlog.WithAttributes(ctx, grp)
	return cc.pulse(ctx, id, pulse, interval)
}
