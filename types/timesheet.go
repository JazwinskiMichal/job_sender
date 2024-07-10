package types

// Timesheet represents a contractor's timesheet.
type Timesheet struct {
	ID           string `json:"id"`
	ContractorID string `json:"contractor_id"`
	GroupID      string `json:"group_id"`

	StorageURL string `json:"storage_url"`
}
