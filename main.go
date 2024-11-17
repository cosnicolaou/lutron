// Copyright 2024 Cosmos Nicolaou. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"errors"
	"os"

	"cloudeng.io/cmdutil"
	"cloudeng.io/cmdutil/cmdyaml"
	"cloudeng.io/cmdutil/subcmd"
	"cloudeng.io/macos/keychainfs"
	"github.com/cosnicolaou/lutron/devices"
	"github.com/cosnicolaou/lutron/homeworks"
)

const cmdSpec = `name: lutron
summary: lutron is a command line tool for interacting with lutron systems
commands:
  - name: control
    summary: issue a series of commands to control/interact with a lutron system
    arguments:
      - <name.op> - name of the device or controller and the operation to perform
      - <parameters>...
  - name: config
    summary: query/inspect the configuration file
    commands:
	  - name: display
`

func cli() *subcmd.CommandSetYAML {
	cmd := subcmd.MustFromYAML(cmdSpec)
	control := &Control{}
	cmd.Set("control").MustRunner(control.Run, &ControlFlags{})
	config := &Config{}
	cmd.Set("config", "display").MustRunner(config.Display, &ConfigFlags{})
	return cmd
}

var URIHandlers = map[string]cmdyaml.URLHandler{
	"keychain": keychainfs.NewSecureNoteFSFromURL,
}

func init() {
	devices.AvailableControllers = homeworks.SupportedControllers()
	devices.AvailableDevices = homeworks.SupportedDevices()
}

var interrupt = errors.New("interrupt")

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancelCause(ctx)
	cmdutil.HandleSignals(func() { cancel(interrupt) }, os.Interrupt)
	err := cli().Dispatch(ctx)
	if context.Cause(ctx) == interrupt {
		cmdutil.Exit("%v", interrupt)
	}
	cmdutil.Exit("%v", err)
}
