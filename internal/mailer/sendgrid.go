package mailer

import (
	"fmt"
	"time"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type SendGridMailer struct {
	fromEmail string
	apiKey    string
	client    *sendgrid.Client
}

func NewSendGrid(apiKey, fromEmail string) *SendGridMailer {
	client := sendgrid.NewSendClient(apiKey)

	return &SendGridMailer{
		apiKey:    apiKey,
		fromEmail: fromEmail,
		client:    client,
	}
}

func (m *SendGridMailer) Send(templateFileName, username, email string, data any, isSandbox bool) (int, error) {
	from := mail.NewEmail(FromName, m.fromEmail)
	to := mail.NewEmail(username, email)

	Email, err := RenderEmailTemplate(templateFileName, data)
	if err != nil {
		return -1, err
	}

	message := mail.NewSingleEmail(from, Email.Subject, to, "", Email.Body)

	message.SetMailSettings(&mail.MailSettings{
		SandboxMode: &mail.Setting{
			Enable: &isSandbox,
		},
	})

	var retryErr error
	for i := 0; i < maxRetries; i++ {
		response, retryErr := m.client.Send(message)
		if retryErr != nil {
			time.Sleep(time.Second * time.Duration(i+1))
			continue
		}

		return response.StatusCode, retryErr
	}

	return -1, fmt.Errorf("failed to send email after %d attempts, error: %v", maxRetries, retryErr)
}
