package core

import (
	"fmt"
	"net/smtp"
	"strconv"

	"job_sender/interfaces"
	constants "job_sender/utils/constants"
)

type EmailService struct {
	email       string
	appPassword string
}

// EnsureEmailService implements the IEmailService interface.
var _ interfaces.IEmailService = &EmailService{}

func NewEmailService(email string, appPassword string) *EmailService {
	return &EmailService{
		email:       email,
		appPassword: appPassword,
	}
}

// SendVerificationEmail sends a verification email to the user.
func (h *EmailService) SendVerificationEmail(to string, link string) error {
	subject := "Job sender account verification"
	body := fmt.Sprintf("Click the link to verify your Job sender account: %s", link)
	msg := fmt.Sprintf("From: %s\nTo: %s\nSubject: %s\n\n%s", h.email, to, subject, body)

	// Use smtp.PlainAuth with the app password
	auth := smtp.PlainAuth("", h.email, h.appPassword, constants.SmtpGmailAddress)

	// Gmail SMTP server requires TLS connection on port 587
	err := smtp.SendMail(fmt.Sprintf("%s:%s", constants.SmtpGmailAddress, strconv.Itoa(constants.SmtpGmailPort)), auth, h.email, []string{to}, []byte(msg))
	if err != nil {
		return err
	}

	return nil
}
