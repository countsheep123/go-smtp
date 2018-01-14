package smtp

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"net/mail"
	"net/smtp"
)

type SMTPClient struct {
	host string
	port uint64
	auth smtp.Auth
}

func NewSMTPClient(host, username, password string, port uint64) (*SMTPClient, error) {
	if host == "" {
		return nil, errors.New("host is empty")
	}
	switch port {
	case 25, 465, 587:
	default:
		return nil, errors.New("invalid port")
	}

	var auth smtp.Auth
	if username != "" && password != "" {
		auth = smtp.PlainAuth("", username, password, host)
	}

	return &SMTPClient{
		host: host,
		port: port,
		auth: auth,
	}, nil
}

type Mail struct {
	From    *mail.Address
	To      []*mail.Address
	Cc      []*mail.Address
	Bcc     []*mail.Address
	Subject string
	Body    string
}

func (m *Mail) IsValid() bool {
	if m.From == nil ||
		m.To == nil ||
		m.Cc == nil ||
		m.Bcc == nil {
		return false
	}
	return true
}

func (m *Mail) buildMsg() string {
	msg := ""
	msg += fmt.Sprintf("From: %s\r\n", m.From.String())

	to := ""
	for _, t := range m.To {
		to += fmt.Sprintf("%s;", t.String())
	}
	if len(to) > 0 {
		msg += fmt.Sprintf("To: %s\r\n", to)
	}

	cc := ""
	for _, c := range m.Cc {
		cc += fmt.Sprintf("%s;", c.String())
	}
	if len(to) > 0 {
		msg += fmt.Sprintf("Cc: %s\r\n", cc)
	}

	msg += fmt.Sprintf("Subject: %s\r\n", m.Subject)
	msg += fmt.Sprintf("\r\n")
	msg += fmt.Sprintf("%s", m.Body)

	return msg
}

func (cli *SMTPClient) SendMail(m *Mail) error {
	if m == nil || !m.IsValid() {
		return errors.New("invalid mail")
	}

	c, err := cli.getClient()
	if err != nil {
		return err
	}
	defer c.Close()

	if cli.auth != nil {
		if err := c.Auth(cli.auth); err != nil {
			return err
		}
	}

	if err := c.Mail(m.From.Address); err != nil {
		return err
	}

	recipients := []*mail.Address{}
	recipients = append(recipients, m.To...)
	recipients = append(recipients, m.Cc...)
	recipients = append(recipients, m.Bcc...)

	for _, r := range recipients {
		if err := c.Rcpt(r.Address); err != nil {
			return err
		}
	}

	wc, err := c.Data()
	if err != nil {
		return err
	}

	buf := bytes.NewBufferString(m.buildMsg())
	if _, err = buf.WriteTo(wc); err != nil {
		return err
	}

	if err := wc.Close(); err != nil {
		return err
	}

	if err := c.Quit(); err != nil {
		return err
	}

	return nil
}

func (cli *SMTPClient) getClient() (*smtp.Client, error) {
	switch cli.port {
	case 25: // w/o ssl
		c, err := smtp.Dial(fmt.Sprintf("%s:%d", cli.host, cli.port))
		if err != nil {
			return nil, err
		}
		return c, nil
	case 465, 587: // implicit ssl, explicit ssl
		tlsConfig := &tls.Config{
			InsecureSkipVerify: true,
			ServerName:         cli.host,
		}

		conn, err := tls.Dial("tcp", fmt.Sprintf("%s:%d", cli.host, cli.port), tlsConfig)
		if err != nil {
			return nil, err
		}

		if conn == nil {
			return nil, errors.New("connetion is nil")
		}

		c, err := smtp.NewClient(conn, cli.host)
		if err != nil {
			return nil, err
		}
		return c, nil
	default:
		return nil, errors.New("invalid port")
	}
}
