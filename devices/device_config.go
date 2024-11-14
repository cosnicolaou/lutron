// Copyright 2024 Cosmos Nicolaou. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package devices

import (
	"gopkg.in/yaml.v3"
)

type ControllerConfigCommon struct {
	Name string `yaml:"name"`
	Type string `yaml:"type"`
}

type ControllerConfig struct {
	ControllerConfigCommon
	Config yaml.Node `yaml:",inline"`
}

func (lp *ControllerConfig) UnmarshalYAML(node *yaml.Node) error {
	if err := node.Decode(&lp.ControllerConfigCommon); err != nil {
		return err
	}
	return node.Decode(&lp.Config)
}

type DeviceConfigCommon struct {
	Name       string `yaml:"name"`
	Type       string `yaml:"type"`
	Controller string `yaml:"controller"`
}

type DeviceConfig struct {
	DeviceConfigCommon
	Config yaml.Node `yaml:",inline"`
}

func (lp *DeviceConfig) UnmarshalYAML(node *yaml.Node) error {
	if err := node.Decode(&lp.DeviceConfigCommon); err != nil {
		return err
	}
	return node.Decode(&lp.Config)
}
