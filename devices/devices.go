// Copyright 2024 Cosmos Nicolaou. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package devices

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"gopkg.in/yaml.v3"
)

type Controller interface {
	SetConfig(ControllerConfigCommon)
	Config() ControllerConfigCommon
	UnmarshalYAML(*yaml.Node) error
	Implementation() any
}

type Operation func(ctx context.Context) error

type Device interface {
	SetConfig(DeviceConfigCommon)
	Config() DeviceConfigCommon
	SetController(Controller)
	UnmarshalYAML(*yaml.Node) error
	ControlledByName() string
	ControlledBy() Controller
	Operations() map[string]Operation
	Timeout() time.Duration
}

type Option func(*Options)

type Options struct {
	Logger *slog.Logger
}

type SupportedControllers map[string]func(typ string, opts ...Option) (Controller, error)

type SupportedDevices map[string]func(typ string, opts ...Option) (Device, error)

func BuildDevices(controllerCfg []ControllerConfig, deviceCfg []DeviceConfig, supportedControllers SupportedControllers, supportedDevices SupportedDevices) (map[string]Controller, map[string]Device, error) {
	controllers, err := CreateControllers(controllerCfg, supportedControllers)
	if err != nil {
		return nil, nil, err
	}
	devices, err := CreateDevices(deviceCfg, supportedDevices)
	if err != nil {
		return nil, nil, err
	}
	for _, dev := range devices {
		if ctrl, ok := controllers[dev.ControlledByName()]; ok {
			dev.SetController(ctrl)
		}
	}
	return controllers, devices, nil
}

func CreateControllers(config []ControllerConfig, supported SupportedControllers) (map[string]Controller, error) {
	controllers := map[string]Controller{}
	for _, ctrlcfg := range config {
		f, ok := supported[ctrlcfg.Type]
		if !ok {
			return nil, fmt.Errorf("unsupported controller type: %s", ctrlcfg.Type)
		}
		ctrl, err := f(ctrlcfg.Type)
		if err != nil {
			return nil, fmt.Errorf("failed to create controller %v: %w", ctrlcfg.Type, err)
		}
		ctrl.SetConfig(ctrlcfg.ControllerConfigCommon)
		if err := ctrl.UnmarshalYAML(&ctrlcfg.Config); err != nil {
			return nil, fmt.Errorf("failed to unmarshal controller %v: %w", ctrlcfg.Type, err)
		}
		controllers[ctrlcfg.Name] = ctrl
	}
	return controllers, nil
}

func CreateDevices(config []DeviceConfig, supported SupportedDevices) (map[string]Device, error) {
	devices := map[string]Device{}
	for _, devcfg := range config {
		f, ok := supported[devcfg.Type]
		if !ok {
			return nil, fmt.Errorf("unsupported device type: %s", devcfg.Type)
		}
		dev, err := f(devcfg.Type)
		if err != nil {
			return nil, fmt.Errorf("failed to create device %v: %w", devcfg.Type, err)
		}
		dev.SetConfig(devcfg.DeviceConfigCommon)
		if err := dev.UnmarshalYAML(&devcfg.Config); err != nil {
			return nil, fmt.Errorf("failed to unmarshal device %v: %w", devcfg.Type, err)
		}
		devices[devcfg.Name] = dev
	}
	return devices, nil
}
