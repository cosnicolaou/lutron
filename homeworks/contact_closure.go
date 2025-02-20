// Copyright 2025 Cosmos Nicolaou. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package homeworks

import (
	"context"
	"strconv"
	"time"

	"github.com/cosnicolaou/automation/devices"
	"github.com/cosnicolaou/automation/net/streamconn"
	"github.com/cosnicolaou/lutron/protocol"
)

type ContactClosureConfig struct {
	ID       int           `yaml:"id"`
	Duration time.Duration `yaml:"duration"`
}

type ContactClosure struct {
	devices.DeviceBase[ContactClosureConfig]

	processor *QSProcessor
}

func (cc *ContactClosure) Operations() map[string]devices.Operation {
	return map[string]devices.Operation{
		"pulse-on":  cc.PulseOn,
		"pulse-off": cc.PulseOff,
	}
}

func (cc *ContactClosure) OperationsHelp() map[string]string {
	return map[string]string{
		"pulse-on":  "pulse the contact closure to on",
		"pulse-off": "pulse the contact closure to off",
	}
}

func (cc *ContactClosure) SetController(c devices.Controller) {
	cc.processor = c.Implementation().(*QSProcessor)
}

func (cc *ContactClosure) ControlledBy() devices.Controller {
	return cc.processor
}

func (cc *ContactClosure) PulseOn(ctx context.Context, _ devices.OperationArgs) (any, error) {
	s := cc.processor.Session(ctx)
	return contactClosurePulse(ctx, s, []byte(strconv.Itoa(cc.DeviceConfigCustom.ID)), cc.DeviceConfigCustom.Duration, '1', '0')
}

func (cc *ContactClosure) PulseOff(ctx context.Context, _ devices.OperationArgs) (any, error) {
	s := cc.processor.Session(ctx)
	return contactClosurePulse(ctx, s, []byte(strconv.Itoa(cc.DeviceConfigCustom.ID)), cc.DeviceConfigCustom.Duration, '0', '1')
}

func contactClosurePulse(ctx context.Context, s streamconn.Session, id []byte, pulse time.Duration, l0, l1 byte) (any, error) {
	pars := make([]byte, 0, 32)
	pars = append(pars, id...)
	pars = append(pars, ',', '1', ',', l0)
	_, err := protocol.NewCommand(protocol.OutputCommands, true, pars).Call(ctx, s)
	if err != nil {
		return nil, err
	}
	time.Sleep(pulse)
	pars[len(pars)-1] = l1
	_, err = protocol.NewCommand(protocol.OutputCommands, true, pars).Call(ctx, s)
	return nil, err
}

type ContactClosureOpenCloseConfig struct {
	OpenID   int           `yaml:"open_id"`
	CloseID  int           `yaml:"close_id"`
	PulseLow bool          `yaml:"pulse_low"`
	Duration time.Duration `yaml:"duration"`
}

type ContactClosureOpenClose struct {
	devices.DeviceBase[ContactClosureOpenCloseConfig]
	processor *QSProcessor
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

func (cc *ContactClosureOpenClose) pulse(ctx context.Context, id []byte) (any, error) {
	s := cc.processor.Session(ctx)
	if cc.DeviceConfigCustom.PulseLow {
		return contactClosurePulse(ctx, s, id, cc.DeviceConfigCustom.Duration, '0', '1')
	}
	return contactClosurePulse(ctx, s, id, cc.DeviceConfigCustom.Duration, '1', '0')
}

func (cc *ContactClosureOpenClose) Open(ctx context.Context, _ devices.OperationArgs) (any, error) {
	id := []byte(strconv.Itoa(cc.DeviceConfigCustom.OpenID))
	return cc.pulse(ctx, id)
}

func (cc *ContactClosureOpenClose) Close(ctx context.Context, _ devices.OperationArgs) (any, error) {
	id := []byte(strconv.Itoa(cc.DeviceConfigCustom.CloseID))
	return cc.pulse(ctx, id)
}
