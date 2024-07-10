package types

// Attachment represents an email attachment.
type Attachment struct {
	Filename string
	Content  []byte
}
