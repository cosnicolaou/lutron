// Copyright 2024 Cosmos Nicolaou. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package homeworks

import (
	"context"

	"github.com/cosnicolaou/lutron/devices"
	"gopkg.in/yaml.v3"
)

type HWBlindConfig struct {
	ID    int `yaml:"id"`
	Level int `yaml:"level"`
}

type hwBlind struct {
	hwDeviceBase
	HWBlindConfig
}

func (b *hwBlind) Operations() map[string]devices.Operation {
	return map[string]devices.Operation{
		"raise": b.raise,
		"level": b.level,
	}
}

func (b *hwBlind) CustomConfig() any {
	return b.HWBlindConfig
}

func (b *hwBlind) raise(context.Context, ...string) error {
	return nil
}

func (b *hwBlind) level(context.Context, ...string) error {
	return nil
}

func (d *hwBlind) UnmarshalYAML(node *yaml.Node) error {
	return node.Decode(&d.HWBlindConfig)
}
