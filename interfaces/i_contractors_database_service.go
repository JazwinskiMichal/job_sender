package interfaces

import (
	"job_sender/types"
)

// IContractorsDatabaseService is an interface for a database service that manages contractors.
type IContractorsDatabaseService interface {
	// GetContractors lists all for a group.
	GetContractors(groupID string) ([]*types.Contractor, error)

	// GetContractor gets a contractor by ID.
	GetContractor(id string) (*types.Contractor, error)

	// AddContractor adds a contractor to a group.
	AddContractor(groupID string, contractor *types.Contractor) error

	// UpdateContractor updates a contractor.
	UpdateContractor(contractor *types.Contractor) error

	// DeleteContractor deletes a contractor.
	DeleteContractor(id string) error
}
