package mailer

import (
	"bytes"
	"embed"
	"path"
	"text/template"
)

const (
	FromName            = "GopherSocial"
	maxRetries          = 5
	UserWelcomeTemplate = "user_invitatoin.gotmpl"
)

//go:embed templates/*
var FS embed.FS

type Email struct {
	Subject string
	Body    string
}

type Client interface {
	Send(templateFile, username, email string, data any, isSandbox bool) (int, error)
}

func RenderEmailTemplate(templateFileName string, data any) (Email, error) {
	path := path.Join("./templates", templateFileName)
	tpl, err := template.ParseFS(FS, path)
	if err != nil {
		return Email{}, err
	}

	subject := new(bytes.Buffer)
	if err := tpl.ExecuteTemplate(subject, "subject", nil); err != nil {
		return Email{}, err
	}

	body := new(bytes.Buffer)
	if err := tpl.ExecuteTemplate(body, "body", nil); err != nil {
		return Email{}, err
	}

	email := Email{
		Subject: subject.String(),
		Body:    body.String(),
	}

	return email, nil
}
