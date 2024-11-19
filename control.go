// Copyright 2024 Cosmos Nicolaou. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"cloudeng.io/cmdutil/keystore"
	"github.com/cosnicolaou/lutron/devices"
)

type ControlFlags struct {
	ConfigFileFlags
}

type Control struct{}

func (c *Control) runOp(ctx context.Context, system devices.System, nameAndOp string, args []string) error {
	parts := strings.Split(nameAndOp, ".")
	if len(parts) != 2 {
		return fmt.Errorf("invalid operation: %v, should be name.operation", nameAndOp)
	}
	name, op := parts[0], parts[1]
	_, cok := system.Controllers[name]
	_, dok := system.Devices[name]
	if !cok && !dok {
		return fmt.Errorf("unknown controller or device: %v", name)
	}

	if fn, pars, ok := system.ControllerOp(name, op); ok {
		if len(args) == 0 {
			args = pars
		}
		if err := fn(ctx, os.Stdout, args...); err != nil {
			return fmt.Errorf("failed to run operation: %v: %v", op, err)
		}
		return nil
	}
	if fn, pars, ok := system.DeviceOp(name, op); ok {
		if len(args) == 0 {
			args = pars
		}
		if err := fn(ctx, os.Stdout, args...); err != nil {
			return fmt.Errorf("failed to run operation: %v: %v", op, err)
		}
		return nil
	}
	return fmt.Errorf("unknown or not configured operation: %v, %v", name, op)
}

func (c *Control) Run(ctx context.Context, flags any, args []string) error {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	opts := []devices.Option{
		devices.WithLogger(logger),
	}
	fv := flags.(*ControlFlags)
	keys, err := ReadKeysFile(ctx, fv.KeysFile)
	if err != nil {
		return err
	}
	system, err := devices.ParseSystemConfigFile(ctx, " ", fv.SystemFile, opts...)
	if err != nil {
		return err
	}
	cmd := args[0]
	parameters := args[1:]
	ctx = keystore.ContextWithAuth(ctx, keys)
	if err := c.runOp(ctx, system, cmd, parameters); err != nil {
		return err
	}
	return nil
}
