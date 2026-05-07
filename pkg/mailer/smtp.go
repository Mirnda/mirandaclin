package mailer

import (
	"bytes"
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

func (m *smtpMailer) Send(_ context.Context, to, subject, htmlBody string) error {
	addr := fmt.Sprintf("%s:%s", m.host, m.port)

	auth := smtp.PlainAuth("", m.user, m.pass, m.host)

	// msg := fmt.Sprintf(
	// 	"From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s",
	// 	m.from, to, subject, htmlBody,
	// )

	msg := bytes.Buffer{}

	msg.WriteString(fmt.Sprintf("From: %s\r\n", m.from))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", to))
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	msg.WriteString("MIME-Version: 1.0\r\n")
	msg.WriteString("Content-Type: text/html; charset=\"UTF-8\"\r\n")
	msg.WriteString("\r\n")
	msg.WriteString(htmlBody)

	return smtp.SendMail(addr, auth, m.from, []string{to}, msg.Bytes())
}
