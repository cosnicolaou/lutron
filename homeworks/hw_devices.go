// Copyright 2024 Cosmos Nicolaou. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package homeworks

import (
	"context"
	"time"

	"github.com/cosnicolaou/lutron/devices"
	"gopkg.in/yaml.v3"
)

type hwDeviceBase struct {
	deviceConfig devices.DeviceConfigCommon
	controller   devices.Controller
	processor    *qsProcessor
}

func (d *hwDeviceBase) SetConfig(c devices.DeviceConfigCommon) {
	d.deviceConfig = c
}

func (d *hwDeviceBase) Config() devices.DeviceConfigCommon {
	return d.deviceConfig
}

func (d *hwDeviceBase) SetController(c devices.Controller) {
	d.controller = c
	d.processor = c.Implementation().(*qsProcessor)
}

func (d *hwDeviceBase) ControlledByName() string {
	return d.deviceConfig.Controller
}

func (d *hwDeviceBase) ControlledBy() devices.Controller {
	return d.controller
}

func (d *hwDeviceBase) Implementation() any {
	return d
}

func (d *hwDeviceBase) Timeout() time.Duration {
	return time.Minute
}

func (d *hwDeviceBase) UnmarshalYAML(node *yaml.Node) error {
	return nil
}

type hwBlindConfig struct {
	ID    string `yaml:"id"`
	Level int    `yaml:"level"`
}

type hwBlind struct {
	hwDeviceBase
	hwBlindConfig
}

func (b *hwBlind) Operations() map[string]devices.Operation {
	return map[string]devices.Operation{
		"raise": b.raise,
		"level": b.level,
	}
}

func (b *hwBlind) raise(context.Context) error {
	return nil
}

func (b *hwBlind) level(context.Context) error {
	return nil
}

func (d *hwBlind) UnmarshalYAML(node *yaml.Node) error {
	if err := node.Decode(&d.hwBlindConfig); err != nil {
		return err
	}
	return nil
}
