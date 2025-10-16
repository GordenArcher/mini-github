package mail

import (
	"strconv"

	"gopkg.in/gomail.v2"
)

type Mailer struct {
	host string
	port int
	user string
	pass string
}

func New(host, portStr, user, pass string) *Mailer {
	port, _ := strconv.Atoi(portStr)
	return &Mailer{host: host, port: port, user: user, pass: pass}
}

func (m *Mailer) Send(to, subject, body string) error {
	d := gomail.NewDialer(m.host, m.port, m.user, m.pass)
	msg := gomail.NewMessage()
	msg.SetHeader("From", m.user)
	msg.SetHeader("To", to)
	msg.SetHeader("Subject", subject)
	msg.SetBody("text/html", body)
	return d.DialAndSend(msg)
}
