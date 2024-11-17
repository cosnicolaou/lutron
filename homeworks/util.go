// Copyright 2024 Cosmos Nicolaou. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package homeworks

import (
	"fmt"
	"log/slog"
	"os"
)

func log(logger *slog.Logger, verbose bool, format string, args ...interface{}) {
	if logger != nil {
		logger.Info(format, args...)
	}
	if verbose {
		fmt.Fprintf(os.Stderr, format, args...)
	}
}
