package types

// Timesheet represents a contractor's timesheet.
type Timesheet struct {
	ID           string `firestore:"id"`
	ContractorID string `firestore:"contractor_id"`
	RequestID    string `firestore:"request_id"`

	StorageURL string `firestore:"storage_url"`
}
