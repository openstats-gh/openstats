package mail

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
	"log/slog"
)

type Mail struct {
	From    string
	To      string
	Subject string
	Body    string
}

type Mailer interface {
	Send(context.Context, Mail) error
}

type AmazonSesMailer struct {
	SourceArn string
	GetClient func(ctx context.Context) ses.Client
}

func (a *AmazonSesMailer) Send(ctx context.Context, mail Mail) (err error) {
	input := &ses.SendEmailInput{
		Source:    aws.String(mail.From),
		SourceArn: aws.String(a.SourceArn),
		Destination: &types.Destination{
			ToAddresses: []string{mail.To},
		},
		Message: &types.Message{
			Subject: &types.Content{Data: aws.String(mail.Subject)},
			Body: &types.Body{
				Text: &types.Content{Data: aws.String(mail.Body)},
			},
		},
	}
	_, err = a.Client.SendEmail(ctx, input)
	return
}

//// SmtpMailer is a Mailer which sends mail via SMTP
//type SmtpMailer struct {
//	Auth    smtp.Auth
//	Address string
//}
//
//func (s *SmtpMailer) Send(mail Mail) (err error) {
//	to := []string{mail.To}
//	msg := []byte("To: " + mail.To + "\r\n" + mail.Subject + "\r\n\r\n" + mail.Body)
//	return smtp.SendMail(s.Address, s.Auth, mail.To, to, msg)
//}

// LogMailer is a Mailer which doesn't send any mail over the network - it logs all sent mail to a logger, instead.
type LogMailer struct {
	Logger   *slog.Logger
	LogLevel slog.Level
}

func (l *LogMailer) Send(ctx context.Context, mail Mail) (err error) {
	l.Logger.Log(ctx, l.LogLevel, "Sending mail", "to", mail.To, "subject", mail.Subject, "body", mail.Body)
	return
}
