package mailer

import (
	"context"

	"github.com/Mirnda/mirandaclin/pkg/config"
)

type Mailer interface {
	Send(ctx context.Context, to, subject, body string) error
}

func New(cfg config.SMTP) Mailer {
	var ml = NewNoop()

	if cfg.Host != "" {
		ml = NewSMTP(cfg.Host, cfg.Port, cfg.User, cfg.Pass, cfg.From)
	}

	return ml
}
