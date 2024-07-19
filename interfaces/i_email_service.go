package interfaces

import (
	"job_sender/types"
)

type IEmailService interface {
	// SendVerificationEmail sends a verification email to the user.
	SendVerificationEmail(email string, link string) error

	// SendTimsheetRequestEmail sends a timesheet request email to the contractor.
	SendTimesheetRequestEmail(contractor *types.Contractor, weekID string) error

	// SendPasswordResetEmail sends a password reset email to the user.
	// TODO: Implement this method.

	// GetEmailAttachments returns the attachments of an email.
	GetEmailAttachments(subject string) ([]types.Attachment, error)

	// ArchiveEmails archives emails with the specified subject.
	ArchiveEmail(subject string) error
}
