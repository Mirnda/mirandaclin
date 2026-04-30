package mailer

import (
	"context"
	"fmt"
)

type noopMailer struct{}

func NewNoop() Mailer { return &noopMailer{} }

func (n *noopMailer) Send(_ context.Context, to, subject, body string) error {
	fmt.Printf("[mailer:noop] to=%s subject=%s body=%s\n", to, subject, body)
	return nil
}
