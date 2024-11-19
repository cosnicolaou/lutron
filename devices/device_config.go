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
	Config      SystemConfig
	Location    *time.Location
	Controllers map[string]Controller
	Devices     map[string]Device
}

func configuredAndExists(name string, configured map[string][]string, operations map[string]Operation) (Operation, bool) {
	if ops, ok := configured[name]; ok {
		for _, op := range ops {
			if fn, ok := operations[op]; ok {
				return fn, true
			}
		}
	}
	return nil, false

}

func (s System) ControllerConfigs(name string) (ControllerConfig, Controller, bool) {
	if ctrl, ok := s.Controllers[name]; ok {
		for _, cfg := range s.Config.Controllers {
			if cfg.Name == name {
				return cfg, ctrl, true
			}
		}
	}
	return ControllerConfig{}, nil, false
}

func (s System) DeviceConfigs(name string) (DeviceConfig, Device, bool) {
	if dev, ok := s.Devices[name]; ok {
		for _, cfg := range s.Config.Devices {
			if cfg.Name == name {
				return cfg, dev, true
			}
		}
	}
	return DeviceConfig{}, nil, false
}

// ControllerOp returns the operation function (and any configured parameters)
// for the specified operation on the named controller. The operation must be
// 'configured', ie. listed in the operations: list for the controller to be
// returned.
func (s System) ControllerOp(name, op string) (Operation, []string, bool) {
	if cfg, ctrl, ok := s.ControllerConfigs(name); ok {
		if fn, ok := ctrl.Operations()[op]; ok {
			if pars, ok := cfg.Operations[op]; ok {
				return fn, pars, true
			}
		}
	}
	return nil, nil, false
}

// DeviceOp returns the operation function (and any configured parameters)
// for the specified operation on the named controller. The operation must be
// 'configured', ie. listed in the operations: list for the controller to be
// returned.
func (s System) DeviceOp(name, op string) (Operation, []string, bool) {
	if cfg, dev, ok := s.DeviceConfigs(name); ok {
		if fn, ok := dev.Operations()[op]; ok {
			if pars, ok := cfg.Operations[op]; ok {
				return fn, pars, true
			}
		}
	}
	return nil, nil, false
}

// ParseSystemConfigFile parses the supplied configuration file as per ParseSystemConfig.
func ParseSystemConfigFile(ctx context.Context, place string, cfgFile string, opts ...Option) (System, error) {
	var cfg SystemConfig
	if err := cmdyaml.ParseConfigFile(ctx, cfgFile, &cfg); err != nil {
		return System{}, err
	}
	return cfg.CreateSystem(place, opts...)
}

// ParseSystemConfig parses the supplied configuration data and returns
// a System using CreateSystem.
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
		Config:      cfg,
		Location:    cfg.Location.Location,
		Controllers: ctrl,
		Devices:     dev,
	}, nil
}
