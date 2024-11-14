// Copyright 2024 Cosmos Nicolaou. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package schedule

type TimeSource timeSource

func WithTimeSource(ts TimeSource) Option {
	return func(o *options) {
		o.timeSource = ts
	}
}
