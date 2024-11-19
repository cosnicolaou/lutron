// Copyright 2024 Cosmos Nicolaou. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package protocol

import (
	"time"
)

// #DEVICE,2,4,3<CR><LF>
// Integration ID, Component Number, Action Number
// See https://assets.lutron.com/a/documents/040249.pdf
func Device(s Session, id, component, action int) error {
	/*
		cmd := fmt.Sprintf("#DEVICE,%d,%d,%d\r\n", id, component, action)
		s.Send(cmd)
		response := s.ReadUntil("QNET> ")
		if err := s.Err(); err != nil {
			return err
		}
		_ = response*/
	return nil
}

// #OUTPUT,1,1,75,01:30<CR><LF>
// Integration ID, Action Number, Level, Fade Time (seconds)
// See https://assets.lutron.com/a/documents/040249.pdf
func Output(s Session, id, component, level int, time time.Duration) {
}

// OUTPUT,3,1<CR><LF>
// Integration ID, Action
// See https://assets.lutron.com/a/documents/040249.pdf
func Query(s Session, id, action int) {
}
