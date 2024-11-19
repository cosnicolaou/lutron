// Copyright 2024 Cosmos Nicolaou. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package homeworks

import (
	"context"
	"fmt"
	"io"
	"strconv"

	"github.com/cosnicolaou/automation/devices"
	"github.com/cosnicolaou/lutron/protocol"
	"gopkg.in/yaml.v3"
)

func parseShadeLevel(pars []string) (int, error) {
	level := 50
	if len(pars) != 1 {
		return 0, fmt.Errorf("must specify a level")
	}
	l, err := strconv.Atoi(pars[0])
	if err != nil || (level < 0 || level > 100) {
		return 0, fmt.Errorf("level must be in the range 0..100")
	}
	return l, nil
}

type HWShadeConfig struct {
	ID int `yaml:"id"`
}

func (sb HWShadeConfig) OperationsHelp() map[string]string {
	return map[string]string{
		"raise": "raise the shade",
		"lower": "lower the shade",
		"set":   "set the shade level",
	}
}

func (sb HWShadeConfig) operations(raise, lower, set devices.Operation) map[string]devices.Operation {
	return map[string]devices.Operation{
		"raise": raise,
		"lower": lower,
		"set":   set,
	}
}

func (sb HWShadeConfig) shadeCommand(ctx context.Context, s protocol.Session, cg protocol.CommandGroup, pars []byte) error {
	_, err := protocol.NewCommand(cg, true, pars).Call(ctx, s)
	return err
}

func (sb HWShadeConfig) raiseShade(ctx context.Context, s protocol.Session, cg protocol.CommandGroup) error {
	pars := append([]byte(strconv.Itoa(int(sb.ID))), ',', '2')
	return sb.shadeCommand(ctx, s, cg, pars)
}

func (sb HWShadeConfig) lowerShade(ctx context.Context, s protocol.Session, cg protocol.CommandGroup) error {
	pars := append([]byte(strconv.Itoa(int(sb.ID))), ',', '3')
	return sb.shadeCommand(ctx, s, cg, pars)
}

func (sb HWShadeConfig) setShadeLevel(ctx context.Context, s protocol.Session, cg protocol.CommandGroup, args []string) error {
	level, err := parseShadeLevel(args)
	if err != nil {
		return err
	}
	pars := append([]byte(strconv.Itoa(int(sb.ID))), ',', '1', ',')
	pars = append(pars, []byte(strconv.Itoa(level))...)
	return sb.shadeCommand(ctx, s, cg, pars)
}

// HWShadeGroupConfig represents the configuration for a group of shades
// as configured as a single group.
type HWShadeGroup struct {
	hwDeviceBase
	HWShadeConfig
}

// HWShadeConfig represents the configuration for a single shade.
type HWShade struct {
	hwDeviceBase
	HWShadeConfig
}

func (sg *HWShadeGroup) CustomConfig() any {
	return sg.HWShadeConfig
}

func (sg *HWShadeGroup) UnmarshalYAML(node *yaml.Node) error {
	return node.Decode(&sg.HWShadeConfig)
}

func (sg *HWShadeGroup) Operations() map[string]devices.Operation {
	return sg.operations(sg.raise, sg.lower, sg.set)
}

func (sg *HWShadeGroup) raise(ctx context.Context, out io.Writer, _ ...string) error {
	sess := sg.processor.Session(ctx)
	return sg.raiseShade(ctx, sess, protocol.ShadeGroupCommands)
}

func (sg *HWShadeGroup) lower(ctx context.Context, out io.Writer, _ ...string) error {
	sess := sg.processor.Session(ctx)
	return sg.lowerShade(ctx, sess, protocol.ShadeGroupCommands)
}

func (sg *HWShadeGroup) set(ctx context.Context, out io.Writer, pars ...string) error {
	sess := sg.processor.Session(ctx)
	return sg.setShadeLevel(ctx, sess, protocol.ShadeGroupCommands, pars)
}

func (s *HWShade) CustomConfig() any {
	return s.HWShadeConfig
}

func (s *HWShade) UnmarshalYAML(node *yaml.Node) error {
	return node.Decode(&s.HWShadeConfig)
}

func (s *HWShade) Operations() map[string]devices.Operation {
	return s.operations(s.raise, s.lower, s.set)
}

func (sg *HWShade) raise(ctx context.Context, out io.Writer, _ ...string) error {
	sess := sg.processor.Session(ctx)
	return sg.raiseShade(ctx, sess, protocol.OutputCommands)
}

func (sg *HWShade) lower(ctx context.Context, out io.Writer, _ ...string) error {
	sess := sg.processor.Session(ctx)
	return sg.lowerShade(ctx, sess, protocol.OutputCommands)
}

func (sg *HWShade) set(ctx context.Context, out io.Writer, pars ...string) error {
	sess := sg.processor.Session(ctx)
	return sg.setShadeLevel(ctx, sess, protocol.OutputCommands, pars)
}
