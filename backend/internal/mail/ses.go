package mail

import (
	"context"
	"fmt"
	"log"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
	"github.com/ys-1052/yossid/backend/internal/config"
)

type Mailer interface {
	SendVerificationEmail(ctx context.Context, toEmail, token string) error
	SendOTPEmail(ctx context.Context, toEmail, otp string) error
}

type mailer struct {
	cfg       *config.Config
	sesClient *sesv2.Client
}

func NewMailer(ctx context.Context, cfg *config.Config) (Mailer, error) {
	if cfg.Env == "local" {
		// Return mock mailer for local development
		return &mockMailer{cfg: cfg}, nil
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config for SES: %w", err)
	}

	client := sesv2.NewFromConfig(awsCfg)
	return &mailer{
		cfg:       cfg,
		sesClient: client,
	}, nil
}

func (m *mailer) SendVerificationEmail(ctx context.Context, toEmail, token string) error {
	verifyURL := fmt.Sprintf("%s/email/verify?token=%s", m.cfg.Issuer, token)
	subject := "[yossid] Verify your email address"
	body := fmt.Sprintf("Thank you for registering with yossid. Please click the link below to complete your registration:\n\n%s\n\nThis link is valid for 30 minutes.", verifyURL)

	return m.send(ctx, toEmail, subject, body)
}

func (m *mailer) SendOTPEmail(ctx context.Context, toEmail, otp string) error {
	subject := "[yossid] Your two-factor verification code"
	body := fmt.Sprintf("Your verification code is:\n\n%s\n\nThis code is valid for 5 minutes.", otp)

	return m.send(ctx, toEmail, subject, body)
}

func (m *mailer) send(ctx context.Context, toEmail, subject, body string) error {
	input := &sesv2.SendEmailInput{
		FromEmailAddress: &m.cfg.SesFromEmail,
		Destination: &types.Destination{
			ToAddresses: []string{toEmail},
		},
		Content: &types.EmailContent{
			Simple: &types.Message{
				Subject: &types.Content{
					Data: &subject,
				},
				Body: &types.Body{
					Text: &types.Content{
						Data: &body,
					},
				},
			},
		},
	}

	_, err := m.sesClient.SendEmail(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to send SES email to %s: %w", toEmail, err)
	}
	return nil
}

// mockMailer is used for local development, logging emails to standard output instead of sending them.
type mockMailer struct {
	cfg *config.Config
}

func (m *mockMailer) SendVerificationEmail(ctx context.Context, toEmail, token string) error {
	verifyURL := fmt.Sprintf("%s/email/verify?token=%s", m.cfg.Issuer, token)
	log.Printf("[MAIL MOCK] Verification email to %s:\nLink: %s\nToken: %s", toEmail, verifyURL, token)
	return nil
}

func (m *mockMailer) SendOTPEmail(ctx context.Context, toEmail, otp string) error {
	log.Printf("[MAIL MOCK] OTP email to %s:\nOTP: %s", toEmail, otp)
	return nil
}
