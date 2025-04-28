// Copyright 2024 Cosmos Nicolaou. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package homeworks

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"

	"cloudeng.io/logging/ctxlog"
	"github.com/cosnicolaou/automation/devices"
	"github.com/cosnicolaou/automation/net/streamconn"
	"github.com/cosnicolaou/lutron/protocol"
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

type hwShadeBase struct {
	devices.DeviceBase[HWShadeConfig]
	processor *QSProcessor
}

func (sb *hwShadeBase) SetController(c devices.Controller) {
	sb.processor = c.Implementation().(*QSProcessor)
}

func (sb *hwShadeBase) ControlledBy() devices.Controller {
	return sb.processor
}

func (sb hwShadeBase) OperationsHelp() map[string]string {
	return map[string]string{
		"raise": "raise the shade",
		"lower": "lower the shade",
		"stop":  "stop the shade",
		"set":   "set the shade level",
	}
}

func (sb hwShadeBase) operations(raise, lower, stop, set devices.Operation) map[string]devices.Operation {
	return map[string]devices.Operation{
		"raise": raise,
		"lower": lower,
		"stop":  stop,
		"set":   set,
	}
}

func (sb hwShadeBase) runShadeCommand(ctx context.Context, s streamconn.Session, cg protocol.CommandGroup, pars []byte, op string) (any, error) {
	grp := slog.Group("lutron", "device", "shade", "id", sb.DeviceConfigCustom.ID, "op", op)
	ctx = ctxlog.WithAttributes(ctx, grp)
	err := protocol.NewCommand(cg, true, pars).Invoke(ctx, s)
	return nil, err
}

func (sb hwShadeBase) raiseShade(ctx context.Context, s streamconn.Session, cg protocol.CommandGroup) (any, error) {
	pars := append([]byte(strconv.Itoa(sb.DeviceConfigCustom.ID)), ',', '2')
	return sb.runShadeCommand(ctx, s, cg, pars, "raise")
}

func (sb hwShadeBase) lowerShade(ctx context.Context, s streamconn.Session, cg protocol.CommandGroup) (any, error) {
	pars := append([]byte(strconv.Itoa(sb.DeviceConfigCustom.ID)), ',', '3')
	return sb.runShadeCommand(ctx, s, cg, pars, "lower")
}

func (sb hwShadeBase) stopShade(ctx context.Context, s streamconn.Session, cg protocol.CommandGroup) (any, error) {
	pars := append([]byte(strconv.Itoa(sb.DeviceConfigCustom.ID)), ',', '4')
	return sb.runShadeCommand(ctx, s, cg, pars, "stop")
}

func (sb hwShadeBase) setShadeLevel(ctx context.Context, s streamconn.Session, cg protocol.CommandGroup, args []string) (any, error) {
	level, err := parseShadeLevel(args)
	if err != nil {
		return nil, err
	}
	pars := append([]byte(strconv.Itoa(sb.DeviceConfigCustom.ID)), ',', '1', ',')
	pars = append(pars, []byte(strconv.Itoa(level))...)
	return sb.runShadeCommand(ctx, s, cg, pars, "set")
}

// HWShadeGroupConfig represents the configuration for a group of shades
// as configured as a single group.
type HWShadeGroup struct {
	hwShadeBase
}

// HWShadeConfig represents the configuration for a single shade.
type HWShade struct {
	hwShadeBase
}

func (sg *HWShadeGroup) Operations() map[string]devices.Operation {
	return sg.operations(sg.raise, sg.lower, sg.stop, sg.set)
}

func (sg *HWShadeGroup) raise(ctx context.Context, _ devices.OperationArgs) (any, error) {
	ctx, sess := sg.processor.Session(ctx)
	return sg.raiseShade(ctx, sess, protocol.ShadeGroupCommands)
}

func (sg *HWShadeGroup) lower(ctx context.Context, _ devices.OperationArgs) (any, error) {
	ctx, sess := sg.processor.Session(ctx)
	return sg.lowerShade(ctx, sess, protocol.ShadeGroupCommands)
}

func (sg *HWShadeGroup) stop(ctx context.Context, _ devices.OperationArgs) (any, error) {
	ctx, sess := sg.processor.Session(ctx)
	return sg.stopShade(ctx, sess, protocol.ShadeGroupCommands)
}

func (sg *HWShadeGroup) set(ctx context.Context, args devices.OperationArgs) (any, error) {
	ctx, sess := sg.processor.Session(ctx)
	return sg.setShadeLevel(ctx, sess, protocol.ShadeGroupCommands, args.Args)
}

func (s *HWShade) Operations() map[string]devices.Operation {
	return s.operations(s.raise, s.lower, s.stop, s.set)
}

func (s *HWShade) raise(ctx context.Context, _ devices.OperationArgs) (any, error) {
	ctx, sess := s.processor.Session(ctx)
	return s.raiseShade(ctx, sess, protocol.OutputCommands)
}

func (s *HWShade) lower(ctx context.Context, _ devices.OperationArgs) (any, error) {
	ctx, sess := s.processor.Session(ctx)
	return s.lowerShade(ctx, sess, protocol.OutputCommands)
}

func (s *HWShade) stop(ctx context.Context, _ devices.OperationArgs) (any, error) {
	ctx, sess := s.processor.Session(ctx)
	return s.stopShade(ctx, sess, protocol.OutputCommands)
}

func (s *HWShade) set(ctx context.Context, args devices.OperationArgs) (any, error) {
	ctx, sess := s.processor.Session(ctx)
	return s.setShadeLevel(ctx, sess, protocol.OutputCommands, args.Args)
}
