package smtp

import (
	"net/mail"
	"testing"
)

func TestSendMail(t *testing.T) {
	host := "127.0.0.1"
	username := "john@mail.example.com"
	password := "doe"
	port := uint64(25)

	cli, err := NewSMTPClient(host, username, password, port)
	if err != nil {
		t.Fatal(err)
	}

	mail := &Mail{
		From: &mail.Address{
			Address: "john@mail.example.com",
		},
		To: []*mail.Address{
			{
				Address: "hello@mail.example.com",
			},
		},
		Cc:      []*mail.Address{},
		Bcc:     []*mail.Address{},
		Subject: "hello",
		Body:    "world",
	}

	err = cli.SendMail(mail)
	if err != nil {
		t.Fatal(err)
	}
}
