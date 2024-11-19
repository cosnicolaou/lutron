// Copyright 2024 Cosmos Nicolaou. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package protocol

import (
	"sync"
	"time"
)

type IdleTimer struct {
	mu       sync.Mutex
	idleTime time.Duration
	elapsed  time.Duration
	last     time.Time
}

func NewIdleTimer(d time.Duration) *IdleTimer {
	return &IdleTimer{idleTime: d, last: time.Now()}
}

func (d *IdleTimer) Reset() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.elapsed = 0
	d.last = time.Now()
}

func (d *IdleTimer) Expired() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	return time.Now().After(d.last.Add(d.idleTime))
}

func (d *IdleTimer) Remaining() time.Duration {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.last.Add(d.idleTime).Sub(time.Now())
}
