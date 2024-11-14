// Copyright 2024 Cosmos Nicolaou. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package homeworks

func GetQSProcessorConfig(c any) QSProcessorConfig {
	return c.(*qsProcessor).QSProcessorConfig
}

func GetBlindConfig(d any) HWBlindConfig {
	return d.(*hwBlind).HWBlindConfig
}
