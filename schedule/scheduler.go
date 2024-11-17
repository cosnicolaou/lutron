// Copyright 2024 Cosmos Nicolaou. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package schedule

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"log/slog"
	"os"
	"sync"
	"time"

	"cloudeng.io/datetime"
	"cloudeng.io/datetime/schedule"
	"github.com/cosnicolaou/lutron/devices"
)

var OpTimeout = errors.New("op-timeout")

func (s *Scheduler) runOps(ctx context.Context, due time.Time, active schedule.Active[devices.Action]) {
	if len(active.Actions) == 0 {
		return
	}
	for _, action := range active.Actions {
		localAction := action.Action
		timeout := localAction.Device.Timeout()
		ctx, cancel := context.WithTimeoutCause(ctx, timeout, OpTimeout)
		var errCh = make(chan error)
		go func() {
			errCh <- localAction.Action(ctx)
		}()
		var err error
		select {
		case <-ctx.Done():
			err = ctx.Err()
			cancel()
		case err = <-errCh:
			cancel()
		}
		if err != nil {
			s.logger.Warn("failed", "op", localAction.ActionName, "due", due, "err", err)
		} else {
			s.logger.Info("ok", "op", localAction.ActionName, "due", due)
		}
	}
}

func (s *Scheduler) Scheduled(yp datetime.YearAndPlace) iter.Seq[schedule.Active[devices.Action]] {
	return s.scheduler.Scheduled(yp)
}

func actionAndDeviceNames(active schedule.Active[devices.Action]) (actionNames, deviceNames []string) {
	for _, a := range active.Actions {
		actionNames = append(actionNames, a.Action.ActionName)
		deviceNames = append(deviceNames, a.Action.DeviceName)
	}
	return actionNames, deviceNames
}

func (s *Scheduler) RunYear(ctx context.Context, yp datetime.YearAndPlace) error {
	var wg sync.WaitGroup
	for active := range s.scheduler.Scheduled(yp) {
		s.logger.Info("ok", "#actions", len(active.Actions))
		if len(active.Actions) == 0 {
			continue
		}
		dueAt := datetime.Time(yp, active.Date, active.Actions[0].Due)
		select {
		case <-ctx.Done():
			wg.Wait()
			return ctx.Err()
		case <-time.After(dueAt.Sub(s.timeSource.NowIn(s.place))):
			wg.Add(1)
			go func() {
				s.runOps(ctx, dueAt, active)
				wg.Done()
			}()
			now := time.Now().In(s.place)
			late := dueAt.Add(1 * time.Minute).In(s.place)
			if now.After(late) {
				actions, devices := actionAndDeviceNames(active)
				s.logger.Warn("late", "due", dueAt, "late", late, "now", now, "actions", actions, "devices", devices)
			}
		}
	}
	wg.Wait()
	return nil
}

func (s *Scheduler) RunYears(ctx context.Context, yp datetime.YearAndPlace, nYears int) error {
	for y := 0; y < nYears; y++ {
		if err := s.RunYear(ctx, yp); err != nil {
			return err
		}
		yp.Year++
	}
	return nil
}

type Scheduler struct {
	options
	schedule  schedule.Annual[devices.Action]
	scheduler *schedule.AnnualScheduler[devices.Action]
	place     *time.Location
}

type Option func(o *options)

type options struct {
	timeSource timeSource
	logger     *slog.Logger
}

// timeSource is an interface that provides the current time in a specific
// location and is intended for testing purposes. It will be called once
// per iteration of the scheduler to schedule the next action. time.Now().In()
// will be used for all other time operations.
type timeSource interface {
	NowIn(in *time.Location) time.Time
}

type systemTimeSource struct{}

func (systemTimeSource) NowIn(loc *time.Location) time.Time {
	return time.Now().In(loc)
}

// WithTimeSource sets the time source to be used by the scheduler and
// is primarily intended for testing purposes.
func withTimeSource(ts timeSource) Option {
	return func(o *options) {
		o.timeSource = ts
	}
}

func WithLogger(l *slog.Logger) Option {
	return func(o *options) {
		o.logger = l
	}
}

// NewScheduler creates a new scheduler for the supplied schedule and associated devices.
func NewScheduler(sched schedule.Annual[devices.Action], place *time.Location, devices map[string]devices.Device, opts ...Option) (*Scheduler, error) {
	scheduler := &Scheduler{
		schedule: sched,
		place:    place,
	}
	for _, opt := range opts {
		opt(&scheduler.options)
	}
	if scheduler.timeSource == nil {
		scheduler.timeSource = systemTimeSource{}
	}
	if scheduler.logger == nil {
		scheduler.logger = slog.New(slog.NewJSONHandler(os.Stderr, nil))
	}
	for i, a := range sched.Actions {
		dev := devices[a.Action.DeviceName]
		if dev == nil {
			return nil, fmt.Errorf("unknown device: %s", a.Action.DeviceName)
		}
		op := dev.Operations()[a.Action.ActionName]
		if op == nil {
			return nil, fmt.Errorf("unknown operation: %s for device: %v", a.Action.ActionName, a.Action.DeviceName)
		}
		sched.Actions[i].Action.Device = dev
		sched.Actions[i].Action.Action = op
	}
	scheduler.logger = scheduler.logger.With("mod", "scheduler", "sched", sched.Name)
	scheduler.scheduler = schedule.NewAnnualScheduler(sched)
	return scheduler, nil
}

type MasterScheduler struct {
	schedulers []*Scheduler
}

func CreateMasterScheduler(schedules []Schedules, devices map[string]devices.Device, opts ...Option) (*MasterScheduler, error) {
	schedulers := make([]*Scheduler, 0, len(schedules))
	for _, sched := range schedules {
		_ = sched
		//schedulers = append(schedulers, NewScheduler(sched, dev, opts...))
	}
	ms := &MasterScheduler{schedulers: schedulers}
	return ms, nil
}
