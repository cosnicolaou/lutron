// Copyright 2024 Cosmos Nicolaou. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package devices

import (
	"context"
	"time"

	"cloudeng.io/cmdutil/cmdyaml"
	"gopkg.in/yaml.v3"
)

var (
	AvailableControllers = SupportedControllers{}
	AvailableDevices     = SupportedDevices{}
)

type parameters struct {
	Parameters []string `yaml:",flow"`
}

type ControllerConfigCommon struct {
	Name       string              `yaml:"name"`
	Type       string              `yaml:"type"`
	Operations map[string][]string `yaml:"operations"`
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
	Name       string              `yaml:"name"`
	Type       string              `yaml:"type"`
	Controller string              `yaml:"controller"`
	Operations map[string][]string `yaml:"operations"`
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

func locationFromValue(value string) (*time.Location, error) {
	if len(value) == 0 {
		return time.Now().Location(), nil
	}
	location, err := time.LoadLocation(value)
	if err != nil {
		return nil, err
	}
	return location, nil
}

type LocationConfig struct {
	*time.Location
}

func (lc *LocationConfig) UnmarshalYAML(node *yaml.Node) error {
	l, err := locationFromValue(node.Value)
	if err != nil {
		return err
	}
	lc.Location = l
	return nil
}

type SystemConfig struct {
	Location    LocationConfig     `yaml:"location" cmd:"the location for which the schedule is being created in time.Location format"`
	Controllers []ControllerConfig `yaml:"controllers" cmd:"the controllers that are being configured"`
	Devices     []DeviceConfig     `yaml:"devices" cmd:"the devices that are being configured"`
}

type System struct {
	Location    *time.Location
	Controllers map[string]Controller
	Devices     map[string]Device
}

func ParseSystemConfigFile(ctx context.Context, place string, cfgFile string, opts ...Option) (System, error) {
	var cfg SystemConfig
	if err := cmdyaml.ParseConfigFile(ctx, cfgFile, &cfg); err != nil {
		return System{}, err
	}
	return cfg.CreateSystem(place, opts...)
}

func ParseSystemConfig(ctx context.Context, place string, cfgData []byte, opts ...Option) (System, error) {
	var cfg SystemConfig
	if err := yaml.Unmarshal(cfgData, &cfg); err != nil {
		return System{}, err
	}
	return cfg.CreateSystem(place, opts...)
}

// CreateSystem creates a system from the supplied configuration.
// The place argument is used to set the location of the system if
// the location is not specified in the configuration. Note that if the
// location: tag is specified in the configuration without a value
// then the location is set to the current time.Location, ie. timezone of 'Local'
func (cfg SystemConfig) CreateSystem(place string, opts ...Option) (System, error) {
	if cfg.Location.Location == nil {
		var err error
		cfg.Location.Location, err = locationFromValue(place)
		if err != nil {
			return System{}, err
		}
	}
	ctrl, dev, err := BuildDevices(cfg.Controllers, cfg.Devices, AvailableControllers, AvailableDevices, opts...)
	if err != nil {
		return System{}, err
	}
	return System{
		Location:    cfg.Location.Location,
		Controllers: ctrl,
		Devices:     dev,
	}, nil
}
