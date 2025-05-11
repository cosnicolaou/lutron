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

func QSLogin(ctx context.Context, s *streamconn.Session, user, pass string) error {
	if _, err := s.ReadUntil(ctx, "login: "); err != nil {
		return fmt.Errorf("user: %v: %w", user, err)
	}
	s.Send(ctx, []byte(user+"\r\n"))
	if _, err := s.ReadUntil(ctx, "password: "); err != nil {
		return fmt.Errorf("user: %v: %w", user, err)
	}
	s.SendSensitive(ctx, []byte(pass+"\r\n"))
	prompt, err := s.ReadUntil(ctx, qsPromptStr, "login:")
	if err != nil {
		return err
	}
	if !bytes.Contains(prompt, qsPrompt) {
		return fmt.Errorf("user: %v: %w", user, ErrQSLogin)
	}
	return nil
}
