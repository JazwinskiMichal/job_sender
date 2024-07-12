package core

import (
	"fmt"
	"io"
	"net/smtp"
	"strconv"

	"job_sender/interfaces"
	"job_sender/types"
	constants "job_sender/utils/constants"

	"github.com/emersion/go-imap"
	imapClient "github.com/emersion/go-imap/client"
	"github.com/emersion/go-message/mail"
)

type EmailService struct {
	email       string
	appPassword string
}

// Ensure EmailService implements the IEmailService interface.
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

// SendTimesheetRequestEmail sends a timesheet request email to the contractor.
func (h *EmailService) SendTimesheetRequestEmail(contractorEmail string, contractorID string, weekID string) error {
	subject := fmt.Sprintf("[Contractor ID: %s] Timesheet %s", contractorID, weekID)
	body := "Please submit your timesheet. You can submit it by replying to this email with the timesheet attached."
	msg := fmt.Sprintf("From: %s\nTo: %s\nSubject: %s\n\n%s", h.email, contractorEmail, subject, body)

	// Use smtp.PlainAuth with the app password
	auth := smtp.PlainAuth("", h.email, h.appPassword, constants.SmtpGmailAddress)

	// Gmail SMTP server requires TLS connection on port 587
	err := smtp.SendMail(fmt.Sprintf("%s:%s", constants.SmtpGmailAddress, strconv.Itoa(constants.SmtpGmailPort)), auth, h.email, []string{contractorEmail}, []byte(msg))
	if err != nil {
		return err
	}

	return nil
}

// GetEmailAttachments returns the attachments of an email.
func (h *EmailService) GetEmailAttachments(subject string) ([]types.Attachment, error) {
	// Create a new IMAP client instance
	c, err := imapClient.DialTLS(constants.ImapGmailAddress+":993", nil)
	if err != nil {
		return nil, err
	}
	defer c.Logout()

	// Login to the IMAP server
	err = c.Login(h.email, h.appPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to login: %w", err)
	}

	// Select the INBOX
	_, err = c.Select("INBOX", false)
	if err != nil {
		return nil, fmt.Errorf("failed to select INBOX: %w", err)
	}

	// Search for emails with the specified subject
	criteria := imap.NewSearchCriteria()
	criteria.Header.Add("Subject", subject)
	uids, err := c.Search(criteria)
	if err != nil {
		return nil, fmt.Errorf("failed to search emails: %w", err)
	}

	// Fetch the emails
	seqSet := new(imap.SeqSet)
	seqSet.AddNum(uids...)
	section := &imap.BodySectionName{}
	items := []imap.FetchItem{section.FetchItem()}
	messages := make(chan *imap.Message, len(uids))
	err = c.Fetch(seqSet, items, messages)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch emails: %w", err)
	}

	var attachments []types.Attachment
	for msg := range messages {
		r := msg.GetBody(section)
		if r == nil {
			continue
		}

		mr, err := mail.CreateReader(r)
		if err != nil {
			return nil, fmt.Errorf("failed to create mail reader: %w", err)
		}

		// Ensure the mail reader is closed after processing the email
		func() {
			defer mr.Close()

			// Iterate through each part of the email.
			for {
				p, err := mr.NextPart()
				if err == io.EOF {
					break
				} else if err != nil {
					return // Log or handle the error as needed
				}

				switch h := p.Header.(type) {
				case *mail.AttachmentHeader:
					// This is an attachment
					filename, err := h.Filename()
					if err != nil {
						return // Log or handle the error as needed
					}
					content, err := io.ReadAll(p.Body)
					if err != nil {
						return // Log or handle the error as needed
					}
					attachments = append(attachments, types.Attachment{Filename: filename, Content: content})
				}
			}
		}()
	}

	return attachments, nil
}

// ArchiveEmail archives emails matching the given subject by removing them from the INBOX.
func (h *EmailService) ArchiveEmail(subject string) error {
	// Create a new IMAP client instance
	c, err := imapClient.DialTLS(constants.ImapGmailAddress+":993", nil)
	if err != nil {
		return err
	}
	defer c.Logout()

	// Login to the IMAP server
	err = c.Login(h.email, h.appPassword)
	if err != nil {
		return err
	}

	// Select the INBOX
	_, err = c.Select("INBOX", false)
	if err != nil {
		return err
	}

	// Search for emails with the specified subject
	criteria := imap.NewSearchCriteria()
	criteria.Header.Add("Subject", subject)
	uids, err := c.Search(criteria)
	if err != nil {
		return err
	}

	// Nothing to archive
	if len(uids) == 0 {
		return nil
	}

	// Prepare to move the emails by removing the "INBOX" label
	seqSet := new(imap.SeqSet)
	seqSet.AddNum(uids...)
	flags := []interface{}{imap.DeletedFlag}
	err = c.UidStore(seqSet, imap.RemoveFlags, flags, nil)
	if err != nil {
		return err
	}

	// Expunge to permanently remove emails marked as Deleted
	err = c.Expunge(nil)
	if err != nil {
		return err
	}

	return nil
}
