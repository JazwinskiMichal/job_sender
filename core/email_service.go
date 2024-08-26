package core

import (
	"fmt"
	"io"
	"mime"
	"net/smtp"
	"strconv"

	"job_sender/interfaces"
	"job_sender/types"
	constants "job_sender/utils/constants"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
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
func (h *EmailService) SendTimesheetRequestEmail(contractor *types.Contractor, requestID string) error {
	subject := fmt.Sprintf("Timesheet %s [%s]", requestID, contractor.ID)
	body := fmt.Sprintf("Hi %s %s. Please submit your timesheet. You can submit it by replying to this email with the timesheet attached.", contractor.Name, contractor.Surname)
	msg := fmt.Sprintf("From: %s\nTo: %s\nSubject: %s\n\n%s", h.email, contractor.Email, subject, body)

	// Use smtp.PlainAuth with the app password
	auth := smtp.PlainAuth("", h.email, h.appPassword, constants.SmtpGmailAddress)

	// Gmail SMTP server requires TLS connection on port 587
	err := smtp.SendMail(fmt.Sprintf("%s:%s", constants.SmtpGmailAddress, strconv.Itoa(constants.SmtpGmailPort)), auth, h.email, []string{contractor.Email}, []byte(msg))
	if err != nil {
		return err
	}

	return nil
}

// GetEmailAttachments returns the attachments of an email.
func (h *EmailService) GetEmailAttachments(subject string) ([]types.Attachment, error) {
	// Create a new IMAP client instance
	c, err := client.DialTLS(constants.ImapGmailAddress+":993", nil)
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
	criteria.Header.Add("Subject", mime.QEncoding.Encode("utf-8", subject))
	criteria.WithoutFlags = []string{"\\Seen"} // Exclude emails that are marked as Seen
	uids, err := c.Search(criteria)
	if err != nil {
		return nil, fmt.Errorf("failed to search emails: %w", err)
	}

	// Nothing to process
	if len(uids) == 0 {
		return []types.Attachment{}, nil
	}

	// Fetch the emails
	seqSet := new(imap.SeqSet)
	seqSet.AddNum(uids...)

	// We're only interested in the email structure and attachments
	section := &imap.BodySectionName{}
	items := []imap.FetchItem{imap.FetchBodyStructure, section.FetchItem()}
	messages := make(chan *imap.Message)

	var attachments []types.Attachment
	done := make(chan error, 1)

	// Start a goroutine to fetch messages
	go func() {
		done <- c.Fetch(seqSet, items, messages)
	}()

	// Process messages as they come in
	for msg := range messages {
		attachs, err := h.extractAttachments(msg, c)
		if err != nil {
			return nil, fmt.Errorf("failed to extract attachments: %w", err)
		}
		attachments = append(attachments, attachs...)
	}

	// Check for any fetch errors
	if err := <-done; err != nil {
		return nil, fmt.Errorf("error during message fetch: %w", err)
	}

	return attachments, nil
}

// ArchiveEmail archives emails matching the given subject by removing them from the INBOX.
func (h *EmailService) ArchiveEmail(subject string) error {
	// Create a new IMAP client instance
	c, err := client.DialTLS(constants.ImapGmailAddress+":993", nil)
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

func (h *EmailService) extractAttachments(msg *imap.Message, c *client.Client) ([]types.Attachment, error) {
	var attachments []types.Attachment

	section := &imap.BodySectionName{}
	fetchItem := section.FetchItem()
	seqSet := new(imap.SeqSet)
	seqSet.AddNum(msg.SeqNum)
	items := []imap.FetchItem{fetchItem}

	// Create a channel to receive the fetched message
	fetchedMsg := make(chan *imap.Message, 1)

	// Fetch the entire message
	err := c.Fetch(seqSet, items, fetchedMsg)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch message: %w", err)
	}

	// Read the message from the channel
	m := <-fetchedMsg
	r := m.GetBody(section)
	if r == nil {
		return nil, fmt.Errorf("no body found for message")
	}

	// Create a mail reader
	mr, err := mail.CreateReader(r)
	if err != nil {
		return nil, fmt.Errorf("failed to create mail reader: %w", err)
	}
	defer mr.Close()

	// Iterate through each part of the email
	for {
		p, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			//log.Printf("Error reading next part: %v", err)
			continue
		}

		switch h := p.Header.(type) {
		case *mail.AttachmentHeader:
			// This is an attachment
			filename, err := h.Filename()
			if err != nil {
				//log.Printf("Error getting filename: %v", err)
				continue
			}

			content, err := io.ReadAll(p.Body)
			if err != nil {
				//log.Printf("Error reading attachment content: %v", err)
				continue
			}

			attachments = append(attachments, types.Attachment{
				Filename: filename,
				Content:  content,
			})

			//log.Printf("Extracted attachment: %s, size: %d bytes", filename, len(content))
		}
	}

	return attachments, nil
}
