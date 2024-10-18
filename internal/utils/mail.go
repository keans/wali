package utils

import (
	gomail "gopkg.in/mail.v2"
)

type Smtp struct {
	host     string
	port     int
	username string
	password string
	from     string
	to       string
}

func NewSmtp(host string, port int, username string, password string,
	from string, to string) *Smtp {

	return &Smtp{
		host:     host,
		port:     port,
		username: username,
		password: password,
		from:     from,
		to:       to,
	}
}

func (s *Smtp) SendMail(subject string, body string) error {

	// prepare the message
	msg := gomail.NewMessage()
	msg.SetHeader("From", s.from)
	msg.SetHeader("To", s.to)
	msg.SetHeader("Subject", subject)
	msg.SetBody("text/plain", body)

	// connect to the SMTP server and send the msg (TLS is enabled by default)
	d := gomail.NewDialer(s.host, s.port, s.username, s.password)
	if err := d.DialAndSend(msg); err != nil {
		return err
	}

	return nil
}
