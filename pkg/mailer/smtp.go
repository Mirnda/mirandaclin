package mailer

import (
	"context"
	"fmt"
	"net/smtp"
)

type smtpMailer struct {
	host string
	port string
	user string
	pass string
	from string
}

func NewSMTP(host, port, user, pass, from string) Mailer {
	return &smtpMailer{host: host, port: port, user: user, pass: pass, from: from}
}

func (m *smtpMailer) Send(_ context.Context, to, subject, body string) error {
	auth := smtp.PlainAuth("", m.user, m.pass, m.host)
	msg := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s",
		m.from, to, subject, body,
	)
	return smtp.SendMail(m.host+":"+m.port, auth, m.from, []string{to}, []byte(msg))
}
