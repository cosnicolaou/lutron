// Copyright 2024 Cosmos Nicolaou. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package homeworks

/*
import (
	"context"
	"io"

	"github.com/cosnicolaou/lutron/devices"
	"gopkg.in/yaml.v3"
)

type HWShadeGroupConfig struct {
	ID    int `yaml:"id"`
	Level int `yaml:"level"`
}

type HWShadeGroup struct {
	hwDeviceBase
	HWShadeGroupConfig
}

func (b *HWShadeGroup) Operations() map[string]devices.Operation {
	return map[string]devices.Operation{
		"raise": b.raise,
		"level": b.level,
	}
}

func (b *HWShadeGroup) OperationsHelp() map[string]string {
	return map[string]string{
		"raise": "raise the shade",
		"level": "set the shade level",
	}
}

func (b *HWShadeGroup) CustomConfig() any {
	return b.HWShadeGroupConfig
}

func (b *HWShadeGroup) raise(context.Context, io.Writer, ...string) error {
	return nil
}

func (b *HWShadeGroup) level(context.Context, io.Writer, ...string) error {
	return nil
}

func (d *HWShadeGroup) UnmarshalYAML(node *yaml.Node) error {
	return node.Decode(&d.HWShadeGroupConfig)
}

type HWShade struct {
	HWShadeGroup
}

func (b *HWShade) Operations() map[string]devices.Operation {
	return map[string]devices.Operation{
		"raise": b.raise,
		"level": b.level,
	}
}

func (b *HWShade) OperationsHelp() map[string]string {
	return map[string]string{
		"raise": "raise the shade",
		"level": "set the shade level",
	}
}
*/
