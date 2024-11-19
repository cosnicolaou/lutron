// Copyright 2024 Cosmos Nicolaou. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/cosnicolaou/lutron/devices"
	"gopkg.in/yaml.v3"
)

type ConfigFileFlags struct {
	KeysFile   string `subcmd:"keys,$HOME/.lutron-keys.yaml,path/URI to a file containing keys"`
	SystemFile string `subcmd:"system,$HOME/.lutron-system.yaml,path to a file containing the lutron system configuration"`
}

type ConfigFlags struct {
	ConfigFileFlags
}

type Config struct {
}

func marshalYAML(indent string, v any) string {
	p, _ := yaml.Marshal(v)
	lines := strings.Split(string(p), "\n")
	indented := make([]string, len(lines))
	for i, line := range lines {
		indented[i] = indent + line
	}
	return strings.Join(indented, "\n")
}

func (c *Config) Display(ctx context.Context, flags any, args []string) error {
	fv := flags.(*ConfigFlags)
	keys, err := ReadKeysFile(ctx, fv.KeysFile)
	if err != nil {
		return err
	}

	system, err := devices.ParseSystemConfigFile(ctx, "", fv.SystemFile)
	if err != nil {
		return err
	}

	fmt.Printf("Keys:\n")
	for _, key := range keys {
		fmt.Printf("  %v\n", key)
	}

	fmt.Printf("\nLocation: %v\n\n", system.Location)

	for _, controller := range system.Controllers {
		fmt.Printf("Controller:\n%v\n", marshalYAML("  ", controller.Config()))
		fmt.Printf("%v\n", marshalYAML("  ", controller.CustomConfig()))
	}

	for _, device := range system.Devices {
		fmt.Printf("Device: %v\n", marshalYAML("  ", device.Config()))
		fmt.Printf("Device Controlled By: %v\n", device.ControlledByName())
		fmt.Printf("Device Custom Config: %v\n", marshalYAML("  ", device.CustomConfig()))
	}

	return nil
}

func opNames[Map ~map[string]V, V any](m Map) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}

func (c *Config) Operations(ctx context.Context, flags any, args []string) error {
	fv := flags.(*ConfigFlags)
	system, err := devices.ParseSystemConfigFile(ctx, "", fv.SystemFile)
	if err != nil {
		return err
	}

	for _, cfg := range system.Config.Controllers {
		available := system.Controllers[cfg.Name].Operations()
		sorted := opNames(available)
		fmt.Printf("Controller: %v\n", cfg.Name)
		for _, op := range sorted {
			_, configured := cfg.Operations[op]
			if !configured {
				fmt.Printf("  %v: but not configured\n", op)
				continue
			}
			h := system.Controllers[cfg.Name].OperationsHelp()[op]
			fmt.Printf("  %v:  %v\n", op, h)

		}
	}

	for _, cfg := range system.Config.Devices {
		available := system.Devices[cfg.Name].Operations()
		sorted := opNames(available)
		fmt.Printf("Device: %v\n", cfg.Name)
		for _, op := range sorted {
			_, configured := cfg.Operations[op]
			if !configured {
				fmt.Printf("  %v: but not configured\n", op)
				continue
			}
			h := system.Devices[cfg.Name].OperationsHelp()[op]
			fmt.Printf("  %v:  %v\n", op, h)
		}
	}
	return nil
}
