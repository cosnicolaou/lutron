// Copyright 2024 Cosmos Nicolaou. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package protocol

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/cosnicolaou/automation/net/streamconn"
)

var (
	// Failed to login.
	ErrQSLogin = errors.New("QS login failed")

	qsPrompt    = []byte("QNET> ")
	qsPromptStr = string(qsPrompt)
)

func QSLogin(ctx context.Context, s streamconn.Session, user, pass string) error {
	s.ReadUntil(ctx, "login: ")
	s.Send(ctx, []byte(user+"\r\n"))
	s.ReadUntil(ctx, "password: ")
	s.SendSensitive(ctx, []byte(pass+"\r\n"))
	prompt := s.ReadUntil(ctx, qsPromptStr, "login:")
	if err := s.Err(); err != nil {
		return err
	}
	if !bytes.Contains(prompt, qsPrompt) {
		return fmt.Errorf("user: %v: %w", user, ErrQSLogin)
	}
	return nil
}
