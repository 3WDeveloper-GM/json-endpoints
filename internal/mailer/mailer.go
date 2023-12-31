package mailer

import (
	"bytes"
	"embed"
	"html/template"

	"github.com/go-mail/mail/v2"
)

//go:embed "templates"
var templateFS embed.FS

type Mailer struct {
	Dialer *mail.Dialer
	Sender string
}

func (m Mailer) Send(recipient, templatefile string, data interface{}) error {
	tmpl, err := template.New("email").ParseFS(templateFS, "templates/"+templatefile)
	if err != nil {
		return err
	}

	subject := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(subject, "subject", data)
	if err != nil {
		return err
	}

	plainbody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(plainbody, "plainBody", data)
	if err != nil {
		return err
	}

	htmlbody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(htmlbody, "htmlBody", data)
	if err != nil {
		return err
	}

	msg := mail.NewMessage()
	msg.SetHeader("To", recipient)
	msg.SetHeader("From", m.Sender)
	msg.SetHeader("Subject", subject.String())
	msg.SetBody("text/plain", plainbody.String())
	msg.AddAlternative("text/html", htmlbody.String())

	err = m.Dialer.DialAndSend(msg)
	if err != nil {
		return err
	}

	return nil
}
