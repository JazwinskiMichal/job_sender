package core

import (
	"fmt"
	"io"
	"net/smtp"
	"strconv"
	"strings"

	"job_sender/interfaces"
	"job_sender/types"
	constants "job_sender/utils/constants"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
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
	criteria.Header.Add("Subject", subject)
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

	messages := make(chan *imap.Message, len(uids))
	if err := c.Fetch(seqSet, items, messages); err != nil {
		return nil, fmt.Errorf("failed to fetch emails: %w", err)
	}

	var attachments []types.Attachment
	for msg := range messages {
		attachs, err := h.extractAttachments(msg, c)
		if err != nil {
			return nil, fmt.Errorf("failed to extract attachments: %w", err)
		}
		attachments = append(attachments, attachs...)
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

	// Helper function to recursively extract attachments
	var extractParts func(*imap.BodyStructure, []int) error
	extractParts = func(part *imap.BodyStructure, path []int) error {
		if part.Disposition == "attachment" || part.Disposition == "inline" {
			section := &imap.BodySectionName{
				BodyPartName: imap.BodyPartName{
					Path: path,
				},
				Peek: true,
			}

			fetchItem := section.FetchItem()
			seqSet := new(imap.SeqSet)
			seqSet.AddNum(msg.SeqNum)

			items := []imap.FetchItem{fetchItem}

			// Create a channel to receive the fetched message
			fetchedMsg := make(chan *imap.Message, 1)

			// Fetch the message part
			err := c.Fetch(seqSet, items, fetchedMsg)
			if err != nil {
				return fmt.Errorf("failed to fetch attachment: %w", err)
			}

			// Read the message from the channel
			m := <-fetchedMsg

			body := m.GetBody(section)
			if body == nil {
				return fmt.Errorf("no body found for attachment")
			}

			content, err := io.ReadAll(body)
			if err != nil {
				return fmt.Errorf("failed to read attachment content: %w", err)
			}

			// Extract filename from DispositionParams
			filename := ""
			if part.DispositionParams != nil {
				filename = part.DispositionParams["filename"]
			}
			if filename == "" {
				// Fallback to Content-Type name parameter
				filename = part.Params["name"]
			}
			if filename == "" {
				// If still no filename, generate one based on the part path
				filename = fmt.Sprintf("attachment_%s", strings.Join(strings.Fields(fmt.Sprint(path)), "_"))
			}

			attachments = append(attachments, types.Attachment{
				Filename: filename,
				Content:  content,
			})
		}

		for i, subpart := range part.Parts {
			newPath := append(path, i+1)
			if err := extractParts(subpart, newPath); err != nil {
				return err
			}
		}

		return nil
	}

	// Start the recursive extraction from the root
	if err := extractParts(msg.BodyStructure, []int{}); err != nil {
		return nil, err
	}

	return attachments, nil
}
