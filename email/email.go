package email

import (
	"fmt"
	"net/smtp"

	"gopkg.cc/apibase/errx"
	"gopkg.cc/apibase/helper"
)

type Sender uint

const (
	None Sender = iota
	SMTP
	Exchange
)

type EmailConfig struct {
	Sender      Sender
	From        string              `toml:"from"` // Specifies default sender, can be empty, usually the same as username
	Username    string              `toml:"username"`
	Password    helper.SecretString `toml:"password"`
	Host        string              `toml:"host"` // Either Host+Port or ExchangeURI must be specified
	Port        int                 `toml:"port"`
	ExchangeURI string              `toml:"exchange_uri"`
}

func (c EmailConfig) SendEmail(to []string, msg []byte) error {
	if c.From == "" {
		return errx.NewWithType(ErrInvalidConfig, "from email address not specified in config")
	}
	return c.SendMailFrom(c.From, to, msg)
}

func (c EmailConfig) SendMailFrom(from string, to []string, msg []byte) error {
	switch c.Sender {
	case SMTP:
		if c.Host == "" || c.Port == 0 {
			return errx.NewWithType(ErrInvalidConfig, "Host and Port must be specified in config")
		}
		// Only connects via TLS o, if TLS is possible or connecting to localhost
		auth := smtp.PlainAuth("", c.Username, c.Password.GetSecret(), c.Host)
		return smtp.SendMail(fmt.Sprintf("%s:%d", c.Host, c.Port), auth, from, to, msg)
	case Exchange:
		// if c.ExchangeURI == "" {
		// 	return errx.NewWithType(ErrInvalidConfig, "Exchange URI must be specified in config")
		// }
		return errx.NewWithType(errx.ErrNotImplemented, "sending email via exchange")
	}
	return errx.NewWithTypef(ErrInvalidConfig, "Unknown sender type %d", c.Sender)
}
