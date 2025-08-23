package mail

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/dresswithpockets/openstats/app/env"
	"github.com/dresswithpockets/openstats/app/log"
	"github.com/rotisserie/eris"
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
	SourceArn        string
	AwsConfig        aws.Config
	CredentialsCache *aws.CredentialsCache
}

func (a *AmazonSesMailer) getClient(ctx context.Context) (*ses.Client, error) {
	mailerCredentials, err := a.CredentialsCache.Retrieve(ctx)
	if err != nil {
		return nil, eris.Wrap(err, "failed to retrieve AWS credentials from cache")
	}

	// we copy to avoid loading the default config every time we send mail
	awsConfig := a.AwsConfig.Copy()
	awsConfig.Credentials = credentials.NewStaticCredentialsProvider(
		mailerCredentials.AccessKeyID,
		mailerCredentials.SecretAccessKey,
		mailerCredentials.SessionToken)

	return ses.NewFromConfig(awsConfig), nil
}

func (a *AmazonSesMailer) Send(ctx context.Context, mail Mail) error {
	client, clientErr := a.getClient(ctx)
	if clientErr != nil {
		return clientErr
	}

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
	_, err := client.SendEmail(ctx, input)
	return err
}

// LogMailer is a Mailer which doesn't send any mail over the network - it logs all sent mail to a logger, instead.
type LogMailer struct {
	Logger   *slog.Logger
	LogLevel slog.Level
}

func (l *LogMailer) Send(ctx context.Context, mail Mail) (err error) {
	l.Logger.Log(ctx, l.LogLevel, "Sending mail", "to", mail.To, "subject", mail.Subject, "body", mail.Body)
	return
}

var Default Mailer

func Setup(ctx context.Context) (err error) {
	mode := env.GetString("OPENSTATS_MAILER")

	switch mode {
	case "Log":
		Default, err = setupLogMailer()
	case "AmazonSES":
		Default, err = setupAmazonSesMailer(ctx)
	default:
		err = eris.Errorf("invalid value for OPENSTATS_MAILER: %s", mode)
	}

	return
}

func setupLogMailer() (*LogMailer, error) {
	return &LogMailer{
		Logger:   log.Logger,
		LogLevel: slog.LevelInfo,
	}, nil
}

func setupAmazonSesMailer(ctx context.Context) (*AmazonSesMailer, error) {
	awsConfig, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, eris.Wrap(err, "")
	}

	stsClient := sts.NewFromConfig(awsConfig)
	roleArn := env.GetString("OPENSTATS_MAILER_ROLE_ARN")
	roleProvider := stscreds.NewAssumeRoleProvider(stsClient, roleArn)

	return &AmazonSesMailer{
		SourceArn:        env.GetString("OPENSTATS_MAILER_SOURCE_ARN"),
		AwsConfig:        awsConfig,
		CredentialsCache: aws.NewCredentialsCache(roleProvider),
	}, nil
}
