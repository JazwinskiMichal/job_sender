package interfaces

type IEmailService interface {
	// SendVerificationEmail sends a verification email to the user.
	SendVerificationEmail(email string, link string) error
}
